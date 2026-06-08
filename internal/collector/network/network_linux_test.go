//go:build linux

package network

import "testing"

func TestParseSSFlowLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		line    string
		wantOK  bool
		wantSrc string
	}{
		{
			line:    "ESTAB 0 0 127.0.0.1:8080 192.168.1.2:54321",
			wantOK:  true,
			wantSrc: "127.0.0.1:8080",
		},
		{
			line:    "tcp   ESTAB 0 0 10.0.0.1:443 10.0.0.2:51234",
			wantOK:  true,
			wantSrc: "10.0.0.1:443",
		},
		{
			line:    "ESTAB 0 0 127.0.0.1:8080",
			wantOK:  false,
			wantSrc: "",
		},
		{
			line:    "	 cubic wscale:7 bytes_acked:123",
			wantOK:  false,
			wantSrc: "",
		},
	}

	for _, tt := range tests {
		key, ok := parseSSFlowLine(tt.line)
		if ok != tt.wantOK {
			t.Fatalf("line %q: ok=%v want %v", tt.line, ok, tt.wantOK)
		}
		if ok && key.src != tt.wantSrc {
			t.Fatalf("line %q: src=%q want %q", tt.line, key.src, tt.wantSrc)
		}
	}
}
