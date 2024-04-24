package main

import (
	"Gee/day7-panicRecover/gee"
)

func main() {
	r := gee.Defualt()
	r.GET("/", func(c *gee.Context) {
		c.String(200, "Hello Geektutu\n")
	})

	// index out of range for testing Recovery()
	r.GET("/panic", func(c *gee.Context) {
		names := []string{"geektutu"}
		c.String(200, names[100])
	})

	r.Run(":9999")
}
