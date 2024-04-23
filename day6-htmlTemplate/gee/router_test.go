package gee

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParsePattern(t *testing.T) {
	textExample := []string{
		"/p/:lang",
		"/p/*",
		"/p/*name/*",
	}

	result := [][]string{
		{"p", ":lang"},
		{"p", "*"},
		{"p", "*name"},
	}

	for i, pattern := range textExample {
		parts := parsePattern(pattern)

		ok := reflect.DeepEqual(parts, result[i])

		if !ok {
			t.Fatalf("expect %v, but got %v", result[i], parts)
		}
	}
}

func TestDoublePattern(t *testing.T) {
	r := newRouter()

	r.addRoute("GET", "/p/:lang", nil)
	r.addRoute("GET", "/p/a", nil)
}

func TestGetRoute(t *testing.T) {
	newTestRouter := func() *router {
		r := newRouter()

		r.addRoute("GET", "/", nil)
		r.addRoute("GET", "/hello/:name", nil)
		r.addRoute("GET", "/hello/b/c", nil)
		r.addRoute("GET", "/hi/:name", nil)
		r.addRoute("GET", "/assets/*filepath", nil)

		return r
	}

	r := newTestRouter()

	n, ps := r.getRoute("GET", "/hello/geektutu")

	if n == nil {
		t.Fatal("nil shouldn't be returned")
	}

	if n.pattern != "/hello/:name" {
		t.Fatal("should match /hello/:name")
	}

	if ps["name"] != "geektutu" {
		t.Fatal("name should be equal to 'geektutu'")
	}

	fmt.Printf("matched path: %s, params['name']: %s\n", n.pattern, ps["name"])
}
