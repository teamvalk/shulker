package fabric

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

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

func ResolveBaseURL() string {
	return "https://meta.fabricmc.net/v2/versions/"
}

type WriteAction int

const (
	Installer WriteAction = iota
	Loader
	Game
)

func (w WriteAction) Write(body []byte, hash []byte) error {
	var basePath string
	switch w {
	case Loader:
		basePath = "/srv/shulker/fabric/loader.json"
	case Installer:
		basePath = "/srv/shulker/fabric/installer.json"
	case Game:
		basePath = "/srv/shulker/fabric/game.json"
	}
	err := os.WriteFile(basePath, body, 0644)
	if err != nil {
		return err
	}
	err = os.WriteFile(basePath+".sha256", hash, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (w WriteAction) ReadAction() string {
	var basePath string
	switch w {
	case Loader:
		basePath = "/srv/shulker/fabric/loader.json"
	case Installer:
		basePath = "/srv/shulker/fabric/installer.json"
	case Game:
		basePath = "/srv/shulker/fabric/game.json"
	}
	return basePath
}

func (w WriteAction) GetFabricComponent() string {
	var baseURL string = "https://meta.fabricmc.net/v2/versions/"
	switch w {
	case Loader:
		return baseURL + "loader"
	case Installer:
		return baseURL + "installer"
	case Game:
		return baseURL + "game"
	}
	return ""
}

func GetHash(body []byte) []byte {
	sum := sha256.Sum256(body)
	return sum[:]
}

func (w WriteAction) Update() error {
	baseURL := w.GetFabricComponent()
	resp, err := http.Get(baseURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	httpHash := GetHash(body)

	fileHash, err := os.ReadFile(w.ReadAction() + ".sha256")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
		} else {
			return err
		}
	}
	if bytes.Equal(httpHash, fileHash) {
		return nil
	}
	err = w.Write(body, httpHash)
	if err != nil {
		return err
	}
	return nil
}
