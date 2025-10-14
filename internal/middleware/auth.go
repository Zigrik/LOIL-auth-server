package middleware

import (
	"LOIL-auth-server/internal/utils"
	"LOIL-auth-server/internal/config"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}
		
		// Формат: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "Invalid authorization format"}`, http.StatusUnauthorized)
			return
		}
		
		tokenString := parts[1]
		cfg, _ := config.Load()
		
		claims, err := utils.ValidateJWT(tokenString, cfg.JWTSecret)
		if err != nil {
			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
			return
		}
		
		// Добавляем данные пользователя в контекст
		ctx := r.Context()
		// ctx = context.WithValue(ctx, "userID", claims.UserID)
		// ctx = context.WithValue(ctx, "userLogin", claims.Login)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}