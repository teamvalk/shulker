package maven

import (
	"math/rand/v2"
	"net/http"
	"testing"
)

func TestMavenForge(t *testing.T) {
	var f = Which(Forge)
	var versions, err = f.Fetch()
	if err != nil {
		t.Fatalf("failed to fetch metadata: %v", err)
	}

	var numberOfVersions = len(versions.Versioning.Versions)
	var tests = rand.IntN(numberOfVersions)

	pool := rand.Perm(numberOfVersions)

	var passed, failed int

	for i := 0; i < tests; i++ {
		randomVersion := versions.Versioning.Versions[pool[i]]

		s := Server{
			Platform: Forge,
			Version:  randomVersion,
		}

		url, err := s.ResolveMavenURL()
		if err != nil || url == "" {
			t.Errorf("ResolveMavenURL() failed for version %q: %v", randomVersion, err)
			failed++
			continue
		}

		res, err := http.Head(url)
		if err != nil || res.StatusCode != http.StatusOK {
			t.Errorf("[%d/%d] ping failed for version %q (status %d): %v", i, tests, randomVersion, res.StatusCode, err)
			failed++
			continue
		}

		t.Logf("[%d/%d] version %q -> %s", i, tests, randomVersion, url)
		passed++
	}

	t.Logf("results: %d/%d passed, %d/%d failed", passed, tests, failed, tests)
}
