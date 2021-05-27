package api

// NewMessage creates message in application message format
func NewMessage(message string, data interface{}) map[string]interface{} {
	// also can create separate type for that
	return map[string]interface{}{
		"message": message,
		"data":    data,
	}
}
