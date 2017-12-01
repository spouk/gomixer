//---------------------------------------------------------------------------
//  обеспечивает организацию субдоменов
//---------------------------------------------------------------------------

package gomixer

import (
	"fmt"
	"strings"
)

type Group struct {
	path string
	mux  *Gomixer
}

// Add a sub-group to this group
func (g *Group) NewGroup(path string) *Group {
	if len(path) < 1 {
		panic("Group path must not be empty")
	}

	checkPath(path)
	path = g.path + path
	//Don't want trailing slash as all sub-paths start with slash
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return &Group{path, g.mux}
}
//func (g *Group) MiddlewareAdd(prefix string, middleware GomixerMiddleware) {
//	if prefix == "" {
//		g.mux.middlewares[""] = append(g.mux.middlewares[g.path], middleware)
//	} else {
//		g.mux.middlewares[g.path + prefix] = append(g.mux.middlewares[g.path + prefix], middleware)
//	}
//}

// Path elements starting with : indicate a wildcard in the path. A wildcard will only match on a
// single path segment. That is, the pattern `/post/:postid` will match on `/post/1` or `/post/1/`,
// but not `/post/1/2`.
//
// A path element starting with * is a catch-all, whose value will be a string containing all text
// in the URL matched by the wildcards. For example, with a pattern of `/images/*path` and a
// requested URL `images/abc/def`, path would contain `abc/def`.
//
// # Routing Rule Priority
//
// The priority rules in the router are simple.
//
// 1. Static path segments take the highest priority. If a segment and its subtree are able to match the URL, that match is returned.
//
// 2. Wildcards take second priority. For a particular wildcard to match, that wildcard and its subtree must match the URL.
//
// 3. Finally, a catch-all rule will match when the earlier path segments have matched, and none of the static or wildcard conditions have matched. Catch-all rules must be at the end of a pattern.
//
// So with the following patterns, we'll see certain matches:
//	 router = httptreemux.New()
//	 router.GET("/:page", pageHandler)
//	 router.GET("/:year/:month/:post", postHandler)
//	 router.GET("/:year/:month", archiveHandler)
//	 router.GET("/images/*path", staticHandler)
//	 router.GET("/favicon.ico", staticHandler)
//
//	 /abc will match /:page
//	 /2014/05 will match /:year/:month
//	 /2014/05/really-great-blog-post will match /:year/:month/:post
//	 /images/CoolImage.gif will match /images/*path
//	 /images/2014/05/MayImage.jpg will also match /images/*path, with all the text after /images stored in the variable path.
//	 /favicon.ico will match /favicon.ico
//
// # Trailing Slashes
//
// The router has special handling for paths with trailing slashes. If a pattern is added to the
// router with a trailing slash, any matches on that pattern without a trailing slash will be
// redirected to the version with the slash. If a pattern does not have a trailing slash, matches on
// that pattern with a trailing slash will be redirected to the version without.
//
// The trailing slash flag is only stored once for a pattern. That is, if a pattern is added for a
// method with a trailing slash, all other methods for that pattern will also be considered to have a
// trailing slash, regardless of whether or not it is specified for those methods too.
//
// This behavior can be turned off by setting Gomixer.RedirectTrailingSlash to false. By
// default it is set to true. The specifics of the redirect depend on RedirectBehavior.
//
// One exception to this rule is catch-all patterns. By default, trailing slash redirection is
// disabled on catch-all patterns, since the structure of the entire URL and the desired patterns
// can not be predicted. If trailing slash removal is desired on catch-all patterns, set
// Gomixer.RemoveCatchAllTrailingSlash to true.
//
// 	router = httptreemux.New()
// 	router.GET("/about", pageHandler)
// 	router.GET("/posts/", postIndexHandler)
// 	router.POST("/posts", postFormHandler)
//
// 	GET /about will match normally.
// 	GET /about/ will redirect to /about.
// 	GET /posts will redirect to /posts/.
// 	GET /posts/ will match normally.
// 	POST /posts will redirect to /posts/, because the GET method used a trailing slash.
func (g *Group) Handle(method string, path string, handler HandlerFunc, static bool, staticprefix string, staticrealpath string ) {

	//проверка пути на пустотность
	checkPath(path)
	path = g.path + path
	//g.mux.Log.Printf("GROUP ADDING HANDLERS: %v\n", path)
	hnode := showmapnode{
		Handler:&handler,
		Method:method,
		Realpath:path,
		Prefix:g.path,
	}
	g.mux.handlermap = append(g.mux.handlermap, hnode)

	if len(path) == 0 {
		panic("Cannot map an empty path")
	}
	//добавление/удаление конечного слеша + формирование окончательного варианта пути
	addSlash := false
	if len(path) > 1 && path[len(path)-1] == '/' && g.mux.RedirectTrailingSlash {
		addSlash = true
		path = path[:len(path)-1]
	}

	//добавление ноды в дерево
	node := g.mux.root.addPath(path[1:], nil, false, static, staticprefix, staticrealpath)
	//сохранения префикса
	node.prefix = g.path

	if addSlash {
		node.addSlash = true
	}
	//set handler for node
	node.setHandler(method, handler, false)

	//set realpath
	if !strings.Contains(path, "*") {
		node.realpath = path
	}

	if g.mux.HeadCanUseGet && method == "GET" && node.leafHandler["HEAD"] == nil {
		node.setHandler("HEAD", handler, true)
	}
}
//
func (g *Group) MULTI(methods []string, path string, handler HandlerFunc) {
	for _, x := range methods {
		g.Handle(x, path, handler, false, "", "")
	}
}

// Syntactic sugar for Handle("GET", path, handler)
func (g *Group) GET(path string, handler HandlerFunc) {
	g.Handle("GET", path, handler, false, "", "")

}

// Syntactic sugar for Handle("POST", path, handler)
func (g *Group) POST(path string, handler HandlerFunc) {
	g.Handle("POST", path, handler, false, "", "")
}

// Syntactic sugar for Handle("PUT", path, handler)
func (g *Group) PUT(path string, handler HandlerFunc) {
	g.Handle("PUT", path, handler, false, "", "")
}

// Syntactic sugar for Handle("DELETE", path, handler)
func (g *Group) DELETE(path string, handler HandlerFunc) {
	g.Handle("DELETE", path, handler, false, "", "")
}

// Syntactic sugar for Handle("PATCH", path, handler)
func (g *Group) PATCH(path string, handler HandlerFunc) {
	g.Handle("PATCH", path, handler, false, "", "")
}

// Syntactic sugar for Handle("HEAD", path, handler)
func (g *Group) HEAD(path string, handler HandlerFunc) {
	g.Handle("HEAD", path, handler, false, "", "")
}

// Syntactic sugar for Handle("OPTIONS", path, handler)
func (g *Group) OPTIONS(path string, handler HandlerFunc) {
	g.Handle("OPTIONS", path, handler, false, "", "")
}

//add static path
func (g *Group) StaticResourse(prefix string, realpath string) {
	if strings.HasSuffix(prefix, "/") {
		prefix  += "*path"
	} else {
		prefix += "/*path"
	}
	g.Handle("GET", prefix, g.mux.staticHandler, true, prefix, realpath)
	//g.Handle("GET", prefix, g.mux.staticHandler, true, prefix, realpath)
}


func checkPath(path string) {
	// All non-empty paths must start with a slash
	if len(path) > 0 && path[0] != '/' {
		panic(fmt.Sprintf("Path %s must start with slash", path))
	}
}
