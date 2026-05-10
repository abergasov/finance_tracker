package logger

import (
	_ "embed"

	"github.com/google/uuid"
)

func WithService(serviceName string) Field {
	return WithString("service", serviceName)
}

func WithStatusCode(statusCode int) Field {
	return WithInt("status_code", statusCode)
}

func WithUserID(userID uuid.UUID) Field {
	return WithString("user_id", userID.String())
}
