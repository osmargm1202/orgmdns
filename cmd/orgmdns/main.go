package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/osmargm1202/orgmdns/internal/app"
	"github.com/osmargm1202/orgmdns/internal/config"
	"github.com/osmargm1202/orgmdns/internal/logger"
)

func main() {
	debugFlag := flag.Bool("debug", false, "Activa logs de depuración")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Si se pasa --debug, prevalece sobre DEBUG env
	if *debugFlag {
		cfg.Debug = true
	}

	log := logger.Init(cfg.Debug)
	defer log.Close()

	log.Info("Iniciando orgmdns...")
	log.Debug(fmt.Sprintf("Configuración cargada: ZONE_ID=%s, SLEEP_TIME=%d minutos", cfg.ZoneID, cfg.SleepTime))

	// Manejo de señales para shutdown graceful
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	runner := app.NewRunner(cfg, log)

	// Ejecutar en goroutine para poder recibir señales
	go func() {
		if err := runner.Run(); err != nil {
			log.Error(fmt.Sprintf("Error en runner: %v", err))
			os.Exit(1)
		}
	}()

	// Esperar señal de terminación
	<-sigChan
	log.Info("Recibida señal de terminación, cerrando...")
}
