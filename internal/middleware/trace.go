package middleware

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebaespinosa/test_NF/internal/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// TraceMiddleware adds trace and request IDs to context for all requests
// and creates OpenTelemetry spans for distributed tracing
func TraceMiddleware(logger *logging.Logger) gin.HandlerFunc {
	tracer := otel.Tracer("gin-server")

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

		// Create OpenTelemetry span for this request
		ctx, span := tracer.Start(
			c.Request.Context(),
			fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
		)
		defer span.End()

		// Set span attributes
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.route", c.Request.URL.Path),
			attribute.String("request_id", requestID),
			attribute.String("trace_id", traceID),
		)

		// Store in context for downstream handlers
		c.Set(logging.RequestIDKey, requestID)
		c.Set(logging.TraceIDKey, traceID)

		// Add to response headers
		c.Header("X-Request-ID", requestID)
		c.Header("X-Trace-ID", traceID)

		// Create request-scoped context with correlation IDs and span
		ctxWithValues := context.WithValue(ctx, logging.RequestIDKey, requestID)
		ctxWithValues = context.WithValue(ctxWithValues, logging.TraceIDKey, traceID)

		// Log request
		logger.WithContext(ctxWithValues).Info(
			"incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		c.Request = c.Request.WithContext(ctxWithValues)

		c.Next()

		// Set HTTP status on span
		statusCode := c.Writer.Status()
		span.SetAttributes(attribute.Int("http.status_code", statusCode))

		if statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// Log response
		logger.WithContext(ctxWithValues).Info(
			"request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", statusCode),
		)
	}
}
