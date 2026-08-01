// Точка входа в ядро

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"m4jordomo/internal/bus"
	"m4jordomo/internal/storage"
	"m4jordomo/internal/types"
	"m4jordomo/pkg/plugins/devices"
	"m4jordomo/pkg/plugins/voice"
)

func main() {
	log.Println("🏠 Запуск m4jordomo...")

	store, err := storage.New("m4jordomo.db")
	if err != nil {
		log.Fatalf("❌ Ошибка инициализации БД: %v", err)
	}
	defer store.Close()

	b := bus.New()

	devicesPlugin := devices.New(store)
	voicePlugin := voice.New()

	if err := devicesPlugin.Init(b); err != nil {
		log.Fatalf("Ошибка инициализации Devices: %v", err)
	}
	if err := voicePlugin.Init(b); err != nil {
		log.Fatalf("Ошибка инициализации Voice: %v", err)
	}

	log.Println("✅ m4jordomo готов к работе!")
	log.Println("📝 Введите 'статус' для проверки состояния устройств")
	log.Println("📝 Введите 'выход' для завершения")

	b.Subscribe(types.EventDeviceStateChanged, func(e types.Event) {
		state, ok := e.Payload.(types.DeviceState)
		if !ok {
			log.Println("[m4jordomo] Ошибка: неверный формат payload")
			return
		}
		statusText := "🔴 ВЫКЛ"
		if state.Status {
			statusText = "🟢 ВКЛ"
		}
		log.Printf("[m4jordomo] 📢 %s -> %s", state.Name, statusText)
	})

	b.Subscribe(types.EventDeviceStateResponse, func(e types.Event) {
		list, ok := e.Payload.(types.DeviceStateList)
		if !ok {
			log.Println("[m4jordomo] Ошибка: не могу прочитать состояния")
			return
		}
		log.Println("=== Состояние устройств ===")
		for name, status := range list.States {
			statusText := "🔴 ВЫКЛ"
			if status {
				statusText = "🟢 ВКЛ"
			}
			log.Printf("  %s: %s", name, statusText)
		}
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("👋 m4jordomo завершает работу...")
}
