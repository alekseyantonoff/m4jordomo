// Общие типы данных
package types

import "time"

// Event — сообщение, которое передается по шине
type Event struct {
	Type       string      // Например, "device.light.toggle"
	Priority   Priority    // Насколько важно событие
	Payload    interface{} // Типизированные данные: DeviceCommand, DeviceState и т.д.
	Timestamp  time.Time   // Время создания
	RetryCount int         // Сколько раз уже пытались доставить
}

// Priority — уровень важности события
type Priority int

// Константы для Priority
const (
	Critical Priority = iota // 0 — самое важное (пожар, протечка)
	High                     // 1 — важное (команда пользователя)
	Medium                   // 2 — обычное (статистика)
	Low                      // 3 — фоновое (логи)
)

// String — превращает Priority в текст
func (p Priority) String() string {
	names := []string{"🔴 CRITICAL", "🟠 HIGH", "🟡 MEDIUM", "🟢 LOW"}
	if int(p) < len(names) {
		return names[p]
	}
	return "UNKNOWN"
}
