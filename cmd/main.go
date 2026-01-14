package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Thorlik/marketplace/internal/app/config"
	"github.com/Thorlik/marketplace/internal/app/handler"
	"github.com/Thorlik/marketplace/internal/app/middleware"
	"github.com/Thorlik/marketplace/internal/pkg/db"
	"github.com/Thorlik/marketplace/internal/repository"
	"github.com/Thorlik/marketplace/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	dbConfig := &db.DatabaseConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}
	connManager, err := db.NewConnectionManager(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer connManager.Close()

	queryManager := db.NewQueryManager(connManager.GetDB(), logger)
	txManager := db.NewTransactionManager(connManager.GetDB(), logger)

	userRepo := repository.NewUserRepository(txManager, queryManager, logger)
	productRepo := repository.NewProductRepository(txManager, queryManager, logger)
	orderRepo := repository.NewOrderRepository(txManager, queryManager, logger)

	userService := service.NewUserService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiryHours, logger)
	productService := service.NewProductService(productRepo, logger)
	orderService := service.NewOrderService(orderRepo, productRepo, logger)

	userHandler := handler.NewUserHandler(userService, logger)
	productHandler := handler.NewProductHandler(productService, logger)
	orderHandler := handler.NewOrderHandler(orderService, logger)

	corsMiddleware := middleware.NewCORSMiddleware([]string{"*"})
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)
	authMiddleware := middleware.NewAuthMiddleware(userService, logger)

	r := mux.NewRouter()

	r.Use(corsMiddleware.EnableCORS)
	r.Use(loggingMiddleware.Log)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods(http.MethodGet)

	api := r.PathPrefix("/api/v1").Subrouter()

	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", userHandler.Register).Methods(http.MethodPost)
	auth.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)

	users := api.PathPrefix("/users").Subrouter()

	usersProtected := users.NewRoute().Subrouter()
	usersProtected.Use(authMiddleware.Authenticate)
	usersProtected.HandleFunc("/me", userHandler.GetProfile).Methods(http.MethodGet)
	usersProtected.HandleFunc("/me", userHandler.UpdateProfile).Methods(http.MethodPut)
	usersProtected.HandleFunc("/me/password", userHandler.ChangePassword).Methods(http.MethodPut)

	usersAdmin := users.NewRoute().Subrouter()
	usersAdmin.Use(authMiddleware.Authenticate)
	usersAdmin.Use(authMiddleware.RequireAdmin)
	usersAdmin.HandleFunc("", userHandler.GetAllUsers).Methods(http.MethodGet)
	usersAdmin.HandleFunc("/{id}", userHandler.GetUser).Methods(http.MethodGet)
	usersAdmin.HandleFunc("/{id}/role", userHandler.UpdateUserRole).Methods(http.MethodPut)
	usersAdmin.HandleFunc("/{id}", userHandler.DeleteUser).Methods(http.MethodDelete)

	products := api.PathPrefix("/products").Subrouter()

	productsProtected := products.NewRoute().Subrouter()
	productsProtected.Use(authMiddleware.Authenticate)
	productsProtected.HandleFunc("/my", productHandler.GetMyProducts).Methods(http.MethodGet)
	productsProtected.HandleFunc("", productHandler.CreateProduct).Methods(http.MethodPost)
	productsProtected.HandleFunc("/{id}", productHandler.UpdateProduct).Methods(http.MethodPut)
	productsProtected.HandleFunc("/{id}", productHandler.DeleteProduct).Methods(http.MethodDelete)

	products.HandleFunc("", productHandler.GetAllProducts).Methods(http.MethodGet)
	products.HandleFunc("/{id}", productHandler.GetProduct).Methods(http.MethodGet)

	productsMod := api.PathPrefix("/moderate/products").Subrouter()
	productsMod.Use(authMiddleware.Authenticate)
	productsMod.Use(authMiddleware.RequireModerator)
	productsMod.HandleFunc("/pending", productHandler.GetPendingProducts).Methods(http.MethodGet)
	productsMod.HandleFunc("/{id}", productHandler.ModerateProduct).Methods(http.MethodPost)

	cart := api.PathPrefix("/cart").Subrouter()
	cart.Use(authMiddleware.Authenticate)
	cart.HandleFunc("", orderHandler.GetCart).Methods(http.MethodGet)
	cart.HandleFunc("", orderHandler.AddToCart).Methods(http.MethodPost)
	cart.HandleFunc("/clear", orderHandler.ClearCart).Methods(http.MethodDelete)
	cart.HandleFunc("/{id}", orderHandler.UpdateCartItem).Methods(http.MethodPut)
	cart.HandleFunc("/{id}", orderHandler.RemoveFromCart).Methods(http.MethodDelete)

	orders := api.PathPrefix("/orders").Subrouter()
	orders.Use(authMiddleware.Authenticate)
	orders.HandleFunc("", orderHandler.GetMyOrders).Methods(http.MethodGet)
	orders.HandleFunc("", orderHandler.CreateOrder).Methods(http.MethodPost)
	orders.HandleFunc("/from-cart", orderHandler.CreateOrderFromCart).Methods(http.MethodPost)
	orders.HandleFunc("/{id}", orderHandler.GetOrder).Methods(http.MethodGet)

	payments := api.PathPrefix("/payments").Subrouter()
	payments.Use(authMiddleware.Authenticate)
	payments.HandleFunc("", orderHandler.ProcessPayment).Methods(http.MethodPost)

	ordersAdmin := api.PathPrefix("/admin/orders").Subrouter()
	ordersAdmin.Use(authMiddleware.Authenticate)
	ordersAdmin.Use(authMiddleware.RequireAdmin)
	ordersAdmin.HandleFunc("", orderHandler.GetAllOrders).Methods(http.MethodGet)
	ordersAdmin.HandleFunc("/{id}/status", orderHandler.UpdateOrderStatus).Methods(http.MethodPut)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Server starting", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server")

	if err := srv.Close(); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited properly")
}
