package logger

import (
	_ "embed"
)

func WithService(serviceName string) Field {
	return WithString("service", serviceName)
}

func WithStatusCode(statusCode int) Field {
	return WithInt("status_code", statusCode)
}
