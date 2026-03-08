package dto

// Типы соответствуют docs/openapi.json #/components/schemas/MessageResponse.
// Ошибки 4xx/5xx — ServiceErrorResponse (apierrors), не MessageResponse.

// MessageStatus — статус в MessageResponse.
type MessageStatus string

const (
	MessageStatusError   MessageStatus = "error"
	MessageStatusSuccess MessageStatus = "success"
	MessageStatusInfo    MessageStatus = "info"
)

// MessageResponse — ответ с сообщением и опциональными деталями (OAS MessageResponse).
type MessageResponse struct {
	Message string          `json:"message"`
	Status  MessageStatus   `json:"status"`
	Details []MessageDetail `json:"details,omitempty"`
}

// MessageDetail — элемент details в MessageResponse.
type MessageDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}
