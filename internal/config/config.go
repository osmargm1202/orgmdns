package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Cloudflare
	AccountID string
	APIKey    string
	ZoneID    string
	APIEmail  string // Opcional: para autenticación con API Key (método legacy)

	// Email
	Email         string
	EmailFrom     string
	EmailTo       string
	EmailPassword string
	SMTPHost      string
	SMTPPort      string

	// App
	SleepTime   int // minutos
	RecordNames []string
	Debug       bool
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Cloudflare
	cfg.AccountID = os.Getenv("ACCOUNT_ID")
	if cfg.AccountID == "" {
		return nil, fmt.Errorf("ACCOUNT_ID es requerido")
	}

	cfg.APIKey = os.Getenv("API_KEY")
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API_KEY es requerido")
	}

	cfg.ZoneID = os.Getenv("ZONE_ID")
	if cfg.ZoneID == "" {
		return nil, fmt.Errorf("ZONE_ID es requerido")
	}

	// API Email (opcional, solo para método legacy API Key)
	// Si no está configurado, intentar usar EMAIL como fallback (como en Python)
	cfg.APIEmail = os.Getenv("API_EMAIL")
	if cfg.APIEmail == "" {
		cfg.APIEmail = os.Getenv("EMAIL") // Fallback a EMAIL si API_EMAIL no está configurado
	}

	// Email
	cfg.Email = os.Getenv("EMAIL")
	cfg.EmailFrom = os.Getenv("EMAIL_FROM")
	if cfg.EmailFrom == "" {
		return nil, fmt.Errorf("EMAIL_FROM es requerido")
	}

	cfg.EmailTo = os.Getenv("EMAIL_TO")
	if cfg.EmailTo == "" {
		return nil, fmt.Errorf("EMAIL_TO es requerido")
	}

	cfg.EmailPassword = os.Getenv("EMAIL_PASSWORD")
	if cfg.EmailPassword == "" {
		return nil, fmt.Errorf("EMAIL_PASSWORD es requerido")
	}

	// SMTP configuración (con valores por defecto para Gmail)
	cfg.SMTPHost = os.Getenv("SMTP_HOST")
	if cfg.SMTPHost == "" {
		cfg.SMTPHost = "smtp.gmail.com" // default Gmail
	}

	cfg.SMTPPort = os.Getenv("SMTP_PORT")
	if cfg.SMTPPort == "" {
		cfg.SMTPPort = "587" // default STARTTLS
	}

	// App
	sleepTimeStr := os.Getenv("SLEEP_TIME")
	if sleepTimeStr == "" {
		cfg.SleepTime = 10 // default 10 minutos
	} else {
		sleepTime, err := strconv.Atoi(sleepTimeStr)
		if err != nil {
			return nil, fmt.Errorf("SLEEP_TIME debe ser un número entero: %w", err)
		}
		if sleepTime <= 0 {
			return nil, fmt.Errorf("SLEEP_TIME debe ser mayor que 0")
		}
		cfg.SleepTime = sleepTime
	}

	recordNamesStr := os.Getenv("RECORD_NAMES")
	if recordNamesStr == "" {
		return nil, fmt.Errorf("RECORD_NAMES es requerido")
	}

	// Parsear RECORD_NAMES separados por comas y hacer trim
	parts := strings.Split(recordNamesStr, ",")
	cfg.RecordNames = make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			cfg.RecordNames = append(cfg.RecordNames, trimmed)
		}
	}

	if len(cfg.RecordNames) == 0 {
		return nil, fmt.Errorf("RECORD_NAMES debe contener al menos un registro")
	}

	// Debug
	cfg.Debug = os.Getenv("DEBUG") == "true"

	return cfg, nil
}

// SleepDuration retorna el tiempo de espera como time.Duration
func (c *Config) SleepDuration() time.Duration {
	return time.Duration(c.SleepTime) * time.Minute
}
