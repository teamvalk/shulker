package forge

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
	h := &Handler{maven: maven.Forge}
	g := r.Group("/forge")
	{
		g.GET("/mc", h.getMinecraftVersions)
		g.GET("/loader", h.getLoaderVersions)
		g.POST("/update", h.postUpdate)
		g.GET("/download", h.getDownload)
	}
}

// @Summary  List Minecraft versions for Forge
// @Tags     forge
// @Produce  json
// @Success  200
// @Failure  500
// @Router   /forge/mc [get]
func (h *Handler) getMinecraftVersions(c *gin.Context) {
	content, err := os.ReadFile("/srv/shulker/forge/mc.json")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", content)
}

// @Summary  List Forge loader versions
// @Tags     forge
// @Produce  json
// @Success  200
// @Failure  500
// @Router   /forge/loader [get]
func (h *Handler) getLoaderVersions(c *gin.Context) {
	content, err := os.ReadFile("/srv/shulker/forge/loader.json")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", content)
}

// @Summary  Refresh cached Forge version data from Maven
// @Tags     forge
// @Success  200
// @Failure  500
// @Router   /forge/update [post]
func (h *Handler) postUpdate(c *gin.Context) {
	if err := h.maven.Update(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

// @Summary   Download Forge server jar
// @Tags      forge
// @Param     mc      query  string  true   "Minecraft version (or 'latest'/'release')"
// @Param     loader  query  string  false  "Forge loader version (required unless mc is 'latest' or 'release')"
// @Success   302
// @Failure   400  {string}  string
// @Router    /forge/download [get]
func (h *Handler) getDownload(c *gin.Context) {
	mc := c.Query("mc")
	loader := c.Query("loader")

	if mc == "" {
		c.String(http.StatusBadRequest, "missing 'mc' query parameter")
		return
	}
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
