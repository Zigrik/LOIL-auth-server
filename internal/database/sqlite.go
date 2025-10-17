package database

import (
	"database/sql"
	"fmt"
	"time"

	"LOIL-auth-server/internal/models"

	_ "modernc.org/sqlite"
)

type SQLiteDB struct {
	db *sql.DB
}

func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &SQLiteDB{db: db}, nil
}

func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// Проверка уникальности полей
func (s *SQLiteDB) CheckUniqueFields(login, gameSurname, email string) error {
	var count int

	// Проверка логина
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE login = ?", login).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("login already exists")
	}

	// Проверка игровой фамилии
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE game_surname = ?", gameSurname).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("game surname already exists")
	}

	// Проверка email
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("email already exists")
	}

	return nil
}

// Создание пользователя
func (s *SQLiteDB) CreateUser(user *models.User) error {
	if err := s.CheckUniqueFields(user.Login, user.GameSurname, user.Email); err != nil {
		return err
	}

	insertSQL := `
	INSERT INTO users (login, game_surname, email, password) 
	VALUES (?, ?, ?, ?)
	`

	result, err := s.db.Exec(insertSQL, user.Login, user.GameSurname, user.Email, user.Password)
	if err != nil {
		return err
	}

	// Получаем ID созданного пользователя
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)
	return nil
}

// Получение пользователя по логину
func (s *SQLiteDB) GetUserByLogin(login string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, login, game_surname, email, password, created_at, updated_at 
		FROM users WHERE login = ?
	`, login).Scan(&user.ID, &user.Login, &user.GameSurname, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Получение пользователя по ID
func (s *SQLiteDB) GetUserByID(id int) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, login, game_surname, email, password, created_at, updated_at 
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Login, &user.GameSurname, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Получение пользователя по email
func (s *SQLiteDB) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, login, game_surname, email, password, created_at, updated_at 
		FROM users WHERE email = ?
	`, email).Scan(&user.ID, &user.Login, &user.GameSurname, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Обновление профиля пользователя
func (s *SQLiteDB) UpdateUser(userID int, updates map[string]string) error {
	// Базовая проверка - нельзя обновлять пароль через этот метод
	if _, exists := updates["password"]; exists {
		return fmt.Errorf("password cannot be updated through this method")
	}

	// Начинаем построение запроса
	query := "UPDATE users SET "
	params := []interface{}{}
	first := true

	// Динамически добавляем поля для обновления
	for field, value := range updates {
		if !first {
			query += ", "
		}
		query += field + " = ?"
		params = append(params, value)
		first = false
	}

	// Добавляем обновление времени и условие WHERE
	query += ", updated_at = ? WHERE id = ?"
	params = append(params, time.Now(), userID)

	_, err := s.db.Exec(query, params...)
	return err
}

// Обновление пароля
func (s *SQLiteDB) UpdatePassword(userID int, newPasswordHash string) error {
	_, err := s.db.Exec(`
		UPDATE users SET password = ?, updated_at = ? 
		WHERE id = ?
	`, newPasswordHash, time.Now(), userID)

	return err
}

// Проверка существования пользователя
func (s *SQLiteDB) UserExists(login string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE login = ?", login).Scan(&count)
	return count > 0, err
}
