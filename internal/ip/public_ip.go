package ip

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pion/stun"
)

// GetPublicIP obtiene la IP pública usando STUN como método principal
// y HTTP como fallback si STUN falla
func GetPublicIP() (string, error) {
	// Intentar primero con STUN
	ip, err := getPublicIPSTUN()
	if err == nil {
		return ip, nil
	}

	// Fallback a HTTP
	return getPublicIPHTTP()
}

// getPublicIPSTUN obtiene la IP pública usando STUN
func getPublicIPSTUN() (string, error) {
	c, err := stun.Dial("udp", "stun.l.google.com:19302")
	if err != nil {
		return "", fmt.Errorf("error conectando a STUN: %w", err)
	}
	defer c.Close()

	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	var xorAddr stun.XORMappedAddress
	done := make(chan error, 1)

	c.Do(message, func(res stun.Event) {
		if res.Error != nil {
			done <- res.Error
			return
		}
		if err := res.Message.Parse(&xorAddr); err != nil {
			done <- err
			return
		}
		done <- nil
	})

	// Timeout de 5 segundos
	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("error en respuesta STUN: %w", err)
		}
		return xorAddr.IP.String(), nil
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("timeout esperando respuesta STUN")
	}
}

// getPublicIPHTTP obtiene la IP pública usando un servicio HTTP como fallback
func getPublicIPHTTP() (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Intentar con varios servicios
	services := []string{
		"https://api.ipify.org?format=text",
		"https://icanhazip.com",
		"https://ifconfig.me/ip",
	}

	for _, url := range services {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		scanner := bufio.NewScanner(resp.Body)
		if scanner.Scan() {
			ipStr := strings.TrimSpace(scanner.Text())
			// Validar que sea una IP válida
			if net.ParseIP(ipStr) != nil {
				return ipStr, nil
			}
		}
	}

	return "", fmt.Errorf("no se pudo obtener IP pública desde ningún servicio HTTP")
}
