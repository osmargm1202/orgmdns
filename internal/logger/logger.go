package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"log/slog"
)

type Logger struct {
	*slog.Logger
	file *os.File
}

func Init(debug bool) *Logger {
	// Asegurar que existe el directorio logs/
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		panic(fmt.Sprintf("No se pudo crear directorio logs/: %v", err))
	}

	// Abrir archivo de log
	logFile, err := os.OpenFile(
		filepath.Join(logsDir, "app.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		panic(fmt.Sprintf("No se pudo abrir archivo de log: %v", err))
	}

	// Configurar nivel de log
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	// Crear writers múltiples (consola + archivo)
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Crear handler de texto con colores opcionales
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Formatear tiempo
			if a.Key == slog.TimeKey {
				return slog.String("time", a.Value.Time().Format(time.RFC3339))
			}
			return a
		},
	}

	handler := slog.NewTextHandler(multiWriter, opts)
	logger := slog.New(handler)

	return &Logger{
		Logger: logger,
		file:   logFile,
	}
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
