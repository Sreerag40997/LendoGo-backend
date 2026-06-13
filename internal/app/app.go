	package app

	import (
		"context"
		"os"

		"github.com/gofiber/fiber/v2"

		"lendogo-backend/database"
		"lendogo-backend/internal/consumers"
		
		// 👇 CHANGED: Back to your unified admin_controller folder
		"lendogo-backend/internal/controllers/admin_controller" 
		"lendogo-backend/internal/controllers/auth_controller"
		"lendogo-backend/internal/controllers/chat_controller"
		consultation_controller "lendogo-backend/internal/controllers/consultation_controller"
		"lendogo-backend/internal/controllers/loan_controller"
		"lendogo-backend/internal/controllers/payment_controller"
		"lendogo-backend/internal/controllers/user_profile_controller"
		"lendogo-backend/internal/controllers/wallet_controller"
		"lendogo-backend/internal/jobs"
		"lendogo-backend/internal/repositories"
		"lendogo-backend/internal/routes"
		"lendogo-backend/internal/services"
		"lendogo-backend/utils"
		"lendogo-backend/internal/websockets"
	)

	// SetupApp is now incredibly clean. You can read it like a book!
	func SetupApp(app *fiber.App) {
		// Start the Admin Broadcast Hub
		go websockets.StartHub()

		// 1. Setup Infrastructure
		kafkaProducer := setupKafka()

		// 2. Wire Dependencies
		repos := setupRepositories()
		businessServices := setupServices(repos, kafkaProducer)

		// 3. Boot Background Workers
		startConsumers(businessServices)

		// Boot the Time Machine (Cronjobs)
		startCronJobs(businessServices)

		// 4. Mount Routes 
		setupRoutes(app, businessServices, repos)
	}

	// ==========================================
	// 🛠️ HELPER FUNCTIONS
	// ==========================================

	func setupKafka() *utils.KafkaProducer {
		broker := os.Getenv("KAFKA_BROKER")
		if broker == "" {
			broker = "localhost:9092"
		}
		return utils.NewKafkaProducer(broker)
	}

	// Struct to hold all our Repositories
	type Repositories struct {
		User         repositories.UserRepository
		Consultation repositories.ConsultationRepository
		Loan         repositories.LoanRepository
		Wallet       repositories.WalletRepository
		Chat         repositories.ChatRepository
		Profile      repositories.UserProfileRepository
		Payment      repositories.PaymentRepository 
		Admin        repositories.AdminRepository
	}

	func setupRepositories() Repositories {
		return Repositories{
			User:         repositories.NewUserRepository(database.DB),
			Consultation: repositories.NewConsultationRepository(database.DB),
			Loan:         repositories.NewLoanRepository(database.DB),
			Wallet:       repositories.NewWalletRepository(database.DB),
			Chat:         repositories.NewChatRepository(database.DB),
			Profile:      repositories.NewUserProfileRepository(database.DB),
			Payment:      repositories.NewPaymentRepository(database.DB),
			Admin:        repositories.NewAdminRepository(database.DB), 
		}
	}

	// Struct to hold all our Services
	type Services struct {
		Auth         services.AuthService
		Consultation services.ConsultationService
		Loan         services.LoanService
		Wallet       services.WalletService
		Profile      services.UserProfileService
		ChatHub      *services.ChatHub
		Payment      services.PaymentService
		Admin        services.AdminService 
	}

	func setupServices(r Repositories, producer *utils.KafkaProducer) Services {
		hub := services.NewChatHub(r.Chat)
		go hub.Run()

		return Services{
			Auth:         services.NewAuthService(r.User),
			Consultation: services.NewConsultationService(r.Consultation),
			Loan:         services.NewLoanService(r.Loan),
			Wallet:       services.NewWalletService(r.Wallet, producer),
			Profile:      services.NewUserProfileService(r.Profile),
			ChatHub:      hub,
			Payment:      services.NewPaymentService(), 
			Admin:        services.NewAdminService(r.Admin), 
		}
	}

	func startConsumers(s Services) {
		broker := os.Getenv("KAFKA_BROKER")
		if broker == "" {
			broker = "localhost:9092"
		}

		paymentConsumer := consumers.NewPaymentConsumer(broker, "telemetry.payments", "payment-processor-group", s.Loan)
		go paymentConsumer.Start(context.Background())

		loanConsumer := consumers.NewLoanConsumer(broker, "telemetry.loans", "loan-processor-group", s.Loan)
		go loanConsumer.Start(context.Background())
	}

	func startCronJobs(s Services) {
		emiJob := jobs.NewEMICheckerJob(s.Loan)
		emiJob.Start()
	}

	func setupRoutes(app *fiber.App, s Services, r Repositories) {
		api := app.Group("/api")

		// Initialize Standard Controllers
		authController := auth_controller.NewAuthController(s.Auth)
		consultationController := consultation_controller.NewConsultationController(s.Consultation)
		loanController := loan_controller.NewLoanController(s.Loan)
		walletController := wallet_controller.NewWalletController(s.Wallet)
		chatController := chat_controller.NewChatController(s.ChatHub)
		profileController := user_profile_controller.NewUserProfileController(s.Profile)
		paymentController := payment_controller.NewPaymentController(s.Payment, r.Payment)

		// 👇 CHANGED: Initialize our ONE Unified Admin Controller
		adminController := admin_controller.NewAdminController(s.Admin)

		// Setup Standard Routes
		routes.SetupAuthRoutes(api, authController)
		routes.SetupConsultationRoutes(api, consultationController)
		routes.SetupLoanRoutes(api, loanController)
		routes.SetupWalletRoutes(api, walletController)
		routes.SetupChatRoutes(api, chatController)
		routes.SetupUserProfileRoutes(api, profileController)
		routes.SetupPaymentRoutes(api, paymentController)
		
		// 👇 CHANGED: Setup Admin Routes (passing just the ONE controller!)
		routes.SetupAdminRoutes(api,adminController)
	}