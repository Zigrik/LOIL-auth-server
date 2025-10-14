package database

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func (s *SQLiteDB) RunMigrations(migrationFS fs.FS) error {
	// Создаем таблицу для отслеживания миграций
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Получаем список примененных миграций
	appliedVersions := make(map[int]bool)
	rows, err := s.db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("scan migration version: %w", err)
		}
		appliedVersions[version] = true
	}

	// Загружаем миграции из файловой системы
	migrations, err := s.loadMigrations(migrationFS)
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	// Применяем непримененные миграции
	for _, migration := range migrations {
		if !appliedVersions[migration.Version] {
			if err := s.applyMigration(migration); err != nil {
				return fmt.Errorf("apply migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

func (s *SQLiteDB) loadMigrations(migrationFS fs.FS) ([]Migration, error) {
	var migrations []Migration

	// Читаем все .sql файлы из папки миграций
	err := fs.WalkDir(migrationFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(d.Name()) == ".sql" {
			migration, err := s.parseMigrationFile(migrationFS, d.Name())
			if err != nil {
				return err
			}
			migrations = append(migrations, migration)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Сортируем миграции по версии
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (s *SQLiteDB) parseMigrationFile(migrationFS fs.FS, filename string) (Migration, error) {
	// Извлекаем версию из имени файла: 001_create_users.sql -> 1
	versionStr := strings.Split(filename, "_")[0]
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return Migration{}, fmt.Errorf("invalid migration filename %s: %w", filename, err)
	}

	// Читаем содержимое файла
	content, err := fs.ReadFile(migrationFS, filename)
	if err != nil {
		return Migration{}, fmt.Errorf("read migration file %s: %w", filename, err)
	}

	// Простой парсинг Up и Down секций
	sqlContent := string(content)
	upSQL, downSQL := parseMigrationSections(sqlContent)

	return Migration{
		Version: version,
		Name:    filename,
		UpSQL:   upSQL,
		DownSQL: downSQL,
	}, nil
}

func parseMigrationSections(content string) (string, string) {
	// Простой парсер: разделяем по комментариям -- +migrate Up/Down
	lines := strings.Split(content, "\n")
	var upLines, downLines []string
	currentSection := ""

	for _, line := range lines {
		if strings.Contains(line, "-- +migrate Up") {
			currentSection = "up"
			continue
		} else if strings.Contains(line, "-- +migrate Down") {
			currentSection = "down"
			continue
		}

		if currentSection == "up" {
			upLines = append(upLines, line)
		} else if currentSection == "down" {
			downLines = append(downLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(upLines, "\n")),
		strings.TrimSpace(strings.Join(downLines, "\n"))
}

func (s *SQLiteDB) applyMigration(migration Migration) error {
	fmt.Printf("Applying migration: %s\n", migration.Name)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Выполняем Up миграцию
	if _, err := tx.Exec(migration.UpSQL); err != nil {
		return fmt.Errorf("execute up migration: %w", err)
	}

	// Записываем в таблицу миграций
	_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Version)
	if err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit()
}

// Добавь в начало файла (рядом с структурой SQLiteDB)
type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}
