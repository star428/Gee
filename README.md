# 从0实现go web框架（仿照Gin）

## day0-基础预备知识

实现web应用总是使用框架而不是使用标准库，我们需要知道框架为我们解决了怎样的问题。

使用标准库来处理一个请求的代码如下：

```go
func main() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/count", counter)
    log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
}
```

`net/http`提供了基础的web功能，比如监听端口、映射静态路由、解析HTTP报文。但是某些简单的需求并不支持，需要我们手动实现。

* 动态路由：例如 `hello/:name`，`hello/*`这种规则。（:name表明可以传递这个name字段）

```go
func main() {
    http.HandleFunc("/hello/", helloHandler)

    log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    name := strings.TrimPrefix(r.URL.Path, "/hello/")
    fmt.Fprintf(w, "Hello, %s!", name)
}
```

* 模板：没有统一简化的HTML机制。

离开框架使用基础库的时候，需要频繁手工处理的地方就是框架的价值所在。

## day1-HTTP基础

* 简单介绍 `net/http`库以及 `http.Handler`接口
* 搭建 `Gee`框架雏形

http.Handler接口如下：

```go
package http

type Handler interface {
    ServeHTTP(w ResponseWriter, r *Request)
}

func ListenAndServe(address string, h Handler) error
```

`ListenAndServe`第二个参数是一个接口，实现ServeHTTP这个函数即可，如果使用默认值nil的话，就交给HTTP包自己去管理路由，但是如果传入一个实现了 `ServeHTTP`函数的实例，所有的HTTP请求都会走 `ServeHTTP`函数。

那么我们目前的目标就是搞一个实例同时实现 `Handler`接口即可

```
base3/
    gee/
    |--gee.go
main.go
```

1. 定义了类型 `HandlerFunc`（func类型），提供给框架用户用来定义路由映射的方法。
2. 在struct `Engine`中添加一张路由表，`key`由请求方法和动态路由组成，例如：`GET-/`，`GET-/hello`等，中间用 `-`来风格，`value`就是用户映射的对应的处理方法。
3. 当用户调用 `(*Engine).GET()`方法时，会将路由和处理方法注册到映射表 `router`中，`(*Engine).Run()`方法其实时对 `ListenAndServe`的封装
4. `Engine`实现的ServeHTTP方法是，解析请求的路径，查找相关的路由映射表，如果查到，就执行相关的注册方法。如果查不到则是404 NOT FOUND。
