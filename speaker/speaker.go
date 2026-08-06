package speaker

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	pl "ntfy-speaker/listener"
	"ntfy-speaker/settings"
	"ntfy-speaker/toast"
)

// Встраиваем иконки в бинарник
//
//go:embed src/wtf-icon.png
var embeddedIcon_wtf []byte

//go:embed src/gooden-icon.png
var embeddedIcon_gooden []byte

//go:embed src/information-icon.png
var embeddedIcon_info []byte

//go:embed src/Hay-icon.png
var embeddedIcon_warn []byte

//go:embed src/warning-icon.png
var embeddedIcon_err []byte

// Пути к извлеченным иконкам
var tempIconPath_wtf string
var tempIconPath_gooden string
var tempIconPath_info string
var tempIconPath_warn string
var tempIconPath_err string

var iconsInitialized bool
var iconMutex sync.Mutex

// InitIcons извлекает иконки во временную папку (вызывается один раз при старте)
func InitIcons(settings *settings.SettingsType) error {
	iconMutex.Lock()
	defer iconMutex.Unlock()

	if iconsInitialized {
		return nil
	}

	// Формируем пути к иконкам
	tempIconPath_wtf = filepath.Join(settings.TmpFolder, "wtf-icon.png")
	tempIconPath_gooden = filepath.Join(settings.TmpFolder, "gooden-icon.png")
	tempIconPath_info = filepath.Join(settings.TmpFolder, "information-icon.png")
	tempIconPath_warn = filepath.Join(settings.TmpFolder, "Hay-icon.png")
	tempIconPath_err = filepath.Join(settings.TmpFolder, "warning-icon.png")

	// Извлекаем каждую иконку, если её ещё нет
	icons := []struct {
		path string
		data []byte
	}{
		{tempIconPath_wtf, embeddedIcon_wtf},
		{tempIconPath_gooden, embeddedIcon_gooden},
		{tempIconPath_info, embeddedIcon_info},
		{tempIconPath_warn, embeddedIcon_warn},
		{tempIconPath_err, embeddedIcon_err},
	}

	for _, icon := range icons {
		if _, err := os.Stat(icon.path); err != nil {
			// Файл не существует, создаем его
			err := os.WriteFile(icon.path, icon.data, 0644)
			if err != nil {
				return fmt.Errorf("ошибка записи иконки %s: %w", icon.path, err)
			}
		}
	}

	iconsInitialized = true
	return nil
}

// Speak отправляет toast-уведомление
func Speak(msg pl.NtfyMessageType, settings *settings.SettingsType) error {
	var action []toast.Action

	notification := toast.Notification{
		AppID:   "NtfySpeaker - " + msg.Topic,
		Title:   msg.Title,
		Message: msg.Message,
	}

	// Добавляем кнопку с ссылкой, если есть
	if msg.Click != "" {
		notification.ActivationArguments = settings.ServerName + ":" + settings.ServerPort + "/" + settings.Topik
		action = append(action, toast.Action{
			Type:      "protocol",
			Label:     "Ссылка URL",
			Arguments: msg.Click,
		})
	} else {
		notification.ActivationArguments = settings.ServerName + ":" + settings.ServerPort + "/" + settings.Topik
	}

	// Выбираем иконку и длительность в зависимости от приоритета
	switch msg.Priority {
	case 0: // Средний приоритет
		notification.Icon = tempIconPath_info
		notification.Duration = "long"

	case 1: // Низкий приоритет
		notification.Duration = "short"
		notification.Icon = tempIconPath_wtf

	case 2: // Низкий приоритет
		notification.Duration = "short"
		notification.Icon = tempIconPath_gooden

	case 3: // Средний приоритет
		notification.Icon = tempIconPath_info
		notification.Duration = "long"

	case 4: // Высокий приоритет
		notification.Icon = tempIconPath_warn
		notification.Scenario = "reminder"

	case 5: // Высокий приоритет
		notification.Icon = tempIconPath_err
		notification.Scenario = "reminder"
	}

	// Добавляем вложение, если есть
	if msg.Attachment != nil {
		action = append(action, toast.Action{
			Type:      "protocol",
			Label:     msg.Attachment.Name,
			Arguments: msg.Attachment.URL,
		})
	}

	notification.Actions = action

	// Отправляем уведомление
	err := notification.Push()
	if err != nil {
		return fmt.Errorf("ошибка при отправке уведомления в панель уведомлений Windows: %w", err)
	}

	return nil
}
