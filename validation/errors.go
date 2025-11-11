package validation

import "fmt"

// BelloError envuelve el error de la API con el código de estado HTTP.
type BelloError struct {
	StatusCode int
	Message    string
}

// Implementación de la interfaz 'error' de Go
func (e *BelloError) Error() string {
	return fmt.Sprintf("API returned status %d: %s", e.StatusCode, e.Message)
}