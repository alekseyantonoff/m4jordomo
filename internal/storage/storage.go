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

// DeadLetterRecord — запись в таблице dead_letters
type DeadLetterRecord struct {
	ID        int64  // автоинкремент из БД
	EventType string // тип события, которое не доставилось
	Priority  int    // приоритет (число)
	Payload   string // сериализованный payload в JSON
	Reason    string // причина провала
	Attempts  int    // сколько раз пытались
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

	createDeadLettersTable := `
	CREATE TABLE IF NOT EXISTS dead_letters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT NOT NULL,
		priority INTEGER NOT NULL,
		payload TEXT NOT NULL,
		reason TEXT NOT NULL,
		attempts INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := s.db.Exec(createDeadLettersTable); err != nil {
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

// DeleteDevice — удаляет устройство из БД
func (s *Storage) DeleteDevice(name string) error {
	query := `DELETE FROM devices WHERE name = ?`
	_, err := s.db.Exec(query, name)
	return err
}

// SaveDeadLetter — сохраняет запись в DLQ
func (s *Storage) SaveDeadLetter(rec DeadLetterRecord) error {
	query := `
	INSERT INTO dead_letters (event_type, priority, payload, reason, attempts)
	VALUES (?, ?, ?, ?, ?);
	`
	_, err := s.db.Exec(query, rec.EventType, rec.Priority, rec.Payload, rec.Reason, rec.Attempts)
	return err
}

// GetDeadLetters — возвращает все записи из DLQ
func (s *Storage) GetDeadLetters() ([]DeadLetterRecord, error) {
	rows, err := s.db.Query("SELECT id, event_type, priority, payload, reason, attempts FROM dead_letters ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []DeadLetterRecord
	for rows.Next() {
		var r DeadLetterRecord
		if err := rows.Scan(&r.ID, &r.EventType, &r.Priority, &r.Payload, &r.Reason, &r.Attempts); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// DeleteDeadLetter — удаляет запись из DLQ (после успешного реплея)
func (s *Storage) DeleteDeadLetter(id int64) error {
	query := `DELETE FROM dead_letters WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

// Close — закрывает соединение с БД
func (s *Storage) Close() error {
	return s.db.Close()
}
