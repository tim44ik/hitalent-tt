package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hitalent-test/internal/config"
	"hitalent-test/internal/handlers"
	"hitalent-test/internal/repositories"
	"hitalent-test/internal/server"
	"hitalent-test/internal/services"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	config *config.Config
	db     *gorm.DB
	srv    *http.Server
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Println("Database connected")

	deptRepo := repositories.NewDepartmentRepository(db)
	empRepo := repositories.NewEmployeeRepository(db)

	deptService := services.NewDepartmentService(deptRepo)
	empService := services.NewEmployeeService(empRepo, deptRepo)

	deptHandler := handlers.NewDepartmentHandler(deptService)
	empHandler := handlers.NewEmployeeHandler(empService)

	handlers.SetReady(true)

	router := server.NewRouter(deptHandler, empHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &App{
		config: cfg,
		db:     db,
		srv:    srv,
	}, nil
}

func (a *App) Run() error {
	go func() {
		log.Printf("Starting server on %s", a.srv.Addr)
		if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	sqlDB, _ := a.db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
	log.Println("Server exited")
	return nil
}
