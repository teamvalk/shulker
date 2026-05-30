package fabric

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	baseURL = "https://meta.fabricmc.net/v2/versions/"
	baseDir = "/srv/shulker/fabric"
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
	URL     string `json:"url"`
	Maven   string `json:"maven"`
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

type Component string

const (
	ComponentGame      Component = "game"
	ComponentLoader    Component = "loader"
	ComponentInstaller Component = "installer"
)

func (c Component) URL() string      { return baseURL + string(c) }
func (c Component) Path() string     { return filepath.Join(baseDir, string(c)+".json") }
func (c Component) HashPath() string { return c.Path() + ".sha256" }

type Server struct {
	Game      string
	Loader    string
	Installer string
}

func (s Server) ResolveFabricURL() string {
	return fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/%s/server/jar",
		s.Game, s.Loader, s.Installer)
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// update fetches the component list, parses it into []T, and writes it to disk
// if the upstream content has changed since the last run.
func update[T any](ctx context.Context, c Component) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL(), nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", c, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fetch %s: status %d", c, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	sum := sha256.Sum256(body)
	newHash := sum[:]

	oldHash, err := os.ReadFile(c.HashPath())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read existing hash: %w", err)
	}
	if bytes.Equal(newHash, oldHash) {
		return nil
	}

	var versions []T
	if err := json.Unmarshal(body, &versions); err != nil {
		return fmt.Errorf("parse %s: %w", c, err)
	}

	encoded, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", c, err)
	}

	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(c.Path(), encoded, 0o644); err != nil {
		return fmt.Errorf("write json: %w", err)
	}
	if err := os.WriteFile(c.HashPath(), newHash, 0o644); err != nil {
		return fmt.Errorf("write hash: %w", err)
	}
	return nil
}

func UpdateGame(ctx context.Context) error {
	return update[GameVersion](ctx, ComponentGame)
}

func UpdateLoader(ctx context.Context) error {
	return update[LoaderVersion](ctx, ComponentLoader)
}

func UpdateInstaller(ctx context.Context) error {
	return update[InstallerVersion](ctx, ComponentInstaller)
}

func UpdateAll(ctx context.Context) error {
	if err := UpdateGame(ctx); err != nil {
		return err
	}
	if err := UpdateLoader(ctx); err != nil {
		return err
	}
	return UpdateInstaller(ctx)
}