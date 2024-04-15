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

## day2-上下文context

* 将 `路由（router）`独立出来，方便之后增强
* 设计 `上下文（context）`，封装request和response，提供对JSON、HTML等返回类型的支持

### 为什么要设计context？

1. 对于web应用而言，基础就是根据请求 `*http.Request`，构造响应的 `http.ResponseWriter`。但是这两个对象提供的接口粒度过细，要构造一个完整的响应需要考虑 `消息头(Header)`和 `消息体(Body)`。每个路由处理函数都得重复的写这些东西。现在看发送一个JSON response包的对比：

```go
obj = map[string]interface{}{
    "name": "geektutu",
    "password": "1234",
}
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
encoder := json.NewEncoder(w)
if err := encoder.Encode(obj); err != nil {
    http.Error(w, err.Error(), 500)
}
```

但是我们封装后的代码像这样：

```go
c.JSON(http.StatusOK, gee.H{
    "username": c.PostForm("username"),
    "password": c.PostForm("password"),
})
```

相当于对消息头的设置已经由内部函数完成，只需要提供你想要什么类型的消息头；同时消息体也使用函数来进行封装

2. 对于框架而言还需要实现其他额外功能。比如解析动态路由 `/hello:name`，解析出来的参数 `:name`的值存放在何处？框架支持的中间件产生的信息也无处存放。所以设计Context意味着将扩展性和复杂性留在内部，对外简化了相关接口。

### context具体实现

参照 `./gee/context.go`

* `map[string]interface{}`起了一个别名，只是方便后面构建JSON数据，JSON数据构建使用这个数据结构
* `Context`目前只包含 `*http.Request `, `http.ResponseWriter`。另外提供了对Method和Path常用属性的直接访问。
* 提供了访问Query和PostForm参数的方法。（对应GET url中的参数和POST body中含的内容）
* 提供了快速构造String/Data/JSON/HTML响应的方法。

### 路由（router）

其实唯一不同就是以前的HandlerFunc函数是：（这是函数类型的重命名）

```go
type HandlerFunc func(http.ResponseWriter, *http.Request)
```

现在是：

```go
type HandlerFunc func(*Context)
```

因为Context相当于对上面两个参数进行封装同时增强

### 框架入口

其实其他都没变，此时 `Engine`的目标还是实现 `ListenAndServe`中的第二个参数也就是 `Handler`接口中的 `ServeHTTP`函数，如下所示：

```go
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)
}
```

可以看到在调用router.handle之前，先构造了一个Context对象，然后再是之前相同的流程通过method（GET/POST）和pattern（/，/hello）来判断路由从而进行相关操作

### 测试

最后使用 `main.go`测试以下：

使用语句：

```shell
curl "http://localhost:9999/hello?name=geektutu" // GET
结果是返回GET中包含的信息

curl "http://localhost:9999/login" -X POST -d 'username=geektutu&password=1234' // POST
结果是以json格式返回POST Body中的信息
```

## day3-前缀树路由Router

* 使用Trie树（前缀树）实现动态路由解析
* 支持两种模式匹配（动态匹配），`:name`和 `*filepath`

### Trie树

之前我们实现router是这么实现的：

`./day2-context/gee/router.go`

```go
type router struct {
	// key: method + pattern
	handlers map[string]HandlerFunc
}
```

这种方式只可以存储静态路由，对于动态路由（一条路由规则匹配多种路由），比如：`/hello/:name`，可以匹配 `/hello/geektutu`,`/hello/jack`等。

我们使用前缀树来作为动态路由匹配的底层数据结构，url请求路径是由 /来分隔的，那么每一段都可以当作前缀树的节点，查询树（GET树和POST树）即可获得目前输入的url是否匹配某一段路由。

实现的动态路由具备下面两个功能：

* 参数匹配 `:`。例如 `/p/:lang/doc`，匹配 `/p/c/doc`和 `/p/go/doc`
* 通配符 `*`。例如 `/static/*filePath`，可以匹配 `/static/fav.ico`，也可以匹配 `/static/js/jQuery.js`。

### Trie树实现

首先定义每个node的数据结构：

```go
type node struct {
	pattern  string // 待匹配路由，例如 /p/:lang
	part     string // 路由中的一部分，例如 :lang
	children []*node // 子节点，例如 [doc, tutorial, intro]
	isWild   bool // 是否精确匹配，part 含有 : 或 * 时为true
}
```

可以看到在最终的节点上pattern才有值，若是匹配结束后的层高上没有相关的pattern匹配字段，说明匹配失败。

`iswild`代表这一层可以随意匹配，不管这一层上究竟上面的值是多少

路由实现了两个功能，就是注册与匹配。

* **注册**：开发服务时，注册路由规则映射handler。

* **访问**：匹配路由规则，查找到相关的handler。

所以Trie树要实现节点的插入与查询。

* **插入：**递归查找每一层的节点，如果没有匹配到当前part的节点，就新建一个同时加入children list中
* **查询：**退出规则为层高为parts的高度或者匹配到了*，则查询当前node的pattern是否为 `""`（**只有在插入的最后一个node才放匹配的pattern，中间节点是没有的**），然后输出相关的node（内含pattern匹配规则用来后面匹配处理的func）

### Router

实现注册与访问函数：

```go
type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}
// roots key eg, roots['GET'] roots['POST']
// handlers key eg, handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']
```

其中访问函数 `getRoute`中，解析了 `:`和 `*`两种匹配符的参数，返回一个map， 例如 `/p/go/doc`匹配到 `/p/:lang/doc`，解析结果为：`{lang: "go"}`，`/static/css/geektutu.css`匹配到 `/static/*filepath`，解析结果为 `{filepath: "css/geektutu.css"}`。

### Context和handle（router.go）的变化

context中增加对url解析出来 `:`和 `*`的内容方便后面做处理

handle中getRoute得到相关的node后，再根据handlers中的func进行接下来的处理和操作

```go
func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		r.handlers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
```
