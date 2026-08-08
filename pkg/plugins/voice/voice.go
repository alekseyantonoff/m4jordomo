// Плагин голосового управления

package voice

import (
	"bufio"
	"log"
	"m4jordomo/internal/bus"
	"m4jordomo/internal/types"
	"os"
	"strings"
)

type VoicePlugin struct{}

func New() *VoicePlugin {
	return &VoicePlugin{}
}

func (p *VoicePlugin) Name() string {
	return "voice"
}

func (p *VoicePlugin) Init(bus *bus.Bus) error {
	log.Println("[Voice] 🎤 Запущен голосовой интерфейс (вводите команды в консоль)")
	log.Println("[Voice] Доступные команды:")
	log.Println("  - включи/выключи <устройство> (например: включи свет)")
	log.Println("  - открой/закрой дверь")
	log.Println("  - сломай дверь")
	log.Println("  - покажи ошибки")
	log.Println("  - повтори публикации из ошибок")
	log.Println("  - статус")
	log.Println("  - выход")
	go p.listenConsole(bus)
	return nil
}

func (p *VoicePlugin) listenConsole(bus *bus.Bus) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("[Voice] Ошибка чтения: %v", err)
			continue
		}
		text = strings.TrimSpace(strings.ToLower(text))

		switch {
		case text == "выход":
			log.Println("[Voice] 👋 Завершение работы...")
			os.Exit(0)
		case text == "статус":
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceGetAll,
				Priority: types.Medium,
				Payload:  nil,
			})
		case text == "покажи ошибки":
			bus.Publish(types.Event{
				Type:     types.EventCommandDeadLetterGetAll,
				Priority: types.Medium,
				Payload:  nil,
			})
		case text == "повтори публикации из ошибок":
			bus.Publish(types.Event{
				Type:     types.EventCommandDeadLetterReplay,
				Priority: types.Medium,
				Payload:  nil,
			})
		case strings.Contains(text, "включи свет"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "light",
					Status: true,
				},
			})
		case strings.Contains(text, "выключи свет"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "light",
					Status: false,
				},
			})
		case strings.Contains(text, "включи отопление"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "heating",
					Status: true,
				},
			})
		case strings.Contains(text, "выключи отопление"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "heating",
					Status: false,
				},
			})

		case strings.Contains(text, "открой дверь"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "door",
					Status: true,
				},
			})
		case strings.Contains(text, "закрой дверь"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceSet,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "door",
					Status: false,
				},
			})

		case strings.Contains(text, "сломай дверь"):
			bus.Publish(types.Event{
				Type:     types.EventCommandDeviceBreakIt,
				Priority: types.High,
				Payload: types.DeviceCommand{
					Name:   "door",
					Status: true,
				},
			})

		default:
			if verb, target, ok := parseToggle(text); ok {
				toggleDevice(bus, verb, target)
			} else {
				log.Printf("[Voice] 🤔 Не понял команду: '%s'", text)
			}
		}
	}
}

// parseToggle — разбирает «включи X», «выключи X», «открой X», «закрой X»
func parseToggle(text string) (verb string, target string, ok bool) {
	words := strings.Fields(text)
	for i, w := range words {
		switch w {
		case "включи", "включите", "открой", "откройте":
			return "on", strings.Join(words[i+1:], " "), true
		case "выключи", "выключите", "закрой", "закройте":
			return "off", strings.Join(words[i+1:], " "), true
		}
	}
	return "", "", false
}

// toggleDevice — публикует команду; разрешение имени и алиасов делает реестр (devices)
func toggleDevice(bus *bus.Bus, verb, target string) {
	if target == "" {
		log.Printf("[Voice] 🤔 Не понял команду: не указано устройство")
		return
	}
	bus.Publish(types.Event{
		Type:     types.EventCommandDeviceSet,
		Priority: types.High,
		Payload: types.DeviceCommand{
			Name:   target,
			Status: verb == "on",
		},
	})
}
