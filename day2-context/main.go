package main

import "Gee/day2-context/gee"

func main() {
	r := gee.New()

	// curl "http://localhost:9999/"
	r.GET("/", func(c *gee.Context) {
		c.HTML(200, "<h1>Hello Gee</h1>")
	})

	// curl "http://localhost:9999/hello?name=geektutu"
	r.GET("/hello", func(c *gee.Context) {
		// expect /hello?name=geektutu
		c.String(200, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	// curl "http://localhost:9999/login" -X POST -d 'username=geektut&password=1234'
	r.POST("/login", func(c *gee.Context) {
		c.JSON(200, gee.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})

	r.Run(":9999")
}
