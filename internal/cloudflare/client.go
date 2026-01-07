package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	accountID string
	apiKey    string
	apiEmail  string // Para método legacy API Key + Email
	zoneID    string
	baseURL   string
	httpClient *http.Client
}

type DNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

type DNSRecordResponse struct {
	Result []DNSRecord `json:"result"`
	Success bool       `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

type DNSRecordUpdateRequest struct {
	Content string `json:"content"`
}

type DNSRecordUpdateResponse struct {
	Result  DNSRecord `json:"result"`
	Success bool      `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func NewClient(accountID, apiKey, zoneID, apiEmail string) *Client {
	return &Client{
		accountID: accountID,
		apiKey:    apiKey,
		apiEmail:  apiEmail,
		zoneID:    zoneID,
		baseURL:   "https://api.cloudflare.com/client/v4",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// setAuthHeaders configura los headers de autenticación según el método disponible
func (c *Client) setAuthHeaders(req *http.Request) {
	// Si hay API_EMAIL, usar método legacy (API Key + Email)
	if c.apiEmail != "" {
		req.Header.Set("X-Auth-Email", c.apiEmail)
		req.Header.Set("X-Auth-Key", c.apiKey)
	} else {
		// Usar API Token (método recomendado)
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

// ListDNSRecords obtiene todos los registros DNS tipo A de la zona
func (c *Client) ListDNSRecords() ([]DNSRecord, error) {
	url := fmt.Sprintf("%s/zones/%s/dns_records?type=A", c.baseURL, c.zoneID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error haciendo request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error de API: status %d, body: %s", resp.StatusCode, string(body))
	}

	var recordResp DNSRecordResponse
	if err := json.Unmarshal(body, &recordResp); err != nil {
		return nil, fmt.Errorf("error parseando respuesta: %w", err)
	}

	if !recordResp.Success {
		errMsg := "error desconocido"
		if len(recordResp.Errors) > 0 {
			errMsg = recordResp.Errors[0].Message
		}
		return nil, fmt.Errorf("API retornó error: %s", errMsg)
	}

	return recordResp.Result, nil
}

// GetDNSRecordByName obtiene el registro DNS A por nombre (filtra localmente como en Python)
func (c *Client) GetDNSRecordByName(name string) (*DNSRecord, error) {
	// Obtener todos los registros tipo A (como hace Python)
	records, err := c.ListDNSRecords()
	if err != nil {
		return nil, fmt.Errorf("error listando registros DNS: %w", err)
	}

	// Filtrar localmente por nombre (como hace Python)
	for _, record := range records {
		if record.Name == name {
			return &record, nil
		}
	}

	return nil, fmt.Errorf("no se encontró registro DNS con nombre %s", name)
}

// UpdateDNSRecordIP actualiza la IP de un registro DNS A
func (c *Client) UpdateDNSRecordIP(recordID, newIP string) error {
	url := fmt.Sprintf("%s/zones/%s/dns_records/%s", c.baseURL, c.zoneID, recordID)

	updateReq := DNSRecordUpdateRequest{
		Content: newIP,
	}

	jsonData, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("error serializando request: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creando request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error haciendo request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error de API: status %d, body: %s", resp.StatusCode, string(body))
	}

	var updateResp DNSRecordUpdateResponse
	if err := json.Unmarshal(body, &updateResp); err != nil {
		return fmt.Errorf("error parseando respuesta: %w", err)
	}

	if !updateResp.Success {
		errMsg := "error desconocido"
		if len(updateResp.Errors) > 0 {
			errMsg = updateResp.Errors[0].Message
		}
		return fmt.Errorf("API retornó error: %s", errMsg)
	}

	return nil
}
