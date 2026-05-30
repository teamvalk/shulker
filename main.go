package main

import (
	"github.com/gin-gonic/gin"
	"shulker/api/forge"
	"shulker/api/neoforge"
)

func main() {
	router := gin.Default()
	forge.RegisterRoutes(router)
	neoforge.RegisterRoutes(router)
	err := router.Run()
	if err != nil {
		return
	}
}
