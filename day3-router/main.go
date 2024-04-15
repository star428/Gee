package main

import "Gee/day3-router/gee"

func main() {
	r := gee.New()

	r.GET("/", func(c *gee.Context) {
		c.HTML(200, "<h1>Hello Gee</h1>")
	})

	// Test it with:
	// curl http://localhost:9999/hello?name=geektutu
	r.GET("/hello", func(c *gee.Context) {
		c.String(200, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	// Test it with:
	// curl http://localhost:9999/hello/geektutu
	r.GET("/hello/:name", func(c *gee.Context) {
		c.String(200, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})

	// Test it with:
	// curl http://localhost:9999/assets/file1.txt
	r.GET("/assets/*filepath", func(c *gee.Context) {
		c.JSON(200, gee.H{"filepath": c.Param("filepath")})
	})

	r.Run(":9999")
}
