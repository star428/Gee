package main

import (
	"Gee/day5-middleware/gee"
	"log"
	"time"
)

func onlyForV2() gee.HandlerFunc {
	return func(c *gee.Context) {
		// start time
		t := time.Now()

		// if a server error occurred
		c.Fail(500, "Internal Server Error") // 阻断器，在这里阻断了后续的middleware和handler的执行

		// after response
		// c.Next()

		// calculate resolution time
		log.Printf("[%d] %s in %v for group V2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}

}

func main() {
	r := gee.New()
	r.Use(gee.Logger())

	r.GET("/", func(c *gee.Context) {
		c.HTML(200, "<h1>Hello Gee</h1>")
	})

	v2 := r.Group("/v2")
	v2.Use(onlyForV2())
	{
		v2.GET("/hello/:name", func(c *gee.Context) {
			// expect /hello/wangye
			c.String(200, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
	}

	r.Run(":9999")
}
