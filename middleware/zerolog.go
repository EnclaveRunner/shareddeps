package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ZerologMiddleware returns a gin middleware that logs HTTP requests using
// zerolog. It logs the HTTP method, path, client IP, status code, latency, and
// other request details.
func Zerolog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request after processing
		param := gin.LogFormatterParams{
			Request:    c.Request,
			TimeStamp:  time.Now(),
			Latency:    time.Since(start),
			ClientIP:   c.ClientIP(),
			Method:     c.Request.Method,
			StatusCode: c.Writer.Status(),
			BodySize:   c.Writer.Size(),
			Path:       path,
		}

		if raw != "" {
			path = path + "?" + raw
		}

		// Choose log level based on status code
		var logEvent *zerolog.Event
		switch {
		case param.StatusCode >= http.StatusBadRequest:
			logEvent = log.Warn()
		default:
			logEvent = log.Info()
		}

		logEvent.
			Str("method", param.Method).
			Str("path", path).
			Str("ip", param.ClientIP).
			Int("status", param.StatusCode).
			Dur("latency", param.Latency).
			Int("size", param.BodySize).
			Str("user_agent", c.Request.UserAgent()).
			Msg("HTTP Request")
	}
}
