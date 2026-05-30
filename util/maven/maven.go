package maven

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Which identifies the modloader maven repo we're working with.
type Which int

const (
	Forge Which = iota
	Neoforge
)

// MavenMetadata works for both Forge and NeoForge — they use the same schema.
type MavenMetadata struct {
	Versioning struct {
		Latest   string   `xml:"latest"`
		Release  string   `xml:"release"`
		Versions []string `xml:"versions>version"`
	} `xml:"versioning"`
}

// Ver is a parsed version: the Minecraft version it targets and the loader's
// own version string.
type Ver struct {
	Minecraft string // e.g. "1.20.1"
	Loader    string // e.g. "47.3.0"
}

// ParseNeoforgeVersion turns "21.1.172" into Minecraft="1.21.1", Loader="21.1.172".
// NeoForge encodes the MC version in the first two dotted segments.
func ParseNeoforgeVersion(version string) (Ver, error) {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) < 2 {
		return Ver{}, fmt.Errorf("unexpected neoforge version format: %q", version)
	}
	minecraft := fmt.Sprintf("1.%s.%s", parts[0], parts[1])
	return Ver{Minecraft: minecraft, Loader: version}, nil
}

// ParseForgeVersion splits "1.20.1-47.3.0" into its MC and loader halves.
func ParseForgeVersion(version string) (Ver, error) {
	parts := strings.SplitN(version, "-", 2)
	if len(parts) != 2 {
		return Ver{}, fmt.Errorf("unexpected forge version format: %q", version)
	}
	return Ver{Minecraft: parts[0], Loader: parts[1]}, nil
}

func (w Which) String() string {
	if w == Neoforge {
		return "neoforge"
	}
	return "forge"
}

func (w Which) mavenURL() string {
	if w == Neoforge {
		return "https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml"
	}
	return "https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml"
}

func (w Which) hashPath() string {
	if w == Neoforge {
		return "/srv/shulker/hash/neoforge"
	}
	return "/srv/shulker/hash/forge"
}

func (w Which) baseDir() string {
	if w == Neoforge {
		return "/srv/shulker/neoforge"
	}
	return "/srv/shulker/forge"
}

// fetchRaw downloads the maven-metadata.xml body, with a real HTTP status check
// so we don't silently try to parse a 404 HTML page as XML later.
func (w Which) fetchRaw() ([]byte, error) {
	res, err := http.Get(w.mavenURL())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("maven request failed: %s", res.Status)
	}

	return io.ReadAll(res.Body)
}

func (w Which) Update() error {
	body, err := w.fetchRaw()
	if err != nil {
		return err
	}

	sum := sha256.Sum256(body)
	newHash := sum[:]

	if existing, err := os.ReadFile(w.hashPath()); err == nil {
		if bytes.Equal(existing, newHash) {
			fmt.Println("No changes detected, skipping update")
			return nil
		}
	}

	var meta MavenMetadata
	if err := xml.Unmarshal(body, &meta); err != nil {
		return err
	}

	// Deduplicate using sets, since many raw versions share an MC target.
	mcVersions := map[string]bool{}
	loaderVersions := map[string]bool{}

	for _, raw := range meta.Versioning.Versions {
		var fv Ver
		var parseErr error
		if w == Neoforge {
			fv, parseErr = ParseNeoforgeVersion(raw)
		} else {
			fv, parseErr = ParseForgeVersion(raw)
		}
		if parseErr != nil {
			continue
		}
		mcVersions[fv.Minecraft] = true
		loaderVersions[fv.Loader] = true
	}

	mcSlice := make([]string, 0, len(mcVersions))
	for v := range mcVersions {
		mcSlice = append(mcSlice, v)
	}
	loaderSlice := make([]string, 0, len(loaderVersions))
	for v := range loaderVersions {
		loaderSlice = append(loaderSlice, v)
	}

	if err := writeJSON(w.baseDir()+"/mc.json", mcSlice); err != nil {
		return err
	}
	if err := writeJSON(w.baseDir()+"/loader.json", loaderSlice); err != nil {
		return err
	}

	// Persist the hash LAST: if anything above failed, we want the next run
	// to retry rather than treat the broken state as up-to-date.
	if err := os.WriteFile(w.hashPath(), newHash, 0644); err != nil {
		return err
	}

	fmt.Println("Updated mc.json and loader.json")
	return nil
}

func writeJSON(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (w Which) Fetch() (*MavenMetadata, error) {
	body, err := w.fetchRaw()
	if err != nil {
		return nil, err
	}

	var meta MavenMetadata
	if err := xml.Unmarshal(body, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// Server describes a desired install: which loader, which MC version, which
// loader build. Minecraft may also be "release" or "latest" to follow upstream.
type Server struct {
	Platform  Which
	Minecraft string
	Loader    string // ignored when Minecraft is "release" or "latest"
}

func (s Server) ResolveMavenURL() (string, error) {
	metadata, err := s.Platform.Fetch()
	if err != nil {
		return "", err
	}

	var selectedVersion string

	switch s.Minecraft {
	case "release":
		selectedVersion = metadata.Versioning.Release
	case "latest":
		selectedVersion = metadata.Versioning.Latest
	default:
		if s.Platform == Neoforge {
			parsed, err := ParseNeoforgeVersion(s.Loader)
			if err != nil {
				return "", fmt.Errorf("invalid neoforge loader %q: %w", s.Loader, err)
			}
			if s.Minecraft != "" && parsed.Minecraft != s.Minecraft {
				return "", fmt.Errorf(
					"loader %q targets Minecraft %s, not %s",
					s.Loader, parsed.Minecraft, s.Minecraft,
				)
			}
			selectedVersion = s.Loader
		} else {
			selectedVersion = fmt.Sprintf("%s-%s", s.Minecraft, s.Loader)
		}

		found := false
		for _, ver := range metadata.Versioning.Versions {
			if selectedVersion == ver {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf(
				"version %q not found in %s repository (check that loader %q is published for Minecraft %s)",
				selectedVersion, s.Platform, s.Loader, s.Minecraft,
			)
		}
	}
	if s.Platform == Neoforge {
		return fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/neoforge-%s-installer.jar", selectedVersion, selectedVersion), nil
	}
	return fmt.Sprintf("https://maven.minecraftforge.net/net/minecraftforge/forge/%s/forge-%s-installer.jar", selectedVersion, selectedVersion), nil
}
