package fabric

import "fmt"

type GameVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

type LoaderVersion struct {
	Separator string `json:"separator"`
	Build     int    `json:"build"`
	Maven     string `json:"maven"`
	Version   string `json:"version"`
	Stable    bool   `json:"stable"`
}

type InstallerVersion struct {
	Url     string `json:"url"`
	Maven   string `json:"maven"`
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

type Server struct {
	Game      string
	Loader    string
	Installer string
}

func (s Server) ResolveFabricURL() string {
	url := fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/%s/server/jar", s.Game, s.Loader, s.Installer)
	return url
}
