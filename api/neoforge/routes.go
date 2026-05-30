package neoforge

import (
	"net/http"
	"os"

	"shulker/util/maven"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	maven maven.Which
}

func RegisterRoutes(r *gin.Engine) {
	h := &Handler{maven: maven.Neoforge}
	g := r.Group("/neoforge")
	{
		g.GET("/mc", h.getMinecraftVersions)
		g.GET("/loader", h.getLoaderVersions)
		g.POST("/update", h.postUpdate)
		g.GET("/download", h.getDownload)
	}
}

func (h *Handler) getMinecraftVersions(c *gin.Context) {
	content, err := os.ReadFile("/srv/shulker/neoforge/mc.json")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", content)
}

func (h *Handler) getLoaderVersions(c *gin.Context) {
	content, err := os.ReadFile("/srv/shulker/neoforge/loader.json")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", content)
}

func (h *Handler) postUpdate(c *gin.Context) {
	if err := h.maven.Update(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) getDownload(c *gin.Context) {
	mc := c.Query("mc")
	loader := c.Query("loader")

	if mc == "" {
		c.String(http.StatusBadRequest, "missing 'mc' query parameter")
		return
	}
	// loader is only optional when mc is a sentinel
	if loader == "" && mc != "release" && mc != "latest" {
		c.String(http.StatusBadRequest, "missing 'loader' query parameter")
		return
	}

	server := maven.Server{
		Platform:  h.maven,
		Minecraft: mc,
		Loader:    loader,
	}

	url, err := server.ResolveMavenURL()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.Redirect(http.StatusFound, url)
}
