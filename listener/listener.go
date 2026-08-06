package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ListenerType struct {
	url    string
	client *http.Client
}

type AttachmentType struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Size    int    `json:"size"`
	Expires int64  `json:"expires"`
	URL     string `json:"url"`
}

type NtfyMessageType struct {
	ID         string          `json:"id"`
	Time       int64           `json:"time"`
	Event      string          `json:"event"` // Тип события: "message", "keepalive", "open"
	Topic      string          `json:"topic"`
	Message    string          `json:"message"`         // Текст уведомления
	Title      string          `json:"title,omitempty"` // Заголовок
	Priority   int             `json:"priority"`
	Tags       []string        `json:"tags"`
	Click      string          `json:"click"`
	Attachment *AttachmentType `json:"attachment,omitempty"`
	Scenario   string          `json:"scenario"`
}

// Создает новый экземпляр слушателя
func New(url string) *ListenerType {
	return &ListenerType{
		url: url,
		// Для потоковых соединений таймаут должен быть 0 (бесконечный)
		client: &http.Client{
			Timeout: 0,
		},
	}
}

func (l *ListenerType) Start(ctx context.Context, msgChan chan<- NtfyMessageType) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.url, nil)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Указываем, что хотим получать JSON
	req.Header.Set("Accept", "application/json")

	// Выполняем запрос
	resp, err := l.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("ошибка выполнения HTTP-запроса: %w", err)
	}

	// Закрываем тело ответа при выходе
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер вернул неожиданный статус: %d %s", resp.StatusCode, resp.Status)
	}

	// Создаем декодер для потокового чтения JSON
	decoder := json.NewDecoder(resp.Body)

	// Бесконечный цикл чтения сообщений
	for {
		var msg NtfyMessageType

		// Читаем следующее сообщение
		if err := decoder.Decode(&msg); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// Если сервер закрыл соединение штатно, возвращаем nil
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("ошибка декодирования или разрыв соединения: %w", err)
		}

		// Обрабатываем только реальные сообщения, игнорируем keepalive
		if msg.Event == "message" {
			select {
			case msgChan <- msg:
				// Сообщение отправлено
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
