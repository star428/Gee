package gee

import (
	"log"
	"time"
)

// Logger is a middleware that logs the server requests
func Logger() HandlerFunc {
	return func(c *Context) {
		// Start timer
		t := time.Now()
		c.Next()
		// Calculate resolution time
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
