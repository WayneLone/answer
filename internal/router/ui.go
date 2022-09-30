package router

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/segmentfault/answer/ui"
	"github.com/segmentfault/pacman/log"
)

const UIIndexFilePath = "build/index.html"
const UIRootFilePath = "build"
const UIStaticPath = "build/static"

// UIRouter is an interface that provides ui static file routers
type UIRouter struct {
}

// NewUIRouter creates a new UIRouter instance with the embed resources
func NewUIRouter() *UIRouter {
	return &UIRouter{}
}

// _resource is an interface that provides static file, it's a private interface
type _resource struct {
	fs embed.FS
}

// Open to implement the interface by http.FS required
func (r *_resource) Open(name string) (fs.File, error) {
	name = fmt.Sprintf(UIStaticPath+"/%s", name)
	log.Debugf("open static path %s", name)
	return r.fs.Open(name)
}

// Register a new static resource which generated by ui directory
func (a *UIRouter) Register(r *gin.Engine) {
	staticPath := os.Getenv("ANSWER_STATIC_PATH")

	// if ANSWER_STATIC_PATH is set and not empty, ignore embed resource
	if staticPath != "" {
		info, err := os.Stat(staticPath)

		if err != nil || !info.IsDir() {
			log.Error(err)
		} else {
			log.Debugf("registering static path %s", staticPath)

			r.LoadHTMLGlob(staticPath + "/*.html")
			r.Static("/static", staticPath+"/static")
			r.NoRoute(func(c *gin.Context) {
				c.HTML(http.StatusOK, "index.html", gin.H{})
			})

			// return immediately if the static path is set
			return
		}
	}

	// handle the static file by default ui static files
	r.StaticFS("/static", http.FS(&_resource{
		fs: ui.Build,
	}))

	// specify the not router for default routes and redirect
	r.NoRoute(func(c *gin.Context) {
		name := c.Request.URL.Path
		filePath := ""
		var file []byte
		var err error
		switch name {
		case "/favicon.ico":
			c.Header("content-type", "image/vnd.microsoft.icon")
			filePath = UIRootFilePath + name
		case "/logo192.png":
			filePath = UIRootFilePath + name
		case "/logo512.png":
			filePath = UIRootFilePath + name
		default:
			filePath = UIIndexFilePath
			c.Header("content-type", "text/html;charset=utf-8")
		}
		file, err = ui.Build.ReadFile(filePath)
		if err != nil {
			log.Error(err)
			c.Status(http.StatusNotFound)
			return
		}
		c.String(http.StatusOK, string(file))
	})
}