package main

import (
	"fmt"
	"shulker/util/maven"
)

func main() {
	var server = maven.Server{
		Platform: maven.Neoforge,
		Version:  "latest",
	}
	x, err := server.ResolveMavenURL()

	fmt.Println(x, err)
}
