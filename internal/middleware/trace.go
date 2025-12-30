package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebaespinosa/test_NF/internal/logging"
	"go.uber.org/zap"
)

// TraceMiddleware adds trace and request IDs to context for all requests
func TraceMiddleware(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID if not provided
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Get or generate trace ID
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Store in context for downstream handlers
		c.Set(logging.RequestIDKey, requestID)
		c.Set(logging.TraceIDKey, traceID)

		// Add to response headers
		c.Header("X-Request-ID", requestID)
		c.Header("X-Trace-ID", traceID)

		// Create request-scoped context with correlation IDs
		ctxWithValues := c.Request.Context()
		ctxWithValues = context.WithValue(ctxWithValues, logging.RequestIDKey, requestID)
		ctxWithValues = context.WithValue(ctxWithValues, logging.TraceIDKey, traceID)

		// Log request
		logger.WithContext(ctxWithValues).Info(
			"incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		c.Request = c.Request.WithContext(ctxWithValues)

		c.Next()

		// Log response
		logger.WithContext(ctxWithValues).Info(
			"request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
		)
	}
}
