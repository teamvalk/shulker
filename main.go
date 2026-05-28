package main

import (
	"fmt"
	"shulker/util/maven"
)

func main() {
	var server = maven.Server{
		Platform: maven.Neoforge,
		Version:  "1.12",
	}
	x, err := server.ResolveMavenURL()
	y := maven.Which(0)
	y.Fetch()
	fmt.Println(y.Fetch())
	fmt.Println(x, err)
}
