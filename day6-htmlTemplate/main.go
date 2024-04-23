package main

import (
	"Gee/day6-htmlTemplate/gee"
	"fmt"
	"html/template"
	"time"
)

type student struct {
	Name string
	Age  int
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := gee.New()
	r.Use(gee.Logger())
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "/home/wangye/golearn/Gee/day6-htmlTemplate/static")

	r.GET("/", func(c *gee.Context) {
		c.HTML(200, "css.html", nil)
		// c.HTMLNormal(200, "<h1>Hello Gee 123</h1>")
	})

	stu1 := &student{Name: "Geektutu", Age: 20}
	stu2 := &student{Name: "Jack", Age: 22}

	r.GET("/students", func(c *gee.Context) {
		c.HTML(200, "arr.html", gee.H{
			"title":  "gee",
			"stuArr": [2]*student{stu1, stu2},
		})
	})

	r.GET("/date", func(c *gee.Context) {
		c.HTML(200, "custom_func.html", gee.H{
			"title": "gee",
			"now":   time.Now(),
		})
	})
	r.Run(":9999")
}
