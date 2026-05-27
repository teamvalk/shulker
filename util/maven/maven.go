package maven

import "net/http"

// Which - Which enum is it, Forge or Neoforge
type Which int

const (
	Forge = iota
	Neoforge
)

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

func (w Which) Fetch() (error, *MavenMetadata) {
	// sel - Selected
	var sel string

	if w == 0 {
		sel = "https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml"
	} else {
		sel = "https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml"
	}

	res, err := http.Get(sel)
	if err != nil {
		return err, nil
	}
	
}
