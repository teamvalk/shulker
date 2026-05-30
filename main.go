// @title           Shulker API
// @version         1.0
// @description     Minecraft mod loader download proxy for Forge and NeoForge.
// @host            localhost:8080
// @basePath        /

package main

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "shulker/docs"

	"shulker/api/forge"
	"shulker/api/neoforge"
)

func main() {
	router := gin.Default()
	forge.RegisterRoutes(router)
	neoforge.RegisterRoutes(router)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	err := router.Run()
	if err != nil {
		return
	}
}
