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
	// Configurar nivel de log
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	// Intentar usar directorio de logs (puede ser volumen montado o directorio local)
	logsDir := os.Getenv("LOGS_DIR")
	if logsDir == "" {
		logsDir = "logs"
	}

	var logFile *os.File
	var multiWriter io.Writer = os.Stdout

	// Intentar crear directorio y archivo de log
	// Si falla, solo usaremos stdout (no hacemos panic)
	if err := os.MkdirAll(logsDir, 0755); err == nil {
		logPath := filepath.Join(logsDir, "app.log")
		file, err := os.OpenFile(
			logPath,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
		if err == nil {
			logFile = file
			multiWriter = io.MultiWriter(os.Stdout, logFile)
		} else {
			// Si no puede escribir al archivo, solo usar stdout
			fmt.Fprintf(os.Stderr, "Warning: No se pudo abrir archivo de log (%s), usando solo stdout: %v\n", logPath, err)
		}
	} else {
		// Si no puede crear el directorio, solo usar stdout
		fmt.Fprintf(os.Stderr, "Warning: No se pudo crear directorio logs/ (%s), usando solo stdout: %v\n", logsDir, err)
	}

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
