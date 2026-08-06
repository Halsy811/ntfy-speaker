package listener

import (
	"context"
	"encoding/json"
	"fmt"
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
	Message    string          `json:"message"`              // Текст самого уведомления
	Title      string          `json:"title,omitempty"`      // omitempty означает: если заголовка нет, не включать это поле
	Priority   int             `json:"priority"`             //
	Tags       []string        `json:"tags"`                 // Список тегов
	Click      string          `json:"click"`                // При нажатии
	Attachment *AttachmentType `json:"attachment,omitempty"` // Ссылка на прикрепленный файл
	Scenario   string          `json:"scenario"`
}

// Создает новый экземпляр слушателя
func New(url string) *ListenerType {
	return &ListenerType{
		url: url,
		// ВАЖНО: Для потоковых (streaming) соединений таймаут должен быть 0 (бесконечный),
		// иначе HTTP-клиент разорвет соединение через 30 секунд по умолчанию.
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

	// Явно указываем, что хотим получать JSON (хорошая практика)
	req.Header.Set("Accept", "application/json")

	// 2. Выполняем запрос через НАШ кастомный клиент (где Timeout: 0)
	resp, err := l.client.Do(req)
	if err != nil {
		// Если контекст был отменен во время установки соединения, вернем эту ошибку
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("ошибка выполнения HTTP-запроса: %w", err)
	}

	// ГАРАНТИРОВАННО закрываем тело ответа при выходе из функции, чтобы не было утечек памяти
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер вернул неожиданный статус: %d %s", resp.StatusCode, resp.Status)
	}

	// 3. Инициализируем потоковый декодер
	decoder := json.NewDecoder(resp.Body)

	// 4. Бесконечный цикл чтения
	for {
		var msg NtfyMessageType

		// Decode блокирует выполнение, пока не придет новый JSON или не закроется соединение
		if err := decoder.Decode(&msg); err != nil {
			// Если контекст отменен, возвращаем специальную ошибку контекста
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// Иначе это разрыв сети или ошибка парсинга
			return fmt.Errorf("ошибка декодирования или разрыв соединения: %w", err)
		}

		// Фильтруем только полезные сообщения, игнорируя keepalive
		if msg.Event == "message" {
			// Отправляем сообщение в канал.
			// Используем select, чтобы проверить, не отменили ли контекст во время отправки
			select {
			case msgChan <- msg:
				// Сообщение успешно отправлено в main.go
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
