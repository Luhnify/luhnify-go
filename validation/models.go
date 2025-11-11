package validation

// Estructura para la solicitud
type ValidationPayload struct {
	CountryCode    string `json:"country_code"`
	DocumentType   string `json:"document_type"`
	DocumentNumber string `json:"document_number"`
}

// Estructura para la respuesta de éxito
type ValidationResult struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// Estructura para errores detallados de la API (ej: 422)
type APIErrorDetail struct {
	Valid          bool   `json:"valid"`
	Message        string `json:"message"`
	Errors         string `json:"errors"`
	ExpectedFormat string `json:"expected_format"`
}

// Estructura genérica para mensajes de error (ej: 403, 400)
type APIMessage struct {
	Message string `json:"message"`
}