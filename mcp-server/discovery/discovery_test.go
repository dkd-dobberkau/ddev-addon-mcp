package discovery

import (
	"testing"
)

func TestParseServices(t *testing.T) {
	jsonData := `{"raw":{"name":"my-project","services":{"web":{"status":"running"},"db":{"status":"running"},"redis":{"status":"running"},"solr":{"status":"running"}}}}`
	services, err := ParseServices([]byte(jsonData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !services.Has("redis") {
		t.Error("expected redis to be detected")
	}
	if !services.Has("solr") {
		t.Error("expected solr to be detected")
	}
	if services.Has("mailhog") {
		t.Error("expected mailhog to not be detected")
	}
}

func TestParseServices_NoExtra(t *testing.T) {
	jsonData := `{"raw":{"name":"my-project","services":{"web":{"status":"running"},"db":{"status":"running"}}}}`
	services, err := ParseServices([]byte(jsonData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if services.Has("redis") {
		t.Error("expected redis to not be detected")
	}
}

func TestParseServices_InvalidJSON(t *testing.T) {
	_, err := ParseServices([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
