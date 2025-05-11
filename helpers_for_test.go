package qilin

import (
	"net/url"
	"testing"
	"time"
)

func MustURL(t *testing.T, uri string) *url.URL {
	t.Helper()
	v, err := url.Parse(uri)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}
	return v
}

func MustTime(t *testing.T, v string) time.Time {
	t.Helper()
	vt, err := time.Parse(time.RFC3339, v)
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}
	return vt
}
