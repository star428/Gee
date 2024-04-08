package gee

import "log"

// HandlerFunc defines the request handler used by gee
type router struct {
	// key: method + pattern
	handlers map[string]HandlerFunc
}

// newRouter is the constructor of router
func newRouter() *router {
	return &router{
		handlers: make(map[string]HandlerFunc),
	}
}

// addRoute adds a route to the router
func (r *router) addRoute(method, pattern string, handler HandlerFunc) {
	log.Printf("Route %4s - %s", method, pattern) // %4s means right-aligned string with length of 4
	key := method + "-" + pattern
	r.handlers[key] = handler
}

// handle is to handle the request
func (r *router) handle(c *Context) {
	key := c.Method + "-" + c.Path
	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		c.String(404, "404 NOT FOUND: %s\n", c.Path)
	}
}
