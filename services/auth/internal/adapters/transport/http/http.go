package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/umohsamuel/eba/services/auth/docs"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/transport/http/handlers"
	"github.com/umohsamuel/eba/services/auth/internal/adapters/transport/http/middlewares"
	"github.com/umohsamuel/eba/services/auth/internal/services"
)

//	@title			Auth Service
//	@version		1.0
//	@description	This is the auth service for EBA.
//	@BasePath		/api/v1
//	@schemes		http

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter "Bearer" followed by a space and the access token.

type CookieConfig = handlers.CookieConfig

type MigrationConfig = handlers.MigrationConfig

type HttpServer struct {
	router       *gin.Engine
	services     *services.Services
	cookieCfg    CookieConfig
	migrationCfg MigrationConfig
}

func NewHttpServer(svc *services.Services, cookieCfg CookieConfig, migrationCfg MigrationConfig) *HttpServer {
	r := gin.Default()
	s := &HttpServer{router: r, services: svc, cookieCfg: cookieCfg, migrationCfg: migrationCfg}

	s.setupRoutes()
	return s
}

const swaggerDarkCSS = `
html { box-sizing: border-box; overflow-y: scroll; }
*, *:before, *:after { box-sizing: inherit; }
body { margin: 0; background: #fafafa; }

html { filter: invert(90%) hue-rotate(180deg); }
img, .microlight, .highlight-code pre { filter: invert(100%) hue-rotate(180deg); }
`

func swaggerDarkTheme(c *gin.Context) {
	if c.Param("any") == "/index.css" {
		c.Data(http.StatusOK, "text/css; charset=utf-8", []byte(swaggerDarkCSS))
		c.Abort()
	}
}

func (s *HttpServer) setupRoutes() {
	s.router.GET("/docs/*any", swaggerDarkTheme, ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := s.router.Group("/api/v1")

	authMW := middlewares.NewAuth(s.services.Auth)
	protected := v1.Group("", authMW.Handle)
	_ = protected

	authHandler := handlers.NewAuthHandler(s.services, s.cookieCfg)
	authHandler.RegisterRoutes(v1)
	s.router.GET("/.well-known/jwks.json", authHandler.JWKS)

	migrationHandler := handlers.NewMigrationHandler(s.migrationCfg)
	migrationHandler.RegisterRoutes(v1)
}

func (s *HttpServer) Handler() http.Handler {
	return s.router
}
