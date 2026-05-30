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

// @Summary  List Minecraft versions for NeoForge
// @Tags     neoforge
// @Produce  json
// @Success  200
// @Failure  500
// @Router   /neoforge/mc [get]
func (h *Handler) getMinecraftVersions(c *gin.Context) {
	content, err := os.ReadFile("/srv/shulker/neoforge/mc.json")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", content)
}

// @Summary  List NeoForge loader versions
// @Tags     neoforge
// @Produce  json
// @Success  200
// @Failure  500
// @Router   /neoforge/loader [get]
func (h *Handler) getLoaderVersions(c *gin.Context) {
	content, err := os.ReadFile("/srv/shulker/neoforge/loader.json")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", content)
}

// @Summary  Refresh cached NeoForge version data from Maven
// @Tags     neoforge
// @Success  200
// @Failure  500
// @Router   /neoforge/update [post]
func (h *Handler) postUpdate(c *gin.Context) {
	if err := h.maven.Update(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

// @Summary   Download NeoForge server jar
// @Tags      neoforge
// @Param     mc      query  string  false  "Minecraft version (or 'latest'/'release')"
// @Param     loader  query  string  false  "NeoForge loader version (required unless mc is 'latest' or 'release')"
// @Success   302
// @Failure   400  {string}  string
// @Router    /neoforge/download [get]
func (h *Handler) getDownload(c *gin.Context) {
	mc := c.Query("mc")
	loader := c.Query("loader")

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
