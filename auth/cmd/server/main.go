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
	"auth/internal/events/producers"
	"auth/internal/middlewares"
	"auth/internal/observability"
	"auth/internal/repositories"
	"auth/internal/services"
	"auth/internal/storage"
	"context"
	"google.golang.org/grpc"
	"net"
)

func main() {
	// Load local .env if present (fallback for local development)
	config.LoadEnvFile()

	// 1. Setup Logger & Observability
	observability.InitLogger("auth", os.Getenv("APP_ENV"))
	metrics := observability.NewRegistry("auth")
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
	participantDocRepo := repositories.NewParticipantDocumentRepository(db)
	permissionRepo := repositories.NewPermissionRepository(db)
	volunteerRepo := repositories.NewVolunteerRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	participantDirectoryRepo := repositories.NewParticipantDirectoryRepository(db)
	storageCleanupRepo := repositories.NewStorageCleanupRepository(db)
	emailVerificationRepo := repositories.NewEmailVerificationRepository(db)
	passwordRecoveryRepo := repositories.NewPasswordRecoveryRepository(db)
	loginAttemptRepo := repositories.NewLoginAttemptRepository(db)
	auditRepo := repositories.NewAuditRepository(db)

	// Setup Storage
	storageSvc, err := storage.NewObjectStorageService(
		context.Background(),
		os.Getenv("GCP_PROJECT_ID"),
		os.Getenv("GCP_BUCKET_NAME"),
		os.Getenv("GCP_EMULATOR_HOST"),
		os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
	)
	if err != nil {
		slog.Warn("GCP Storage initialization failed, skipping file uploads", "error", err)
	}

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

	workerCtx, cancelWorkers := context.WithCancel(context.Background())

	outboxRepo := repositories.NewOutboxRepository(db)
	usuariosProducer, outboxWorker := initOutbox(workerCtx, outboxRepo)

	notificationSyncSvc := services.NewNotificationSyncService()
	verificationService := services.NewVerificationService(emailVerificationRepo, userRepo, notificationSyncSvc, usuariosProducer)
	passwordRecoveryService := services.NewPasswordRecoveryService(passwordRecoveryRepo, userRepo, refreshTokenRepo, notificationSyncSvc, usuariosProducer)
	twofaService := services.NewTwofaService(userRepo, querier, transactor, usuariosProducer)

	authService := services.NewAuthService(
		transactor,
		querier,
		userRepo,
		refreshTokenRepo,
		jwtService,
		firebaseService,
		verificationService,
		usuariosProducer,
		twofaService,
		loginAttemptRepo,
	)

	userService := services.NewUserService(userRepo, querier, notificationSyncSvc, verificationService, usuariosProducer)
	roleService := services.NewRoleService(roleRepo)
	companyService := services.NewCompanyService(companyRepo)
	teacherService := services.NewTeacherService(teacherRepo)
	participantService := services.NewParticipantService(participantRepo, participantDocRepo, storageSvc, userRepo, storageCleanupRepo, usuariosProducer)
	permissionService := services.NewPermissionService(permissionRepo)
	volunteerService := services.NewVolunteerService(volunteerRepo, userRepo)

	// Iniciar worker de reconciliación asíncrona de archivos huérfanos
	cleanupWorker := services.NewStorageCleanupWorker(storageCleanupRepo, storageSvc)
	cleanupWorker.Start(workerCtx, 5*time.Minute)

	// Sincronizar contadores de impacto en Cloudflare Worker KV al iniciar el servicio
	services.NewKVSyncService(userRepo).SyncImpactStatsAsync()

	// 5. Setup Controllers
	firestoreClient, err := storage.NewFirestoreClient(context.Background(), "")
	if err != nil {
		slog.Warn("Firestore initialization failed, push notification tokens will not be persisted", "error", err)
	}
	deviceRepo := repositories.NewFirestoreDeviceRepository(firestoreClient)
	deviceController := controllers.NewDeviceController(deviceRepo)

	authController := controllers.NewAuthController(authService, verificationService, passwordRecoveryService, twofaService)
	userController := controllers.NewUserController(userService)
	roleController := controllers.NewRoleController(roleService)
	companyController := controllers.NewCompanyController(companyService)
	teacherController := controllers.NewTeacherController(teacherService)
	participantController := controllers.NewParticipantController(participantService)
	permissionController := controllers.NewPermissionController(permissionService)
	volunteerController := controllers.NewVolunteerController(volunteerService)
	tokenValidatorController := controllers.NewTokenValidatorController(jwtService)
	auditController := controllers.NewAuditController(auditRepo)
	participantDirectoryService := services.NewParticipantDirectoryService(participantDirectoryRepo)
	directoryController := controllers.NewDirectoryController(participantDirectoryService)
	grpcAuthController := controllers.NewGrpcAuthController(participantService, companyService, userService, participantDirectoryService, deviceRepo)
	grpcAuthController.SetLinkIssuers(verificationService, passwordRecoveryService, twofaService)
	go startGRPCServer(grpcAuthController)

	// 6. Setup Router and Middlewares
	r := chi.NewRouter()
	r.Use(observability.HTTPMiddleware(metrics, func(req *http.Request) string {
		if rc := chi.RouteContext(req.Context()); rc != nil {
			return rc.RoutePattern()
		}
		return ""
	}))
	r.Use(outboxInlineFlushMiddleware(outboxWorker))

	// Middlewares (CORS and Security Headers must be handled here as API Gateway does not support allowCors)
	middlewares.SetupMiddlewares(r)

	// Override chi's default MethodNotAllowed to handle OPTIONS preflight.
	// chi dispatches to MethodNotAllowedHandler when a path exists for another method
	// (e.g., POST /auth/check-credentials) but not for the requested method (OPTIONS).
	// This runs INSIDE chi's routing tree, so it's guaranteed to execute even if
	// external wrappers like OptionsInterceptor are somehow bypassed.
	r.MethodNotAllowed(http.HandlerFunc(handleMethodNotAllowed))

	// DDoS Protection fallback (Global Rate Limiter)
	r.Use(middlewares.RateLimiter())

	// 7. Define Routes

	r.Get("/health", authController.HealthCheck)
	if observability.MetricsEnabled() {
		r.Handle("/metrics", metrics.Handler())
	}
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
			r.With(middlewares.StrictRateLimiter()).Post("/verify-email", authController.VerifyEmail)
			r.With(middlewares.StrictRateLimiter()).Patch("/verify-email", authController.VerifyEmail)
			r.With(middlewares.StrictRateLimiter()).Post("/resend-verification", authController.ResendVerification)
			r.With(middlewares.StrictRateLimiter()).Post("/forgot-password", authController.ForgotPassword)
			r.With(middlewares.StrictRateLimiter()).Post("/reset-password", authController.ResetPassword)
			r.With(middlewares.StrictRateLimiter()).Post("/verify-2fa", authController.VerifyTwofa)
			r.With(middlewares.StrictRateLimiter()).Post("/resend-2fa", authController.ResendTwofa)
		})

		// Protected Routes
		r.Group(func(r chi.Router) {
			// Auditoría (HU-10, HU-03, HU-06): va por fuera de AuthMiddleware
			// para registrar también los tokens rechazados.
			r.Use(middlewares.AuditTrail(auditRepo))
			r.Use(middlewares.AuthMiddleware(jwtService))

			r.With(middlewares.RequireRole("role_admin")).Get("/auth/auditoria", auditController.List)

			r.Get("/auth/me", authController.GetMe)
			r.Get("/profile/me", authController.GetMyProfileDetails)
			r.Put("/profile/me", authController.UpdateMyProfile)
			r.Put("/profile/me/password", authController.UpdateMyPassword)
			r.Post("/profile/me/link-google", authController.LinkGoogleAccount)
			r.Post("/profile/me/unlink-google", authController.UnlinkGoogleAccount)
			r.Post("/profile/me/2fa/setup", authController.SetupTwofa)
			r.Post("/profile/me/2fa/enable", authController.EnableTwofa)
			r.Post("/profile/me/2fa/disable", authController.DisableTwofa)
			r.Post("/profile/me/2fa/revoke-device", authController.RevokeTrustedDevice)
			// Users CRUD
			r.Route("/users", func(r chi.Router) {
				r.Post("/", userController.CreateUser)
				r.Get("/", userController.GetAllUsers)
				// Dashboard: participantes con altas y bajas. Ruta estática:
				// chi la resuelve antes que /{id}.
				r.With(middlewares.RequireRole("role_admin", "role_dev")).
					Get("/participants/history", userController.GetParticipantsHistory)
				r.Post("/me/devices", deviceController.RegisterDevice)
				r.With(middlewares.RequireRole("role_admin", "role_dev", "role_volunteer")).
					Post("/{id}/approve", userController.ApproveUser)
				r.With(middlewares.RequireRole("role_admin", "role_dev", "role_volunteer")).
					Patch("/{id}/approval", userController.ApproveUser)
				r.Get("/{id}", userController.GetUser)
				// Perfil completo (contacto, DNI, datos del rol): solo el propio
				// usuario o la gestión del centro. Antes cualquier usuario
				// logueado podía leer el de cualquiera (D4, 22/11/2026). El
				// nombre del docente de un curso lo resuelve Educación.
				r.With(middlewares.RequireRoleOrSelf("id", "role_admin", "role_dev", "role_volunteer")).
					Get("/{id}/profile", authController.GetUserProfile)
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

				r.Route("/{id}/documents", func(r chi.Router) {
					r.Post("/", participantController.UploadDocument)
					r.Get("/", participantController.GetDocuments)
					r.Get("/{docId}/download", participantController.DownloadDocument)
					r.Put("/{docId}", participantController.UpdateDocument)
					r.Delete("/{docId}", participantController.DeleteDocument)
				})
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

			// Participant Directory (read-only for inter-service communication)
			r.Route("/directory", func(r chi.Router) {
				// El docente solo puede pedir un participante por id (nombre,
				// sin DNI ni contacto): Educación lo usa para mostrarle los
				// nombres de sus alumnos en las entregas.
				r.With(middlewares.RequireRole("role_admin", "role_dev", "role_volunteer", "role_teacher")).Get("/participants/{id}", directoryController.GetParticipantProfile)
				// by-user: además del staff, el propio usuario puede pedir su
				// perfil (Promoción Humana lo usa con el token del
				// participante para "Mi proceso").
				r.With(middlewares.RequireRoleOrSelf("userId", "role_admin", "role_dev", "role_volunteer")).Get("/participants/by-user/{userId}", directoryController.GetProfileByUserID)
				r.Group(func(r chi.Router) {
					r.Use(middlewares.RequireRole("role_admin", "role_dev", "role_volunteer"))
					r.Get("/participants/search", directoryController.SearchParticipants)
					r.Post("/users/summaries", directoryController.GetUserSummaries)
				})
				r.Route("/internal", func(r chi.Router) {
					r.Use(middlewares.RequireRole("role_admin", "role_dev"))
					grpcAuthController.RegisterInternalRoutes(r)
				})
			})
		})
	})

	// 8. Start Server
	port := os.Getenv("PORT")
	slog.Info("Servicio Auth escuchando en el puerto", "port", port)
	// Wrap chi with a net/http level OPTIONS interceptor that runs before chi's
	// routing tree, bypassing the sub-router method dispatch issue in chi v5.
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           middlewares.OptionsInterceptor(r),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if err := observability.ListenAndServeGraceful(context.Background(), server, func(shutdownCtx context.Context) {
		drainAuthWorkers(shutdownCtx, cancelWorkers, outboxWorker, cleanupWorker)
	}); err != nil {
		slog.Error("Error al iniciar el servidor", "error", err)
		os.Exit(1)
	}
}

func initOutbox(ctx context.Context, outboxRepo repositories.OutboxRepository) (producers.UsuariosProducer, services.OutboxWorker) {
	outboxWorker := startOutboxWorker(ctx, outboxRepo)
	var onEnqueue func()
	if outboxWorker != nil {
		onEnqueue = outboxWorker.MarkDirty
	}
	return producers.NewUsuariosProducer(outboxRepo, onEnqueue), outboxWorker
}

func startOutboxWorker(ctx context.Context, outboxRepo repositories.OutboxRepository) services.OutboxWorker {
	pubsubProject := os.Getenv("PUBSUB_PROJECT_ID")
	if pubsubProject == "" {
		return nil
	}
	pubPublisher, err := services.NewPubSubPublisher(ctx, pubsubProject)
	if err != nil {
		slog.Warn("PubSub publisher initialization failed for outbox worker", "error", err)
		return nil
	}
	outboxWorker := services.NewOutboxWorker(outboxRepo, pubPublisher)
	outboxWorker.Start(ctx, 5*time.Second)
	return outboxWorker
}

func outboxInlineFlushMiddleware(worker services.OutboxWorker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req)
			if worker != nil && services.OutboxInlineFlushEnabled() {
				worker.FlushIfDirty(context.Background(), 2*time.Second)
			}
		})
	}
}

func drainAuthWorkers(ctx context.Context, cancel context.CancelFunc, outboxWorker services.OutboxWorker, cleanupWorker services.StorageCleanupWorker) {
	cancel()
	if outboxWorker != nil {
		outboxWorker.FlushBestEffort(ctx, 5*time.Second)
	}
	if cleanupWorker != nil {
		cleanCtx, cleanCancel := context.WithTimeout(ctx, 3*time.Second)
		defer cleanCancel()
		if _, err := cleanupWorker.ProcessDueTasks(cleanCtx); err != nil {
			slog.Warn("Storage cleanup shutdown drain failed", "error", err)
		}
	}
}

func startGRPCServer(grpcAuthController authpb.AuthServiceServer) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen on gRPC port", "error", err)
		return
	}
	grpcServer := grpc.NewServer()
	authpb.RegisterAuthServiceServer(grpcServer, grpcAuthController)
	slog.Info("Servicio Auth (gRPC) escuchando en el puerto 50051")
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("failed to serve gRPC", "error", err)
	}
}

func handleMethodNotAllowed(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		origin := req.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With, X-Skip-Global-Error, X-Redirect-On-404, X-Skip-500-Redirect")
			w.Header().Set("Access-Control-Expose-Headers", "Link, Set-Cookie")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}
