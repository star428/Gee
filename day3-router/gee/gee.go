package gee

import "net/http"

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
	router *router
}

// New is the constructor of gee.Engine
func New() *Engine {
	return &Engine{router: newRouter()}
}

// addRoute adds a route to the router
func (e *Engine) addRoute(method, pattern string, handler HandlerFunc) {
	e.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

// ServeHTTP defines the method to serve HTTP request
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	e.router.handle(c)
}
