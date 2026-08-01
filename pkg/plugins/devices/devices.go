// Плагин управления устройствами
package devices

import (
	"log"
	"m4jordomo/internal/bus"
	"m4jordomo/internal/storage"
	"m4jordomo/internal/types"
	"sync"
	"time"
)

type DevicesPlugin struct {
	mu      sync.RWMutex
	states  map[string]bool
	storage *storage.Storage
}

var defaultDevices = map[string]bool{
	"light":   false,
	"heating": false,
	"door":    false,
}

func New(st *storage.Storage) *DevicesPlugin {
	return &DevicesPlugin{
		states:  make(map[string]bool),
		storage: st,
	}
}

func (p *DevicesPlugin) Name() string {
	return "devices"
}

func (p *DevicesPlugin) Init(b *bus.Bus) error {
	// 1. Пытаемся загрузить состояния из БД
	if err := p.loadStatesFromDB(); err != nil {
		log.Printf("[Devices] ⚠️ Не удалось загрузить состояния: %v", err)
	}

	// 2. Добавляем недостающие устройства (если БД пустая или старая)
	p.mu.Lock()
	for name, status := range defaultDevices {
		if _, exists := p.states[name]; !exists {
			p.states[name] = status
			p.storage.SetDeviceStatus(name, status)
		}
	}

	// Удаляем устройства, которых больше нет в defaultDevices
	for name := range p.states {
		if _, exists := defaultDevices[name]; !exists {
			delete(p.states, name)
			if err := p.storage.DeleteDevice(name); err != nil {
				log.Printf("[Devices] ❌ Ошибка удаления %s: %v", name, err)
			}
		}
	}
	p.mu.Unlock()

	// 3. Проверяем, загрузились ли состояния. Если нет (все еще пусто) — создаем дефолтные принудительно
	p.mu.RLock()
	empty := len(p.states) == 0
	p.mu.RUnlock()

	if empty {
		log.Println("[Devices] База данных пуста. Создаю дефолтные устройства...")
		p.setDefaults()
	}

	log.Printf("[Devices] 📋 Загружены состояния: %v", p.states)

	log.Println("[Devices] Подписываюсь на события...")

	b.Subscribe(types.EventCommandDeviceSet, func(e types.Event) {
		p.handleSetDevice(e, b)
	})

	b.Subscribe(types.EventCommandDeviceGetAll, func(_ types.Event) {
		p.handleGetAllDevices(b)
	})

	log.Println("[Devices] Подписки выполнены.")

	return nil
}

func (p *DevicesPlugin) loadStatesFromDB() error {
	states, err := p.storage.GetAllDevices()
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, s := range states {
		p.states[s.Name] = s.Status
	}
	return nil
}

func (p *DevicesPlugin) setDefaults() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, status := range defaultDevices {
		p.states[name] = status
		// Сохраняем в БД
		if err := p.storage.SetDeviceStatus(name, status); err != nil {
			log.Printf("[Devices] Ошибка сохранения %s: %v", name, err)
		}
	}
	log.Println("[Devices] ✅ Дефолтные устройства созданы и сохранены в БД")
}

func (p *DevicesPlugin) handleSetDevice(e types.Event, bus *bus.Bus) {
	cmd, ok := e.Payload.(types.DeviceCommand)
	if !ok {
		log.Printf("[Devices] Ошибка: неверный формат payload: %v", e.Payload)
		return
	}
	p.mu.Lock()
	_, exists := p.states[cmd.Name]
	if !exists {
		p.mu.Unlock()
		log.Printf("[Devices] ❌ Устройство '%s' не найдено", cmd.Name)
		return
	}
	p.states[cmd.Name] = cmd.Status
	p.mu.Unlock()

	if err := p.storage.SetDeviceStatus(cmd.Name, cmd.Status); err != nil {
		log.Printf("[Devices] ❌ Ошибка сохранения в БД: %v", err)
	}
	log.Printf("[Devices] 💡 %s теперь %v", cmd.Name, cmd.Status)

	bus.Publish(types.Event{
		Type:     types.EventDeviceStateChanged,
		Priority: types.Medium,
		Payload: types.DeviceState{
			Name:      cmd.Name,
			Status:    cmd.Status,
			UpdatedAt: time.Now(),
		},
	})
}

func (p *DevicesPlugin) handleGetAllDevices(bus *bus.Bus) {
	p.mu.RLock()
	states := make(map[string]bool, len(p.states))
	for name, status := range p.states {
		states[name] = status
	}
	p.mu.RUnlock()

	bus.Publish(types.Event{
		Type:     types.EventDeviceStateResponse,
		Priority: types.Medium,
		Payload: types.DeviceStateList{
			States: states,
		},
	})
}
