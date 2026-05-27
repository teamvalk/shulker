package main

import (
	"fmt"
	"shulker/util/maven"
)

func main() {
	var server = maven.Server{
		Platform: maven.Forge,
		Version:  "26.1.2-64.0.8",
	}
	x := server.ResolveMavenURL()
	fmt.Println(server.Platform.Fetch())
	fmt.Println(x)
}
