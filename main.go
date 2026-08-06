package main

import (
	"context"
	"fmt"
	pl "ntfy-speaker/listener"
	"ntfy-speaker/settings"
	"ntfy-speaker/speaker"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {

	settings := &settings.SettingsType{}
	settings.New()

	// logger -> console/file
	encoderConsole := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writeConsole := zapcore.AddSync(os.Stdout)

	encoderFile := zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
	writeFile := zapcore.AddSync(&lumberjack.Logger{
		Filename:   settings.LogFilePath,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     14,   //days
		Compress:   true, // disabled by default
	})

	core := zapcore.NewTee(
		zapcore.NewCore(encoderConsole, writeConsole, zap.InfoLevel),
		zapcore.NewCore(encoderFile, writeFile, zap.DebugLevel),
	)

	logger := zap.New(core, zap.AddCaller())
	defer logger.Sync()

	logger.Info("start server...",
		zap.String("server name", settings.ServerName),
		zap.String("server port", settings.ServerPort),
		zap.String("topik", settings.Topik),
		zap.String("url", settings.URL),
	)

	logger.Info(fmt.Sprintf("LOG file: %s", settings.LogFilePath))

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
		logger.Info("запуск слушателя ntfy...")
		err := listener.Start(ctx, msgChan)
		if err != nil {
			logger.Panic("Слушатель остановлен: " + err.Error())
		}
		// Если слушатель упал, мы тоже можем захотеть завершить всю программу
		cancel()
	}()

	for {
		select {
		case msg := <-msgChan:

			err := speaker.Speak(msg, settings)
			if err != nil {
				logger.Info("ошибка отправки toast-уведомления", zap.Error(err))
			}

		// Вариант Б: Получен сигнал завершения от ОС (Ctrl+C)
		case <-sigChan:
			logger.Info("Получен сигнал завершения...")
			cancel()                           // Отменяем контекст, что заставит listener.Start() выйти
			time.Sleep(500 * time.Millisecond) // Даем горутине время корректно закрыться
			return                             // Выходим из main, программа завершается
		}

	}

}
