package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pl "ntfy-speaker/listener"
	"ntfy-speaker/settings"
	"ntfy-speaker/speaker"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// Используем cfg вместо settings, чтобы не перекрывать имя пакета
	cfg := &settings.SettingsType{}
	cfg.New()

	// Настраиваем логирование в консоль
	encoderConsole := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writeConsole := zapcore.AddSync(os.Stdout)

	// Настраиваем логирование в файл с ротацией
	encoderFile := zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
	writeFile := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.LogFilePath,
		MaxSize:    500, // мегабайты
		MaxBackups: 3,
		MaxAge:     14,   // дни
		Compress:   true, // сжимать старые логи
	})

	// Объединяем оба вывода в один logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoderConsole, writeConsole, zap.InfoLevel),
		zapcore.NewCore(encoderFile, writeFile, zap.DebugLevel),
	)

	logger := zap.New(core, zap.AddCaller())
	defer logger.Sync()

	logger.Info("start server...",
		zap.String("server name", cfg.ServerName),
		zap.String("server port", cfg.ServerPort),
		zap.String("topik", cfg.Topik),
		zap.String("url", cfg.URL),
	)

	logger.Info(fmt.Sprintf("LOG file: %s", cfg.LogFilePath))

	// Инициализируем иконки один раз при старте
	if err := speaker.InitIcons(cfg); err != nil {
		logger.Panic("Ошибка инициализации иконок", zap.Error(err))
	}

	listener := pl.New(cfg.URL)

	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Перехватываем сигналы завершения от ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Канал для сообщений (буфер 10, чтобы слушатель не блокировался)
	msgChan := make(chan pl.NtfyMessageType, 10)

	// WaitGroup для корректного завершения горутин
	var wg sync.WaitGroup
	wg.Add(1)

	// Запускаем слушатель в отдельной горутине с циклом переподключения
	go func() {
		defer wg.Done()
		logger.Info("Запуск слушателя ntfy...")

		for {
			err := listener.Start(ctx, msgChan)

			// Если контекст отменен (завершаем программу), выходим
			if ctx.Err() != nil {
				logger.Info("Слушатель остановлен по команде завершения")
				return
			}

			// Логируем ошибку и пытаемся переподключиться
			if err != nil {
				logger.Error("Соединение разорвано, попытка переподключения...", zap.Error(err))
			} else {
				logger.Warn("Соединение закрыто сервером, переподключение...")
			}

			// Ждем 5 секунд перед повторной попыткой
			select {
			case <-time.After(5 * time.Second):
				continue
			case <-ctx.Done():
				return
			}
		}
	}()

	// Главный цикл обработки сообщений
	for {
		select {
		case msg := <-msgChan:
			err := speaker.Speak(msg, cfg)
			if err != nil {
				logger.Error("ошибка отправки toast-уведомления", zap.Error(err))
			}

		case <-sigChan:
			logger.Info("Получен сигнал завершения, инициируем graceful shutdown...")
			cancel()  // Отменяем контекст
			wg.Wait() // Ждем завершения всех горутин
			logger.Info("Приложение полностью остановлено")
			return
		}
	}
}
