package ip

import (
	"net/http"
	"time"
)

// CheckInternetConnection verifica si hay conexión a internet
// Similar a la función check_internet() de Python
func CheckInternetConnection() bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Intentar hacer una petición a Google (como en Python)
	_, err := client.Get("https://www.google.com")
	return err == nil
}
