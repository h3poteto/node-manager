package context

import "context"

type contextKey string

const RequestIDKey contextKey = "requestID"
const ControllerKey contextKey = "controller"

func SetRequestID(parent context.Context, requestID string) context.Context {
	return context.WithValue(parent, RequestIDKey, requestID)
}

func SetController(parent context.Context, controller string) context.Context {
	return context.WithValue(parent, ControllerKey, controller)
}
