package jin

import (
	"log"
	"net/http"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

type HandlerChain []HandlerFunc

type RouterGroup struct {
	prefix      string
	middlewares HandlerChain // support middleware
	parent      *RouterGroup // support nesting
	engine      *Engine      // all groups share a Engine instance
}

// Engine implement the interface of ServeHTTP
type Engine struct {
	*RouterGroup

	router *router
	groups []*RouterGroup
}

// New is the constructor of gee.Engine
func New() *Engine {
	engine := &Engine{
		router: newRouter(),
		groups: make([]*RouterGroup, 0),
	}

	engine.RouterGroup = &RouterGroup{engine: engine}
	return engine
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

func (group *RouterGroup) Use(handlers ...HandlerFunc) {
	group.middlewares = append(group.middlewares, handlers...)
}

func (group *RouterGroup) Group(prefix string, handlers ...HandlerFunc) *RouterGroup {
	newGroup := &RouterGroup{
		prefix:      group.prefix + prefix,
		parent:      group,
		engine:      group.engine,
		middlewares: handlers,
	}

	group.engine.groups = append(group.engine.groups, newGroup)
	return newGroup
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

func (group *RouterGroup) getMiddleware() HandlerChain {
	if group.parent == nil {
		return make(HandlerChain, 0)
	}

	handlers := group.parent.getMiddleware()
	handlers = append(handlers, group.middlewares...)
	return handlers
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	handlers := group.getMiddleware()
	handlers = append(handlers, handler)
	group.engine.router.addRoute(method, pattern, handlers)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)
}
