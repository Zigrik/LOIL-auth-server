package main

import (
	"log"
	"net/http"
	"os"

	"LOIL-auth-server/internal/config"
	"LOIL-auth-server/internal/database"
	"LOIL-auth-server/internal/handlers"
	"LOIL-auth-server/internal/middleware"
	"LOIL-auth-server/pkg/logger"
)

func main() {
	// Инициализация логгера
	appLogger := logger.NewLogger()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		appLogger.Fatal("Failed to load config:", err)
	}

	// Инициализация базы данных
	db, err := database.NewSQLiteDB(cfg.DatabasePath)
	if err != nil {
		appLogger.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Запуск миграций
	migrationFS := os.DirFS("migrations")
	if err := db.RunMigrations(migrationFS); err != nil {
		log.Fatal("Migrations failed:", err)
	}

	// Инициализация обработчиков
	authHandler := handlers.NewAuthHandler(db, appLogger)
	profileHandler := handlers.NewProfileHandler(db, appLogger)

	// Настройка маршрутов
	router := http.NewServeMux()

	// Публичные маршруты
	router.HandleFunc("POST /api/auth/register", authHandler.Register)
	router.HandleFunc("POST /api/auth/login", authHandler.Login)

	// Защищенные маршруты
	router.Handle("GET /api/auth/profile", middleware.AuthMiddleware(profileHandler.GetProfile))
	router.Handle("PUT /api/auth/profile", middleware.AuthMiddleware(profileHandler.UpdateProfile))

	// Health check
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Настройка CORS middleware
	handler := middleware.CORS(router)

	appLogger.Info("Auth server starting on " + cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, handler); err != nil {
		appLogger.Fatal("Server failed:", err)
	}
}
