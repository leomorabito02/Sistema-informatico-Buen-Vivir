package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"

	"auth/api/proto/authpb"
	"auth/config"
	"auth/internal/controllers"
	"auth/internal/middlewares"
	"auth/internal/repositories"
	"auth/internal/services"
	"google.golang.org/grpc"
	"net"

)

func main() {
	// 1. Setup Logger
	config.InitLogger()
	slog.Info("Starting Auth Service...")

	// 2. Setup Database
	db, err := config.InitDatabase()
	if err != nil {
		slog.Error("Database initialization failed", "error", err)
		os.Exit(1)
	}

	// 3. Setup Repositories
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	companyRepo := repositories.NewCompanyRepository(db)
	teacherRepo := repositories.NewTeacherRepository(db)
	participantRepo := repositories.NewParticipantRepository(db)
	permissionRepo := repositories.NewPermissionRepository(db)
	volunteerRepo := repositories.NewVolunteerRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)

	// 4. Setup Services
	jwtService, err := services.NewJWTServiceFromEnv()
	if err != nil {
		slog.Error("JWT initialization failed", "error", err)
		os.Exit(1)
	}
	firebaseService, err := services.NewFirebaseService()
	if err != nil {
		slog.Error("Firebase initialization failed", "error", err)
		os.Exit(1)
	}

	transactor := repositories.NewGormTransactor(db)
	querier := repositories.NewGormRoleProfileQuerier(db)

	authService := services.NewAuthService(
		transactor,
		querier,
		userRepo,
		refreshTokenRepo,
		jwtService,
		firebaseService,
	)
	userService := services.NewUserService(userRepo)
	roleService := services.NewRoleService(roleRepo)
	companyService := services.NewCompanyService(companyRepo)
	teacherService := services.NewTeacherService(teacherRepo)
	participantService := services.NewParticipantService(participantRepo, userRepo)
	permissionService := services.NewPermissionService(permissionRepo)
	volunteerService := services.NewVolunteerService(volunteerRepo, userRepo)

	// Sincronizar contadores de impacto en Cloudflare Worker KV al iniciar el servicio
	services.NewKVSyncService(userRepo).SyncImpactStatsAsync()

	// 5. Setup Controllers
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	roleController := controllers.NewRoleController(roleService)
	companyController := controllers.NewCompanyController(companyService)
	teacherController := controllers.NewTeacherController(teacherService)
	participantController := controllers.NewParticipantController(participantService)
	permissionController := controllers.NewPermissionController(permissionService)
	volunteerController := controllers.NewVolunteerController(volunteerService)
	tokenValidatorController := controllers.NewTokenValidatorController(jwtService)
	grpcParticipantController := controllers.NewGrpcParticipantController(participantService)

	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			slog.Error("failed to listen on gRPC port", "error", err)
			return
		}
		grpcServer := grpc.NewServer()
		authpb.RegisterAuthServiceServer(grpcServer, grpcParticipantController)
		slog.Info("Servicio Auth (gRPC) escuchando en el puerto 50051")
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("failed to serve gRPC", "error", err)
		}
	}()

	// 6. Setup Router and Middlewares
	r := chi.NewRouter()
	
	// Middlewares (CORS and Security Headers must be handled here as API Gateway does not support allowCors)
	middlewares.SetupMiddlewares(r)

	// Override chi's default MethodNotAllowed to handle OPTIONS preflight.
	// chi dispatches to MethodNotAllowedHandler when a path exists for another method
	// (e.g., POST /auth/check-credentials) but not for the requested method (OPTIONS).
	// This runs INSIDE chi's routing tree, so it's guaranteed to execute even if
	// external wrappers like OptionsInterceptor are somehow bypassed.
	r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodOptions {
			origin := req.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With")
				w.Header().Set("Access-Control-Expose-Headers", "Link, Set-Cookie")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "300")
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}))

	// DDoS Protection fallback (Global Rate Limiter)
	r.Use(middlewares.RateLimiter())

	// 7. Define Routes

	r.Get("/health", authController.HealthCheck)
	r.Get("/internal/validate-token", tokenValidatorController.ValidateToken)

	
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.With(middlewares.StrictRateLimiter()).Post("/register", authController.Register)
			// Step 1: validate credentials, returns list of available accounts/roles
			r.With(middlewares.StrictRateLimiter()).Post("/check-credentials", authController.CheckCredentials)
			r.With(middlewares.StrictRateLimiter()).Post("/google-check-credentials", authController.CheckGoogleCredentials)
			// Step 2: select account/role, receive JWT
			r.With(middlewares.StrictRateLimiter()).Post("/login", authController.Login)
			r.With(middlewares.StrictRateLimiter()).Post("/logout", authController.Logout)
			r.With(middlewares.StrictRateLimiter()).Post("/refresh", authController.Refresh)
		})
		
		// Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(jwtService))
			
			r.Get("/auth/me", authController.GetMe)
			r.Get("/profile/me", authController.GetMyProfileDetails)
			r.Put("/profile/me", authController.UpdateMyProfile)
			r.Put("/profile/me/password", authController.UpdateMyPassword)
			r.Post("/profile/me/link-google", authController.LinkGoogleAccount)
			r.Post("/profile/me/unlink-google", authController.UnlinkGoogleAccount)
			// Users CRUD
			r.Route("/users", func(r chi.Router) {
				r.Post("/", userController.CreateUser)
				r.Get("/", userController.GetAllUsers)
				r.Get("/{id}", userController.GetUser)
				r.Get("/{id}/profile", authController.GetUserProfile)
				r.Put("/{id}", userController.UpdateUser)
				r.Delete("/{id}", userController.DeleteUser)
			})

			// Roles CRUD
			r.With(middlewares.RequireRole("role_admin", "role_dev")).Route("/roles", func(r chi.Router) {
				r.Post("/", roleController.CreateRole)
				r.Get("/", roleController.GetAllRoles)
				r.Get("/{id}", roleController.GetRole)
				r.Put("/{id}", roleController.UpdateRole)
				r.Delete("/{id}", roleController.DeleteRole)
			})

			// Companies CRUD
			r.With(middlewares.RequireRole("role_admin", "role_company")).Route("/companies", func(r chi.Router) {
				r.Post("/", companyController.CreateCompany)
				r.Get("/", companyController.GetAllCompanies)
				r.Get("/{id}", companyController.GetCompany)
				r.Put("/{id}", companyController.UpdateCompany)
				r.Delete("/{id}", companyController.DeleteCompany)
			})

			// Teachers CRUD
			r.With(middlewares.RequireRole("role_admin", "role_teacher")).Route("/teachers", func(r chi.Router) {
				r.Post("/", teacherController.CreateTeacher)
				r.Get("/", teacherController.GetAllTeachers)
				r.Get("/{id}", teacherController.GetTeacher)
				r.Put("/{id}", teacherController.UpdateTeacher)
				r.Delete("/{id}", teacherController.DeleteTeacher)
			})

			// Participants CRUD. role_volunteer también entra (HU-07/08/09 y
			// HU-06: "vista voluntario" puede crear/editar/eliminar perfiles
			// de participantes) — lo necesita education-service para poder
			// validar/listar participantes (inscripción, asistencia).
			r.With(middlewares.RequireRole("role_admin", "role_participant", "role_volunteer")).Route("/participants", func(r chi.Router) {
				r.Post("/", participantController.CreateParticipant)
				r.Get("/", participantController.GetAllParticipants)
				r.Get("/{id}", participantController.GetParticipant)
				r.Put("/{id}", participantController.UpdateParticipant)
				r.Delete("/{id}", participantController.DeleteParticipant)
			})

			// Permissions CRUD
			r.With(middlewares.RequireRole("role_admin", "role_dev")).Route("/permissions", func(r chi.Router) {
				r.Post("/", permissionController.CreatePermission)
				r.With(middlewares.RequirePermissionMiddleware(permissionService, "read", "permissions")).Get("/", permissionController.GetAllPermissions)
				r.Get("/{id}", permissionController.GetPermission)
				r.Put("/{id}", permissionController.UpdatePermission)
				r.Delete("/{id}", permissionController.DeletePermission)
			})

			// Volunteers CRUD
			r.With(middlewares.RequireRole("role_admin", "role_volunteer")).Route("/volunteers", func(r chi.Router) {
				r.Post("/", volunteerController.CreateVolunteer)
				r.Get("/", volunteerController.GetAllVolunteers)
				r.Get("/{id}", volunteerController.GetVolunteer)
				r.Put("/{id}", volunteerController.UpdateVolunteer)
				r.Delete("/{id}", volunteerController.DeleteVolunteer)
			})
		})
	})

	// 8. Start Server
	port := os.Getenv("PORT")
	slog.Info("Servicio Auth escuchando en el puerto", "port", port)
	// Wrap chi with a net/http level OPTIONS interceptor that runs before chi's
	// routing tree, bypassing the sub-router method dispatch issue in chi v5.
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      middlewares.OptionsInterceptor(r),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Error al iniciar el servidor", "error", err)
		os.Exit(1)
	}
}
