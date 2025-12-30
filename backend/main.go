package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httphandler "splitwise/adapter/inbound/http"
	"splitwise/adapter/outbound/repository"
	"splitwise/application/service"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize Jaeger tracing
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://jaeger:14268/api/traces"
	}

	tp, err := httphandler.InitTracing("splitwise-backend", jaegerEndpoint)
	if err != nil {
		log.Printf("Failed to initialize tracing: %v", err)
		log.Println("Continuing without tracing...")
	} else {
		log.Println("Jaeger tracing initialized successfully")
		// Cleanly shutdown tracer on exit
		if tp != nil {
			defer func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := tp.Shutdown(ctx); err != nil {
					log.Printf("Error shutting down tracer provider: %v", err)
				}
			}()
		}
	}

	// Initialize database connections
	log.Println("Connecting to PostgreSQL...")
	db, err := repository.InitPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("PostgreSQL connected successfully")

	// Create database tables
	if err := repository.CreateTables(db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
	log.Println("Database tables initialized")

	// Initialize Redis
	log.Println("Connecting to Redis...")
	redisClient, err := repository.InitRedisClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connected successfully")

	// Initialize outbound adapters (repositories)
	userRepo := repository.NewPostgresUserRepository(db)
	authRepo := repository.NewPostgresAuthRepository(db)
	cache := repository.NewRedisCache(redisClient)
	expenseRepo := repository.NewPostgresExpenseRepository(db)
	groupRepo := repository.NewPostgresGroupRepository(db)

	// Initialize application services (implementing inbound ports)
	authService := service.NewAuthService(authRepo, cache)
	userService := service.NewUserService(userRepo)
	expenseService := service.NewExpenseService(expenseRepo, userRepo, groupRepo)
	groupService := service.NewGroupService(groupRepo, userRepo)

	// Initialize inbound adapters (HTTP handlers)
	authHandler := httphandler.NewAuthHandler(authService)
	userHandler := httphandler.NewUserHandler(userService)
	expenseHandler := httphandler.NewExpenseHandler(expenseService)
	groupHandler := httphandler.NewGroupHandler(groupService)

	// Setup HTTP routes
	router := mux.NewRouter()

	// Add tracing middleware
	router.Use(httphandler.TracingMiddleware())

	// Public auth routes
	router.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/auth/refresh", authHandler.RefreshToken).Methods("POST")

	// Protected routes (require authentication)
	protectedRouter := router.PathPrefix("").Subrouter()
	protectedRouter.Use(httphandler.AuthMiddleware(authService))

	// User routes
	router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	
	// Protected user-specific routes
	protectedRouter.HandleFunc("/users/{user_id}/groups", groupHandler.GetGroupsByUser).Methods("GET")
	protectedRouter.HandleFunc("/users/{user_id}/expenses", expenseHandler.GetExpensesByUser).Methods("GET")

	// Expense routes
	protectedRouter.HandleFunc("/expenses", expenseHandler.CreateExpense).Methods("POST")
	router.HandleFunc("/expenses/{id}", expenseHandler.GetExpense).Methods("GET")
	protectedRouter.HandleFunc("/expenses/{id}", expenseHandler.UpdateExpense).Methods("PUT")
	protectedRouter.HandleFunc("/expenses/{id}", expenseHandler.DeleteExpense).Methods("DELETE")
	router.HandleFunc("/groups/{group_id}/expenses", expenseHandler.GetExpensesByGroup).Methods("GET")

	// Group routes
	protectedRouter.HandleFunc("/groups", groupHandler.CreateGroup).Methods("POST")
	router.HandleFunc("/groups", groupHandler.GetAllGroups).Methods("GET")
	router.HandleFunc("/groups/{id}", groupHandler.GetGroup).Methods("GET")
	protectedRouter.HandleFunc("/groups/{id}", groupHandler.UpdateGroup).Methods("PUT")
	protectedRouter.HandleFunc("/groups/{id}", groupHandler.DeleteGroup).Methods("DELETE")
	protectedRouter.HandleFunc("/groups/{id}/users", groupHandler.AddUserToGroup).Methods("POST")
	protectedRouter.HandleFunc("/groups/{id}/users/{user_id}", groupHandler.RemoveUserFromGroup).Methods("DELETE")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}).Methods("GET")

	// CORS configuration
	corsOrigins := handlers.AllowedOrigins([]string{"*"})
	corsMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	corsHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	// Wrap router with CORS middleware
	handler := handlers.CORS(corsOrigins, corsMethods, corsHeaders)(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port

	fmt.Printf("Server starting on port %s\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  POST   /auth/register")
	fmt.Println("  POST   /auth/login")
	fmt.Println("  POST   /auth/refresh")
	fmt.Println("  POST   /users")
	fmt.Println("  GET    /users")
	fmt.Println("  GET    /users/{id}")
	fmt.Println("  GET    /users/{user_id}/groups (protected)")
	fmt.Println("  GET    /users/{user_id}/expenses (protected)")
	fmt.Println("  POST   /expenses (protected)")
	fmt.Println("  GET    /expenses/{id}")
	fmt.Println("  PUT    /expenses/{id} (protected)")
	fmt.Println("  DELETE /expenses/{id} (protected)")
	fmt.Println("  GET    /groups/{group_id}/expenses")
	fmt.Println("  POST   /groups (protected)")
	fmt.Println("  GET    /groups")
	fmt.Println("  GET    /groups/{id}")
	fmt.Println("  PUT    /groups/{id} (protected)")
	fmt.Println("  DELETE /groups/{id} (protected)")
	fmt.Println("  POST   /groups/{id}/users (protected)")
	fmt.Println("  DELETE /groups/{id}/users/{user_id} (protected)")
	fmt.Println("  GET    /health")

	// Setup graceful shutdown
	server := &http.Server{
		Addr:    port,
		Handler: handler,
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
