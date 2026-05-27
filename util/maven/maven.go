package maven

import (
	"encoding/xml"
	"errors"
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

func (s Server) ResolveMavenURL() (string, error) {
	// Convert "net.neoforged" -> "net/neoforged" to match Maven's folder structure
	var url string
	var metadata, err = s.Platform.Fetch()
	var selectedVersion string
	if err != nil {
		return "", err
	}
	// if we for SURE know that this is
	var knownInArray bool

	switch s.Version {
	case "release":
		selectedVersion = metadata.Versioning.Release
		knownInArray = true
	case "latest":
		selectedVersion = metadata.Versioning.Latest
		knownInArray = true
	default:
		selectedVersion = s.Version
		knownInArray = false
	}
	if !knownInArray {
		for _, ver := range metadata.Versioning.Versions {
			if selectedVersion == ver {
				knownInArray = true
				break
			}
		}
		if !knownInArray {
			return "", errors.New("selected version unfound in repositories")
		}
	}

	if s.Platform == Neoforge {
		url = fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/neoforge-%s-installer.jar", selectedVersion, selectedVersion)
	} else {
		url = fmt.Sprintf("https://maven.minecraftforge.net/net/minecraftforge/forge/%s/forge-%s-installer.jar", selectedVersion, selectedVersion)
	}
	return url, nil
}

