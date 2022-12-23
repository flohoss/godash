package main

func (g *goDash) setupRouter() {
	g.router.Use(g.authMiddleware())
	g.router.GET("/", g.homePage)
	g.router.GET("/ws", g.ws)
	g.router.GET("/robots.txt", robotsHandler)

	auth := g.router.Group("/auth")
	auth.GET("/login", g.loginPage)
	auth.POST("/login", g.loginHandler)
	auth.POST("/logout", g.logoutHandler)

	static := g.router.Group("/static", longCacheLifetime)
	static.Static("/", "static")

	storage := g.router.Group("/storage", longCacheLifetime)
	storage.Static("/icons", "storage/icons")

	g.router.RouteNotFound("/*", redirectHome)
}
