// База данных
package storage

import (
	"database/sql"
	"log"
	_ "modernc.org/sqlite"
)

// Storage — хранилище данных
type Storage struct {
	db *sql.DB
}

// DeviceState — состояние одного устройства
type DeviceState struct {
	Name   string
	Status bool
}

// New — создает новое хранилище и открывает БД
func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &Storage{db: db}
	if err := s.initTables(); err != nil {
		return nil, err
	}
	log.Println("[Storage] База данных инициализирована")
	return s, nil
}

// initTables — создает таблицы при первом запуске
func (s *Storage) initTables() error {
	createDevicesTable := `
	CREATE TABLE IF NOT EXISTS devices (
		name TEXT PRIMARY KEY,
		status INTEGER NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := s.db.Exec(createDevicesTable); err != nil {
		return err
	}

	createEventsTable := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		priority INTEGER NOT NULL,
		payload TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := s.db.Exec(createEventsTable); err != nil {
		return err
	}
	return nil
}

// SetDeviceStatus — сохраняет состояние устройства
func (s *Storage) SetDeviceStatus(name string, status bool) error {
	query := `
	INSERT INTO devices (name, status, updated_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(name) DO UPDATE SET
		status = excluded.status,
		updated_at = CURRENT_TIMESTAMP;
	`
	_, err := s.db.Exec(query, name, status)
	return err
}

// GetAllDevices — загружает все состояния устройств
func (s *Storage) GetAllDevices() ([]DeviceState, error) {
	rows, err := s.db.Query("SELECT name, status FROM devices ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []DeviceState
	for rows.Next() {
		var d DeviceState
		if err := rows.Scan(&d.Name, &d.Status); err != nil {
			return nil, err
		}
		states = append(states, d)
	}
	return states, nil
}

// Close — закрывает соединение с БД
func (s *Storage) Close() error {
	return s.db.Close()
}