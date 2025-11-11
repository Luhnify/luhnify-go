// 📄 validation/client.go
package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Interfaz HTTPClient para inyección de dependencias (necesaria para el mocking)
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ValidationClient contiene el cliente HTTP y las credenciales
type ValidationClient struct {
	apiKey    string
	baseURL   string
	httpClient HTTPClient
}

// NewValidationClient es el constructor
func NewValidationClient(apiKey, baseURL string, client HTTPClient) *ValidationClient {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &ValidationClient{
		apiKey:    apiKey,
		baseURL:   baseURL,
		httpClient: client,
	}
}

// ValidateDocument realiza la petición a la API
func (c *ValidationClient) ValidateDocument(payload ValidationPayload) (*ValidationResult, error) {
    
    // Serializar el payload a JSON
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("error al serializar payload: %w", err)
    }

	// Construir la petición
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/v1/validate-document", bytes.NewBuffer(payloadBytes))
    if err != nil {
        return nil, fmt.Errorf("error al crear la petición: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", c.apiKey)
    
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error al realizar la petición: %w", err)
	}
	defer resp.Body.Close()

    // --- Lógica de manejo de errores HTTP ---
    if resp.StatusCode != http.StatusOK {
        // CORRECCIÓN: Usamos io.ReadAll (la forma moderna)
        errorBody, _ := io.ReadAll(resp.Body) 
        
        // Asumimos que el cuerpo es un mensaje de error simple
        var apiMsg APIMessage
        if err := json.Unmarshal(errorBody, &apiMsg); err == nil && apiMsg.Message != "" {
            return nil, &BelloError{
                StatusCode: resp.StatusCode,
                Message:    apiMsg.Message,
            }
        }
        
        // Si no se pudo decodificar el JSON, usamos el cuerpo crudo.
        return nil, &BelloError{
            StatusCode: resp.StatusCode,
            Message:    string(errorBody), 
        }
    }

    // --- Lógica de éxito ---
    var result ValidationResult
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("error al decodificar resultado: %w", err)
    }
    
    return &result, nil
}