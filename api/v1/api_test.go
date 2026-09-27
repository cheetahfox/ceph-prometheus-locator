package v1

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cheetahfox/ceph-prometheus-locator/cephlocator"
	"github.com/cheetahfox/ceph-prometheus-locator/config"
	"github.com/gofiber/fiber/v2"
)

func TestRewriteTargetDomains(t *testing.T) {
	input := []byte(`[{"targets":["atlas.pleaides:9283","192.168.66.69:8765"],"labels":{"job":"ceph"},"extra":{"keep":true}}]`)

	got, err := rewriteTargetDomains(input, "cheetahfox.com")
	if err != nil {
		t.Fatalf("rewriteTargetDomains() error = %v", err)
	}

	var groups []map[string]json.RawMessage
	if err := json.Unmarshal(got, &groups); err != nil {
		t.Fatalf("rewritten response is not valid JSON: %v", err)
	}

	var targets []string
	if err := json.Unmarshal(groups[0]["targets"], &targets); err != nil {
		t.Fatalf("failed to decode rewritten targets: %v", err)
	}
	if len(targets) != 2 || targets[0] != "atlas.cheetahfox.com:9283" || targets[1] != "192.168.66.69:8765" {
		t.Errorf("rewritten targets = %v, want [atlas.cheetahfox.com:9283 192.168.66.69:8765]", targets)
	}
	if _, exists := groups[0]["labels"]; !exists {
		t.Error("rewritten response lost labels")
	}
	if _, exists := groups[0]["extra"]; !exists {
		t.Error("rewritten response lost unknown fields")
	}
}

func TestRewriteTargetDomainsWithoutDomain(t *testing.T) {
	input := []byte(`[{"targets":["atlas.pleaides:9283"]}]`)

	got, err := rewriteTargetDomains(input, "")
	if err != nil {
		t.Fatalf("rewriteTargetDomains() error = %v", err)
	}
	if string(got) != string(input) {
		t.Errorf("rewriteTargetDomains() = %s, want unchanged %s", got, input)
	}
}

func TestRewriteTargetDomainsRejectsInvalidDomain(t *testing.T) {
	_, err := rewriteTargetDomains([]byte(`[{"targets":["atlas.pleaides:9283"]}]`), "bad/domain")
	if err == nil {
		t.Fatal("rewriteTargetDomains() expected an error for an invalid domain")
	}
}

func TestGetLocationProxiesAndRewritesTargets(t *testing.T) {
	upstreamQuery := make(chan string, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamQuery <- r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `[{"targets":["atlas.pleaides:9283"],"labels":{}}]`)
	}))
	defer upstream.Close()

	originalHosts := cephlocator.Hosts
	cephlocator.Hosts = map[string]*cephlocator.Host{
		"test": {HostUrl: upstream.URL + "/sd-config", Active: true},
	}
	defer func() {
		cephlocator.Hosts = originalHosts
	}()

	originalDomain := config.HostDomain
	config.HostDomain = "cheetahfox.com"
	defer func() {
		config.HostDomain = originalDomain
	}()

	app := fiber.New()
	app.Get("/sd-config", GetLocation)
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/sd-config?service=mgr-prometheus", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := <-upstreamQuery; got != "service=mgr-prometheus" {
		t.Errorf("upstream query = %q, want service=mgr-prometheus", got)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(body) != `[{"labels":{},"targets":["atlas.cheetahfox.com:9283"]}]` {
		t.Errorf("response body = %s, want rewritten target response", body)
	}
}
