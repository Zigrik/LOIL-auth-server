package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"LOIL-auth-server/internal/config"
	"LOIL-auth-server/internal/database"
	"LOIL-auth-server/internal/utils"
	"LOIL-auth-server/pkg/logger"
)

type ProfileHandler struct {
	db     *database.SQLiteDB
	logger *logger.Logger
}

func NewProfileHandler(db *database.SQLiteDB, logger *logger.Logger) *ProfileHandler {
	return &ProfileHandler{
		db:     db,
		logger: logger,
	}
}

type ProfileResponse struct {
	Success bool        `json:"success"`
	User    interface{} `json:"user,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type UpdateProfileRequest struct {
	GameSurname *string `json:"gameSurname,omitempty"`
	Email       *string `json:"email,omitempty"`
}

// Вспомогательная функция для извлечения userID из токена
func (h *ProfileHandler) getUserIDFromToken(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 { // "Bearer " + token
		return 0, fmt.Errorf("invalid authorization header")
	}

	tokenString := authHeader[7:] // Remove "Bearer "
	cfg, _ := config.Load()

	claims, err := utils.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		h.logger.Error("GetProfile: invalid token")
		json.NewEncoder(w).Encode(ProfileResponse{Success: false, Error: "Invalid token"})
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		h.logger.Error("GetProfile: user not found - ID:", userID)
		json.NewEncoder(w).Encode(ProfileResponse{Success: false, Error: "User not found"})
		return
	}

	h.logger.Info("GetProfile: profile retrieved for user -", user.Login)
	json.NewEncoder(w).Encode(ProfileResponse{
		Success: true,
		User:    user.ToResponse(),
	})
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		h.logger.Error("UpdateProfile: invalid token")
		json.NewEncoder(w).Encode(ProfileResponse{Success: false, Error: "Invalid token"})
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("UpdateProfile: invalid JSON input")
		json.NewEncoder(w).Encode(ProfileResponse{Success: false, Error: "Invalid input"})
		return
	}

	// Валидация обновляемых полей
	updates := make(map[string]string)

	if req.GameSurname != nil {
		if !utils.ValidateGameSurname(*req.GameSurname) {
			json.NewEncoder(w).Encode(ProfileResponse{Success: false, Error: "Game surname must contain only Latin letters"})
			return
		}
		normalizedSurname := utils.NormalizeGameSurname(*req.GameSurname)
		updates["game_surname"] = normalizedSurname
	}

	if req.Email != nil {
		if !utils.ValidateEmail(*req.Email) {
			json.NewEncoder(w).Encode(ProfileResponse{Success: false, Error: "Invalid email format"})
			return
		}
		updates["email"] = *req.Email
	}

	// Если есть что обновлять
	if len(updates) > 0 {
		if err := h.db.UpdateUser(userID, updates); err != nil {
			h.logger.Error("UpdateProfile: database error:", err)

			errorMsg := "Update failed"
			if err.Error() == "game surname already exists" {
				errorMsg = "Game surname already exists"
			} else if err.Error() == "email already exists" {
				errorMsg = "Email already exists"
			}

			json.NewEncoder(w).Encode(ProfileResponse{Success: false, Error: errorMsg})
			return
		}
	}

	// Получаем обновленные данные пользователя
	user, err := h.db.GetUserByID(userID)
	if err != nil {
		h.logger.Error("UpdateProfile: failed to get updated user - ID:", userID)
		json.NewEncoder(w).Encode(ProfileResponse{Success: false, Error: "Update completed but failed to retrieve user"})
		return
	}

	h.logger.Info("UpdateProfile: profile updated for user -", user.Login)
	json.NewEncoder(w).Encode(ProfileResponse{
		Success: true,
		User:    user.ToResponse(),
	})
}
