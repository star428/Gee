package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// H is for json data
type H map[string]interface{}

// Context is the context of one http request
type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	// response info
	StatusCode int
}

// NewContext is the constructor of Context
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    r,
		Path:   r.URL.Path,
		Method: r.Method,
	}
}

// PostForm is a helper function that parse form data
func (c *Context) PostForm(key string) string {
	// example: curl http://localhost:9999/form  -X POST -d 'username=geektutu&password=1234'
	return c.Req.FormValue(key)
}

// Query is a helper function that parse query data
func (c *Context) Query(key string) string {
	// example: curl http://localhost:9999/?username=geektutu&password=1234
	return c.Req.URL.Query().Get(key)
}

// Status sets the status code for the response
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// SetHeader sets the header for the response
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// String sets the string data for the response
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)

	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON sets the json data for the response
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)

	// obj is the gee.H type
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// Data sets the data for the response
func (c *Context) Data(code int, data []byte) {
	c.Status(code)

	c.Writer.Write(data)
}

// HTML sets the html data for the response
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)

	c.Writer.Write([]byte(html))
}
