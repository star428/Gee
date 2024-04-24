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

## Day4-分组控制Group

### 分组的意义

分组指的是路由的分组，没有路由分组，我们需要对每一个路由进行控制（也就是写的每一个GET,POST等）。一般分组控制使用前缀来进行区分，比如 `/POST`是一个分组，那么 `/POST/A`和 `/POST/B`是该分组下的子分组，作用于 `/POST`分组上的中间件（middleware），也会作用于子分组，子分组也可以有自己特有的中间件。

### 分组嵌套

一个Group对象应该具有这些属性：

1. 前缀（prefix）：比如 `/`，或者 `/api`
2. 父亲（parent）：支持分组嵌套需要知道父亲节点是谁
3. 该分组所含的中间件（middleware）：中间件应用在分组上
4. 访问router的能力

之前我们调用函数 `(*Engine).addRoute()`来映射路由规则和handler

```go
// addRoute adds a route to the router
func (e *Engine) addRoute(method, pattern string, handler HandlerFunc) {
	e.router.addRoute(method, pattern, handler)
}
```

所以我们需要赋予每个组有访问router的能力（以前都是交给engine来管理），所以可以给每个group设置一个指向engine的指针，这样就可以使用engine的全部其他方法

```go
RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // support middleware
	parent      *RouterGroup  // support nesting
	engine      *Engine       // all groups share a Engine instance
}
```

还可以进一步抽象，最初的engine也算一个group，就是最顶级的分组

```go
Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup // store all groups
}
```

接下来将路由有关的函数都交给 `RouterGroup`来实现而不是 `engine`来实现即可

包括三个函数：

* addRoute
* GET
* POST

全部由Engine转交到RouterGroup来实现

```go
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}
```

因为engine嵌套了 `*RouterGroup`，所以可以使用它的全部方法，使用的时候意味着编译器会生成新的方法，也就是接收者从 `*RouterGroup`变成 `*Engine`（参考gopl）

## Day5-中间件middleware

中间件（middleware）就是非业务的技术类组件。

web框架不可能去理解所有的业务，我们需要一个接口来允许用户去**自定义功能**，嵌入到框架中。

中间件需要考虑两个重要的点：

* 插入点：不可以太接近框架底层，否则逻辑会变得复杂；如果插入点离用户过近，用户直接定义一组函数在相关的handler中调用这个函数即可就没必要使用中间件。
* 中间件的输入：（我们会使用Context上下文作为输入），如果输入暴露的参数过少，用户发挥空间比较有限。

### 中间件设计

* 中间件的定义与路由映射的Handler（也就是HandlerFunc）一致，处理的输入是**Context**对象
* 中间件的插入点是框架接收到request请求后（也就是[`(gee.Engine).ServeHTTP`](vscode-file://vscode-app/c:/Users/Administrator/AppData/Local/Programs/Microsoft%20VS%20Code/resources/app/out/vs/code/electron-sandbox/workbench/workbench.html "https://pkg.go.dev/Gee/day5-middleware/gee#Engine.ServeHTTP")），允许用户使用自己的中间件做一些额外的处理，比如记录日志等，以及对 `Context`进行二次加工。
* 另外调用 `(gee.Context).Next`函数，中间件的执行可以分为两个部分：response之前和response之后处理
* 同时支持多个中间件，依次进行调用。

中间件应用于 `RouterGroup`之上，不作用于每一个路由规则上的原因，是基于每一条路由做中间件还不如直接在Handler中调用某个函数来的直观，通用性太差，不适合定义为中间件。

我们之前的框架逻辑为：接收到请求后，匹配路由，该request的所有信息都会保存在 `Context`之中。所以我们需要接收到请求后，查找所有应用于该路由的中间件，保存在 `Context`之中，然后依次调用。

此时我们给 `Context`添加了2个参数，定义了 `Next`方法。

```go
type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	Params map[string]string
	// response info
	StatusCode int
	// middleware
	handlers []HandlerFunc
	index    int
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Path:   req.URL.Path,
		Method: req.Method,
		Req:    req,
		Writer: w,
		index:  -1,
	}
}

func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}
```

注意handlers和index字段，第一个存储middleware和handlerFunc，第二个代表执行到哪个middleware/HandlerFunc，Next的用法如下所示：

```go
func A(c *Context) {
    part1
    c.Next()
    part2
}
func B(c *Context) {
    part3
    c.Next()
    part4
}
```

此时我们应用了middleware A,B,和路由映射的handlerFunc，`c.handlers`里面的内容为 `[A, B, handler]`，此时调用的顺序为 `part1 -> part3 -> handler -> part2 -> part4`

### 代码实现

* 定义use函数，将中间件应用到某个Group。

**gee.go**

```go
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	engine.router.handle(c)
```

`engine.groups`记录了多少个组（根组也包含在内），然后通过判断组的前缀来判定（根组的前缀为""空字符串，所以都可以匹配）

* handle函数中，将路由匹配得到的Handler添加到 `c.handlers`列表中，执行 `c.Next()`（index初始值为-1，也就是从这块开始执行第一个index=0的项）

```go
func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)

	if n != nil {
		key := c.Method + "-" + n.pattern
		c.Params = params
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
```

## day6-HTML Template

### 服务端渲染

目前其实大多都是前后端分离，后端不传静态HTML页面（也就是渲染好的页面），而是提供RESTful接口，返回结构化的数据（JSON/XML）。

但是前后端分离的话，页面是在客户端渲染，比如浏览器，对爬虫不友好（爬不到页面数据）。所以短期内爬取服务端直接渲染好的HTML页面也很重要。

### 静态文件

网页三大支持JS,CSS,HTML，服务器渲染要支持JS,CSS等静态文件。

我们使用通配符 `*`来匹配多级子路径。比如路由规则 `/assets/*filepath`，我们可以匹配到 `/assets/js/geektutu.js`，然后可以获取filepath参数，为 `js/geektutu.js`。

获取了filepath后，可以得到相对地址，组合为绝对地址后访问服务器上的相关资源（假设放在 `/usr/web`下，我们的filepath就是在这个地址上的相对地址）

我们所做的工作就是解析请求的地址，映射到服务器上的真实地址，交给 `http.FileServer`即可

```go
// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		// 作为fileServer来处理来的HTTP请求，fileServer是一个interface里面只有一个ServeHTTP 			func
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// serve static files
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	group.GET(urlPattern, handler)
}
```

### HTML模板渲染

使用 `html/template`库，主要语句有两个：

```go
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}


func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
```

第一个是将目前所有的pattern下(比如 `./templates/*`)都加载进内存

第二个是根据模板name来选择哪一个模板执行渲染写入 `c.Writer`

最终使用代码如下：

```go
r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.GET("/students", func(c *gee.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", gee.H{
			"title":  "gee",
			"stuArr": [2]*student{stu1, stu2},
		})
	})

	r.GET("/date", func(c *gee.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
			"title": "gee",
			"now":   time.Now(),
		})
	})
```

* 第一个是测试css是否成功加载
* 第二个是实验模板的data是否写入
* 第三个同样也是实验data是否写入

## day7-错误恢复panic recover

整体思路为创建一个错误处理中间件(middleware)，错误发生时返回给用户 `Internal Server error`：500，同时在服务端日志中打印必要的错误信息，方便进行错误定位。

```go
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(500, "Internal Server Error") // 阻断器阻止后面的中间件执行
			}
		}()
		c.Next()
	}
}

```

可以看到这个中间件只有在错误出现的时候会调用defer来进行最后的处理，同时还会调用 `c.Fail`来阻断后续中间件的执行（但是理论上都会执行结束，因为会在最后用户的HandlerFunc上出错，前面所有的中间件都可以顺利执行）

其中 `trace()`函数可以获取堆栈信息，打印出错误信息
