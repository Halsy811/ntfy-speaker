package main

import (
	"context"
	"fmt"
	pl "ntfy_speaker/listener"
	"ntfy_speaker/settings"
	"ntfy_speaker/speaker"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	settings := &settings.SettingsType{}
	settings.New()

	logger.Info("start server...",
		zap.String("server name", settings.ServerName),
		zap.String("server port", settings.ServerPort),
		zap.String("topik", settings.Topik),
		zap.String("url", settings.URL),
	)

	listener := pl.New(settings.URL)

	// Создаем контекст с возможностью отмены (для graceful shutdown)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Гарантируем очистку ресурсов при выходе

	// Настраиваем перехват сигналов ОС (например, Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Создаем канал для приема сообщений из модуля
	// Буферизация (например, 10) полезна, чтобы слушатель не блокировался,
	// если main.go на секунду задумается при обработке.
	msgChan := make(chan pl.NtfyMessageType, 10)

	go func() {
		logger.Info("запуск слушателя ntfy в фоновом режиме...")
		err := listener.Start(ctx, msgChan)
		if err != nil {
			logger.Panic("Слушатель остановлен с ошибкой: " + err.Error())
		}
		// Если слушатель упал, мы тоже можем захотеть завершить всю программу
		cancel()
	}()

	for {
		select {
		case msg := <-msgChan:

			err := speaker.Speak(msg, settings)
			if err != nil {
				logger.Fatal("ошибка отправки toast-уведомления", zap.Error(err))
			}

		// Вариант Б: Получен сигнал завершения от ОС (Ctrl+C)
		case <-sigChan:
			fmt.Println("\n Получен сигнал завершения. Остановка...")
			cancel()                           // Отменяем контекст, что заставит listener.Start() выйти
			time.Sleep(500 * time.Millisecond) // Даем горутине время корректно закрыться
			return                             // Выходим из main, программа завершается
		}

	}

}
