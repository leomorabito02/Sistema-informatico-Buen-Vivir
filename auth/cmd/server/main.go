package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"auth/config"
	"auth/internal/controllers"
	"auth/internal/middlewares"
	"auth/internal/repositories"
	"auth/internal/services"
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
	jwtService := services.NewJWTService()
	firebaseService, err := services.NewFirebaseService()
	if err != nil {
		slog.Error("Firebase initialization failed", "error", err)
		os.Exit(1)
	}

	authService := services.NewAuthService(
		db,
		userRepo,
		refreshTokenRepo,
		jwtService,
		firebaseService,
	)
	userService := services.NewUserService(userRepo)
	roleService := services.NewRoleService(roleRepo)
	companyService := services.NewCompanyService(companyRepo)
	teacherService := services.NewTeacherService(teacherRepo)
	participantService := services.NewParticipantService(participantRepo)
	permissionService := services.NewPermissionService(permissionRepo)
	volunteerService := services.NewVolunteerService(volunteerRepo)

	// 5. Setup Controllers
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	roleController := controllers.NewRoleController(roleService)
	companyController := controllers.NewCompanyController(companyService)
	teacherController := controllers.NewTeacherController(teacherService)
	participantController := controllers.NewParticipantController(participantService)
	permissionController := controllers.NewPermissionController(permissionService)
	volunteerController := controllers.NewVolunteerController(volunteerService)

	// 6. Setup Router and Middlewares
	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r)

	// DDoS Protection fallback (Global Rate Limiter)
	r.Use(middlewares.RateLimiter())

	// 7. Define Routes
	r.Get("/health", authController.HealthCheck)
	
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.With(middlewares.StrictRateLimiter()).Post("/register", authController.Register)
			r.With(middlewares.StrictRateLimiter()).Post("/login", authController.Login)
			r.With(middlewares.StrictRateLimiter()).Post("/google", authController.LoginWithGoogle)
		})
		
		// Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(jwtService))
			
			r.Get("/auth/me", func(w http.ResponseWriter, req *http.Request) {
				userID, ok := req.Context().Value(middlewares.UserIDKey).(string)
				if !ok {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				w.Header().Set("Content-Type", "application/json")

				_ = json.NewEncoder(w).Encode(map[string]any{
					"user_id": userID,
				})
			})

			// Users CRUD
			r.Route("/users", func(r chi.Router) {
				r.Post("/", userController.CreateUser)
				r.Get("/", userController.GetAllUsers)
				r.Get("/{id}", userController.GetUser)
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
			r.With(middlewares.RequireRole("role_company")).Route("/companies", func(r chi.Router) {
				r.Post("/", companyController.CreateCompany)
				r.Get("/", companyController.GetAllCompanies)
				r.Get("/{id}", companyController.GetCompany)
				r.Put("/{id}", companyController.UpdateCompany)
				r.Delete("/{id}", companyController.DeleteCompany)
			})

			// Teachers CRUD
			r.With(middlewares.RequireRole("role_teacher")).Route("/teachers", func(r chi.Router) {
				r.Post("/", teacherController.CreateTeacher)
				r.Get("/", teacherController.GetAllTeachers)
				r.Get("/{id}", teacherController.GetTeacher)
				r.Put("/{id}", teacherController.UpdateTeacher)
				r.Delete("/{id}", teacherController.DeleteTeacher)
			})

			// Participants CRUD
			r.With(middlewares.RequireRole("role_participant")).Route("/participants", func(r chi.Router) {
				r.Post("/", participantController.CreateParticipant)
				r.Get("/", participantController.GetAllParticipants)
				r.Get("/{id}", participantController.GetParticipant)
				r.Put("/{id}", participantController.UpdateParticipant)
				r.Delete("/{id}", participantController.DeleteParticipant)
			})

			// Permissions CRUD
			r.With(middlewares.RequireRole("role_admin", "role_dev")).Route("/permissions", func(r chi.Router) {
				r.Post("/", permissionController.CreatePermission)
				r.Get("/", permissionController.GetAllPermissions)
				r.Get("/{id}", permissionController.GetPermission)
				r.Put("/{id}", permissionController.UpdatePermission)
				r.Delete("/{id}", permissionController.DeletePermission)
			})

			// Volunteers CRUD
			r.With(middlewares.RequireRole("role_volunteer")).Route("/volunteers", func(r chi.Router) {
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
	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("Error al iniciar el servidor", "error", err)
		os.Exit(1)
	}
}
