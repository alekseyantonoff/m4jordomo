// Модель правила автоматизации (спека в каталог-автоматизаций.md)
package types

// Rule — правило: «если [Trigger] и [Check] → [Actions]»
type Rule struct {
	ID       string
	Version  string
	Priority Priority
	Layer    string // "core" | "steward"
	Enabled  bool
	Trigger  []Trigger
	Check    []Check
	Actions  []Action
}

// Trigger — на какое событие реагируем
type Trigger struct {
	Event  string
	Device string
	Status *bool
	Scene  string
	Metric string
	Op     string
	Value  float64
	From   string
	To     string
}

// Check — текущее состояние, которое должно быть правдой
type Check struct {
	Device string
	Status bool
}

// Action — действие (список = сцена)
type Action struct {
	Event  string
	Device string
	Status bool
}
