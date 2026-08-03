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
// Правило: приоритет = гарантия доставки, а не «важность».
// Critical/High не теряются: ретраи + DLQ при провале.
// Medium/Low — best-effort: потеря допустима.
const (
	Critical Priority = iota // 0 — аварии: пожар, протечка, безопасность
	High                     // 1 — команды, меняющие состояние (включи/выключи/открой)
	Medium                   // 2 — запросы и ответы (статус, покажи ошибки, response)
	Low                      // 3 — телеметрия, фоновые данные
)

// String — превращает Priority в текст
func (p Priority) String() string {
	names := []string{"🔴 CRITICAL", "🟠 HIGH", "🟡 MEDIUM", "🟢 LOW"}
	if int(p) < len(names) {
		return names[p]
	}
	return "UNKNOWN"
}
