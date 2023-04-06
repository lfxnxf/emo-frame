package server

import (
	"context"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/lfxnxf/emo-frame/logging"
	"github.com/lfxnxf/emo-frame/zd_http/http_ctx"
	"github.com/lfxnxf/emo-frame/zd_http/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	DebugMode   = "debug"
	ReleaseMode = "release"
	TestMode    = "test"
)

type routes struct {
	gets  []*HttpRoute
	posts []*HttpRoute
	any   []*HttpRoute
}

type HttpServerConfig struct {
	ServiceName string `toml:"service_name"`
	Port        int64  `toml:"port"`
	Mode        string `toml:"mode"`
}

type HttpServer struct {
	routes
	cfg    HttpServerConfig
	engine *gin.Engine
	groups []*HttpGroup
	Ctx    *http_ctx.HttpContext
	server *http.Server
}

type HttpRoute struct {
	route   string
	handler http_ctx.HttpHandler
}

type HttpGroup struct {
	routes
	groups   []*HttpGroup
	ginGroup *gin.RouterGroup
}

func NewHttpServer(cfg HttpServerConfig) *HttpServer {
	// 设置运行模式
	gin.SetMode(cfg.Mode)

	s := &HttpServer{
		engine: gin.New(),
		cfg:    cfg,
	}

	pprof.Register(s.engine) // 性能

	// 初始化中间件
	s.initPublicMiddleware()
	return s
}

func (s *HttpServer) GET(route string, handler http_ctx.HttpHandler) {
	s.gets = append(s.gets, &HttpRoute{
		route:   route,
		handler: handler,
	})
}

func (s *HttpServer) POST(route string, handler http_ctx.HttpHandler) {
	s.posts = append(s.posts, &HttpRoute{
		route:   route,
		handler: handler,
	})
}

func (s *HttpServer) Any(route string, handler http_ctx.HttpHandler) {
	s.any = append(s.any, &HttpRoute{
		route:   route,
		handler: handler,
	})
}

func (s *HttpServer) Use(handlers ...http_ctx.HttpHandler) {
	var ginHandlers []gin.HandlerFunc
	for _, v := range handlers {
		fun := v
		ginHandlers = append(ginHandlers, func(c *gin.Context) {
			fun(&http_ctx.HttpContext{Context: c})
		})
	}
	s.engine.Use(ginHandlers...)
}

func (s *HttpServer) Group(route string, handlers ...http_ctx.HttpHandler) *HttpGroup {
	var ginGroups []gin.HandlerFunc
	for _, v := range handlers {
		ginGroups = append(ginGroups, func(c *gin.Context) {
			v(&http_ctx.HttpContext{Context: c})
		})
	}
	g := &HttpGroup{
		ginGroup: s.engine.Group(route, ginGroups...),
	}
	s.groups = append(s.groups, g)
	return g
}

func (s *HttpServer) registerGetRoute() {
	for _, v := range s.gets {
		fun := v
		s.engine.GET(fun.route, func(c *gin.Context) {
			fun.handler(&http_ctx.HttpContext{Context: c})
		})
	}
}

func (s *HttpServer) registerPostRoute() {
	for _, v := range s.posts {
		fun := v
		s.engine.POST(fun.route, func(c *gin.Context) {
			fun.handler(&http_ctx.HttpContext{Context: c})
		})
	}
}

func (s *HttpServer) registerAnyRoute() {
	for _, v := range s.any {
		fun := v
		s.engine.Any(fun.route, func(c *gin.Context) {
			fun.handler(&http_ctx.HttpContext{Context: c})
		})
	}
}

func (g *HttpGroup) GET(route string, handler http_ctx.HttpHandler) {
	g.gets = append(g.gets, &HttpRoute{
		route:   route,
		handler: handler,
	})
}

func (g *HttpGroup) POST(route string, handler http_ctx.HttpHandler) {
	g.posts = append(g.posts, &HttpRoute{
		route:   route,
		handler: handler,
	})
}

func (g *HttpGroup) Any(route string, handler http_ctx.HttpHandler) {
	g.posts = append(g.posts, &HttpRoute{
		route:   route,
		handler: handler,
	})
}

func (g *HttpGroup) registerGetRoute() {
	setGroupGetRoute(g)
}

func setGroupGetRoute(group *HttpGroup) {
	if group == nil {
		return
	}
	for _, v := range group.gets {
		fun := v
		group.ginGroup.GET(fun.route, func(c *gin.Context) {
			fun.handler(&http_ctx.HttpContext{Context: c})
		})
	}
	for _, g := range group.groups {
		setGroupGetRoute(g)
	}
}

func (g *HttpGroup) registerPostRoute() {
	setGroupPostRoute(g)
}

func setGroupPostRoute(group *HttpGroup) {
	if group == nil {
		return
	}
	for _, v := range group.posts {
		fun := v
		group.ginGroup.POST(fun.route, func(c *gin.Context) {
			fun.handler(&http_ctx.HttpContext{Context: c})
		})
	}
	for _, g := range group.groups {
		setGroupPostRoute(g)
	}
}

func (g *HttpGroup) registerAnyRoute() {
	setGroupAnyRoute(g)
}

func setGroupAnyRoute(group *HttpGroup) {
	if group == nil {
		return
	}
	for _, v := range group.any {
		fun := v
		group.ginGroup.Any(fun.route, func(c *gin.Context) {
			fun.handler(&http_ctx.HttpContext{Context: c})
		})
	}
	for _, g := range group.groups {
		setGroupAnyRoute(g)
	}
}

func (g *HttpGroup) Use(handlers ...http_ctx.HttpHandler) {
	var ginHandlers []gin.HandlerFunc
	for _, v := range handlers {
		fun := v
		ginHandlers = append(ginHandlers, func(c *gin.Context) {
			fun(&http_ctx.HttpContext{Context: c})
		})
	}
	g.ginGroup.Use(ginHandlers...)
}

func (g *HttpGroup) Group(route string, handlers ...http_ctx.HttpHandler) *HttpGroup {
	var ginGroups []gin.HandlerFunc
	for _, v := range handlers {
		ginGroups = append(ginGroups, func(c *gin.Context) {
			v(&http_ctx.HttpContext{Context: c})
		})
	}
	group := &HttpGroup{
		ginGroup: g.ginGroup.Group(route, ginGroups...),
	}
	g.groups = append(g.groups, group)
	return group
}

func (s *HttpServer) Start() error {
	// 初始化路由
	// post
	s.registerPostRoute()
	// get
	s.registerGetRoute()
	// any
	s.registerAnyRoute()
	// 分组
	for _, g := range s.groups {
		g.registerPostRoute()
		g.registerGetRoute()
		g.registerAnyRoute()
	}
	s.server = &http.Server{
		Addr:           fmt.Sprintf(":%d", s.cfg.Port),
		Handler:        s.engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.server.ListenAndServe()
	if err != nil {
		logging.Errorw("start http server failed %v", zap.Error(err))
	}
	return err
}

func (s *HttpServer) Shutdown() {
	err := s.server.Shutdown(context.Background())
	if err != nil {
		fmt.Println(err)
	}
}

func (s *HttpServer) initPublicMiddleware() {
	// 设置中间件
	s.Use(middleware.GetOpts()...)
}
