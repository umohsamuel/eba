package docs

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files/v2"
)

const maxSpecBytes = 10 << 20

const darkCSS = `
html { box-sizing: border-box; overflow-y: scroll; }
*, *:before, *:after { box-sizing: inherit; }
body { margin: 0; background: #fafafa; }

html { filter: invert(90%) hue-rotate(180deg); }
img, .microlight, .highlight-code pre { filter: invert(100%) hue-rotate(180deg); }
`

const initializerJS = `window.onload = function () {
  window.ui = SwaggerUIBundle({
    urls: %s,
    dom_id: "#swagger-ui",
    deepLinking: true,
    persistAuthorization: true,
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    plugins: [SwaggerUIBundle.plugins.DownloadUrl],
    layout: "StandaloneLayout",
  });
};
`

type specLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Handler struct {
	specs       map[string]string
	initializer []byte
	assets      http.FileSystem
	client      *http.Client
}

func New(specs map[string]string) (*Handler, error) {
	names := make([]string, 0, len(specs))
	for name := range specs {
		names = append(names, name)
	}
	slices.Sort(names)

	links := make([]specLink, 0, len(names))
	for _, name := range names {
		links = append(links, specLink{Name: name, URL: "/docs/specs/" + name})
	}

	urls, err := json.Marshal(links)
	if err != nil {
		return nil, fmt.Errorf("encode spec urls: %w", err)
	}

	return &Handler{
		specs:       specs,
		initializer: fmt.Appendf(nil, initializerJS, urls),
		assets:      http.FS(swaggerFiles.FS),
		client:      &http.Client{Timeout: 5 * time.Second},
	}, nil
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	if len(h.specs) == 0 {
		return
	}

	r.GET("/docs/*any", h.serve)
}

func (h *Handler) serve(c *gin.Context) {
	path := c.Param("any")

	switch {
	case path == "/":
		c.FileFromFS("/", h.assets)
	case path == "/index.css":
		c.Data(http.StatusOK, "text/css; charset=utf-8", []byte(darkCSS))
	case path == "/swagger-initializer.js":
		c.Data(http.StatusOK, "application/javascript; charset=utf-8", h.initializer)
	case strings.HasPrefix(path, "/specs/"):
		h.proxySpec(c, strings.TrimPrefix(path, "/specs/"))
	default:
		c.FileFromFS(path, h.assets)
	}
}

func (h *Handler) proxySpec(c *gin.Context, name string) {
	url, ok := h.specs[name]
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
	if err != nil {
		slog.Error("build spec request", "spec", name, "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	resp, err := h.client.Do(req)
	if err != nil {
		slog.Warn("fetch spec", "spec", name, "url", url, "error", err)
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("fetch spec", "spec", name, "url", url, "status", resp.StatusCode)
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}

	c.Header("Cache-Control", "no-store")
	c.DataFromReader(http.StatusOK, -1, "application/json", io.LimitReader(resp.Body, maxSpecBytes), nil)
}
