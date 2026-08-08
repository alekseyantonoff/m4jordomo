// Плагин автоматизаций: «если [Trigger] и [Check] → [Actions]»
package rules

import (
	"log"

	"m4jordomo/internal/bus"
	"m4jordomo/internal/types"
	"sync"
)

type RulesPlugin struct {
	mu    sync.RWMutex
	rules []types.Rule
	last  map[string]bool // последнее известное состояние устройств
}

var defaultRules = []types.Rule{
	{
		ID:       "scene.arrive",
		Version:  "1.0.0",
		Priority: types.High,
		Layer:    "core",
		Enabled:  true,
		Trigger:  []types.Trigger{{Event: types.EventCommandScene, Scene: "arrive"}},
		Actions: []types.Action{
			{Event: types.EventCommandDeviceSet, Device: "light", Status: true},
			{Event: types.EventCommandDeviceSet, Device: "heating", Status: true},
		},
	},
	{
		ID:       "scene.leave",
		Version:  "1.0.0",
		Priority: types.High,
		Layer:    "core",
		Enabled:  true,
		Trigger:  []types.Trigger{{Event: types.EventCommandScene, Scene: "leave"}},
		Actions: []types.Action{
			{Event: types.EventCommandDeviceSet, Device: "light", Status: false},
			{Event: types.EventCommandDeviceSet, Device: "heating", Status: false},
		},
	},
	{
		ID:       "econ.draft",
		Version:  "1.0.0",
		Priority: types.Medium,
		Layer:    "core",
		Enabled:  true,
		Trigger:  []types.Trigger{{Event: types.EventDeviceStateChanged, Device: "door", Status: boolPtr(true)}},
		Check:    []types.Check{{Device: "heating", Status: true}},
		Actions: []types.Action{
			{Event: types.EventCommandDeviceSet, Device: "heating", Status: false},
		},
	},
}

func New() *RulesPlugin {
	return &RulesPlugin{
		rules: defaultRules,
		last:  make(map[string]bool),
	}
}

func (p *RulesPlugin) Name() string {
	return "rules"
}

func (p *RulesPlugin) Init(b *bus.Bus) error {
	log.Println("[Rules] ⚙️ Автоматизации запущены:")
	for _, r := range p.rules {
		log.Printf("  - [%s] %s", r.Priority.String(), r.ID)
	}

	b.Subscribe(types.EventDeviceStateChanged, func(e types.Event) {
		p.handleStateChanged(e, b)
	})
	b.Subscribe(types.EventCommandScene, func(e types.Event) {
		p.handleScene(e, b)
	})

	log.Println("[Rules] Подписки выполнены.")
	return nil
}

func (p *RulesPlugin) handleStateChanged(e types.Event, b *bus.Bus) {
	state, ok := e.Payload.(types.DeviceState)
	if !ok {
		log.Printf("[Rules] ❌ Ошибка: неверный формат payload: %v", e.Payload)
		return
	}

	// Защита от повторов: реагируем только на реальное ИЗМЕНЕНИЕ состояния
	p.mu.Lock()
	prev, seen := p.last[state.Name]
	p.last[state.Name] = state.Status
	p.mu.Unlock()
	if seen && prev == state.Status {
		return
	}

	p.evaluate(e, b)
}

func (p *RulesPlugin) handleScene(e types.Event, b *bus.Bus) {
	if _, ok := e.Payload.(types.Scene); !ok {
		log.Printf("[Rules] ❌ Ошибка: неверный формат payload: %v", e.Payload)
		return
	}
	p.evaluate(e, b)
}

func (p *RulesPlugin) evaluate(e types.Event, b *bus.Bus) {
	for _, r := range p.rules {
		if !r.Enabled {
			continue
		}
		if !triggerMatches(r, e) {
			continue
		}
		if !p.checkPasses(r) {
			continue
		}
		p.fire(r, b)
	}
}

func triggerMatches(r types.Rule, e types.Event) bool {
	for _, t := range r.Trigger {
		if t.Event != e.Type {
			continue
		}
		switch e.Type {
		case types.EventDeviceStateChanged:
			state, ok := e.Payload.(types.DeviceState)
			if !ok {
				continue
			}
			if t.Device != "" && t.Device != state.Name {
				continue
			}
			if t.Status != nil && *t.Status != state.Status {
				continue
			}
			return true
		case types.EventCommandScene:
			scene, ok := e.Payload.(types.Scene)
			if !ok {
				continue
			}
			if t.Scene != "" && t.Scene != scene.Name {
				continue
			}
			return true
		}
	}
	return false
}

func (p *RulesPlugin) checkPasses(r types.Rule) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, c := range r.Check {
		s, ok := p.last[c.Device]
		if !ok || s != c.Status {
			return false
		}
	}
	return true
}

func (p *RulesPlugin) fire(r types.Rule, b *bus.Bus) {
	log.Printf("[Rules] ⚡ Правило '%s' сработало → %d действие(й)", r.ID, len(r.Actions))
	for _, a := range r.Actions {
		switch a.Event {
		case types.EventCommandDeviceSet:
			b.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   a.Device,
					Status: a.Status,
				},
			})
		default:
			log.Printf("[Rules] ⚠️ Неизвестное действие: %s", a.Event)
		}
	}
}

func boolPtr(v bool) *bool {
	return &v
}
