package fasthttprouter

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"path"
	"regexp"
	"sync"
)

type RouterGroup struct {
	Router   *Router
	Handlers []fasthttp.RequestHandler
	basePath string
	mp       sync.Map
}

//	func (group *RouterGroupMap) Use(middleware ...fasthttp.RequestHandler) {
//		r, _ := group.mp.Load(group.basePath)
//		rs := r.([]fasthttp.RequestHandler)
//		rs = append(rs, middleware...)
//		group.mp.Store(group.basePath, rs)
//	}
func (group *RouterGroup) Use(middleware ...fasthttp.RequestHandler) IRoutes {
	//group.Handlers = append(group.Handlers, middleware...)
	if res, ok := group.mp.Load(group.basePath); ok {
		list := res.([]fasthttp.RequestHandler)
		res = append(list, middleware...)
		group.mp.Store(group.basePath, list)
	} else {
		group.mp.Store(group.basePath, middleware)
	}
	return group.returnObj()
}

func (group *RouterGroup) Group(relativePath string) *RouterGroup {
	return &RouterGroup{basePath: relativePath, Router: group.Router, mp: sync.Map{}}
}
func (group *RouterGroup) handle(httpMethod, relativePath string, handlers ...fasthttp.RequestHandler) {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	//group.Router.addRoute(httpMethod, absolutePath, handlers...)
	//group.Router.Handle(httpMethod, absolutePath, handlers)
	switch httpMethod {
	case http.MethodGet:
		group.Router.GET(absolutePath, handlers...)
		break
	case http.MethodPost:
		group.Router.POST(absolutePath, handlers...)
		break
	case http.MethodPut:
		group.Router.PUT(absolutePath, handlers...)
		break
	case http.MethodDelete:
		group.Router.DELETE(absolutePath, handlers...)
		break
	default:
		group.Router.GET(absolutePath, handlers...)
	}
}

func (group *RouterGroup) combineHandlers(handlers []fasthttp.RequestHandler) []fasthttp.RequestHandler {
	list := make([]fasthttp.RequestHandler, 0)
	if res, ok := group.mp.Load(group.basePath); ok {
		list = res.([]fasthttp.RequestHandler)
	}
	group.Handlers = list
	finalSize := len(list) + len(handlers)
	mergedHandlers := make([]fasthttp.RequestHandler, finalSize)
	copy(mergedHandlers, list)
	copy(mergedHandlers[len(list):], handlers)
	return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}
func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}
func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}
func (group *RouterGroup) returnObj() IRoutes {

	return group
}

type IRoutes interface {
	Handle(string, string, fasthttp.RequestHandler) IRoutes
	//Any(string, ...fasthttp.RequestCtx) IRoutes
	GET(string, fasthttp.RequestHandler) IRoutes
	POST(string, fasthttp.RequestHandler) IRoutes
	DELETE(string, fasthttp.RequestHandler) IRoutes
	//PATCH(string, ...fasthttp.RequestCtx) IRoutes
	//PUT(string, ...fasthttp.RequestCtx) IRoutes
	//OPTIONS(string, ...fasthttp.RequestCtx) IRoutes
	//HEAD(string, ...fasthttp.RequestCtx) IRoutes
	//Match([]string, string, ...fasthttp.RequestCtx) IRoutes
}

var (
	// regEnLetter matches english letters for http method name
	regEnLetter = regexp.MustCompile("^[A-Z]+$")

	// anyMethods for RouterGroup Any method
	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

func (group *RouterGroup) Handle(httpMethod, relativePath string, handlers fasthttp.RequestHandler) IRoutes {
	if matched := regEnLetter.MatchString(httpMethod); !matched {
		panic("http method " + httpMethod + " is not valid")
	}
	group.handle(httpMethod, relativePath, handlers)
	return group.returnObj()
}

// POST is a shortcut for router.Handle("POST", path, handlers).
func (group *RouterGroup) POST(relativePath string, handlers fasthttp.RequestHandler) IRoutes {
	group.handle(http.MethodPost, relativePath, handlers)
	return group.returnObj()
}

// GET is a shortcut for router.Handle("GET", path, handlers).
func (group *RouterGroup) GET(relativePath string, handlers fasthttp.RequestHandler) IRoutes {
	group.handle(http.MethodGet, relativePath, handlers)
	return group.returnObj()
}

// DELETE is a shortcut for router.Handle("DELETE", path, handlers).
func (group *RouterGroup) DELETE(relativePath string, handlers fasthttp.RequestHandler) IRoutes {
	group.handle(http.MethodDelete, relativePath, handlers)
	return group.returnObj()
}
