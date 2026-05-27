package maven

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// Which - Which enum is it, Forge or Neoforge
type Which int

const (
	Forge = iota
	Neoforge
)

type Server struct {
	Platform Which
	Version  string
}

// MavenMetadata Works for forge & nf
type MavenMetadata struct {
	Versioning struct {
		Latest  string `xml:"latest"`
		Release string `xml:"release"`
		// Versions
		// Goes into versions, selects all <version>(s)
		Versions []string `xml:"versions>version"`
	} `xml:"versioning"`
}

func (w Which) Fetch() (*MavenMetadata, error) {
	// sel - Selected
	var sel string

	if w == Neoforge {
		sel = "https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml"
	} else {
		sel = "https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml"
	}

	res, err := http.Get(sel)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() // i dont know how to handle that error
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var meta MavenMetadata
	if err := xml.Unmarshal(body, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

/**
func (w Which) GroupID() string {
	if w == Neoforge {
		return "net.neoforged"
	}
	return "net.minecraftforge"
}

// ArtifactID returns the Maven artifact name for the given platform.
// Keeping this on Which means ResolveMavenURL callers don't have to
// remember which string belongs to which platform.
func (w Which) ArtifactID() string {
	if w == Neoforge {
		return "neoforge"
	}
	return "forge"
}
*/

func (s Server) ResolveMavenURL() string {
	// Convert "net.neoforged" -> "net/neoforged" to match Maven's folder structure
	var url string
	if s.Platform == Neoforge {
		url = fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/neoforge-%s-installer.jar", s.Version, s.Version)
	} else {
		url = fmt.Sprintf("https://maven.minecraftforge.net/net/minecraftforge/forge/%s/forge-%s-installer.jar", s.Version, s.Version)
	}
	return url
}

func (s Server) Save() {

}
