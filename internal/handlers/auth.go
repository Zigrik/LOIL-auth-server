package handlers

import (
	"encoding/json"
	"net/http"

	"LOIL-auth-server/internal/config"
	"LOIL-auth-server/internal/database"
	"LOIL-auth-server/internal/models"
	"LOIL-auth-server/internal/utils"
	"LOIL-auth-server/pkg/logger"
)

type AuthHandler struct {
	db     *database.SQLiteDB
	logger *logger.Logger
}

func NewAuthHandler(db *database.SQLiteDB, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		db:     db,
		logger: logger,
	}
}

type RegisterRequest struct {
	Login           string `json:"login"`
	GameSurname     string `json:"gameSurname"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Success bool        `json:"success"`
	Token   string      `json:"token,omitempty"`
	User    interface{} `json:"user,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Register: invalid JSON input")
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Invalid input"})
		return
	}

	// Валидация
	if !utils.ValidateLogin(req.Login) {
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Login must be 3-20 characters (letters, numbers, underscore)"})
		return
	}

	if !utils.ValidateGameSurname(req.GameSurname) {
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Game surname must contain only Latin letters (2-20 characters)"})
		return
	}

	if !utils.ValidateEmail(req.Email) {
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Invalid email format"})
		return
	}

	if !utils.ValidatePassword(req.Password) {
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Password must be at least 6 characters"})
		return
	}

	if req.Password != req.PasswordConfirm {
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Passwords do not match"})
		return
	}

	// Нормализуем фамилию
	normalizedSurname := utils.NormalizeGameSurname(req.GameSurname)

	// Хэшируем пароль
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("Register: password hashing failed:", err)
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Server error"})
		return
	}

	// Создаем пользователя
	user := &models.User{
		Login:       req.Login,
		GameSurname: normalizedSurname,
		Email:       req.Email,
		Password:    hashedPassword,
	}

	// Сохраняем в базу
	if err := h.db.CreateUser(user); err != nil {
		h.logger.Error("Register: database error:", err)

		errorMsg := "Registration failed"
		switch err.Error() {
		case "login already exists":
			errorMsg = "Login already exists"
		case "game surname already exists":
			errorMsg = "Game surname already exists"
		case "email already exists":
			errorMsg = "Email already exists"
		}

		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: errorMsg})
		return
	}

	// Генерируем JWT токен
	cfg, _ := config.Load()
	token, err := utils.GenerateJWT(user.ID, user.Login, user.GameSurname, cfg.JWTSecret)
	if err != nil {
		h.logger.Error("Register: JWT generation failed:", err)
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Server error"})
		return
	}

	h.logger.Info("Register: user created successfully -", user.Login)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Token:   token,
		User:    user.ToResponse(),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Login: invalid JSON input")
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Invalid input"})
		return
	}

	// Ищем пользователя
	user, err := h.db.GetUserByLogin(req.Login)
	if err != nil {
		h.logger.Error("Login: user not found -", req.Login)
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Invalid login or password"})
		return
	}

	// Проверяем пароль
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		h.logger.Error("Login: invalid password for user -", req.Login)
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Invalid login or password"})
		return
	}

	// Генерируем JWT
	cfg, _ := config.Load()
	token, err := utils.GenerateJWT(user.ID, user.Login, user.GameSurname, cfg.JWTSecret)
	if err != nil {
		h.logger.Error("Login: JWT generation failed:", err)
		json.NewEncoder(w).Encode(AuthResponse{Success: false, Error: "Server error"})
		return
	}

	h.logger.Info("Login: user logged in successfully -", user.Login)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Token:   token,
		User:    user.ToResponse(),
	})
}
