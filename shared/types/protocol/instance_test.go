package protocol

import "testing"

func TestParseUserKey(t *testing.T) {
	tests := []struct {
		in   string
		want InstanceKey
	}{
		{"order-processing", InstanceKey{SystemID: "order-processing", Version: VersionLatest}},
		{"order-processing@1.0.0", InstanceKey{SystemID: "order-processing", Version: "1.0.0"}},
		{"order-processing@1.0.0#abc123:dev",
			InstanceKey{SystemID: "order-processing", Version: "1.0.0", Hash: "abc123", Env: "dev"}},
		{"order-processing@1.0.0#abc123",
			InstanceKey{SystemID: "order-processing", Version: "1.0.0", Hash: "abc123"}},
		{"order-processing:2.0.0", InstanceKey{SystemID: "order-processing", Version: "2.0.0"}},
		{"order-processing:2.0.0:def456:prod",
			InstanceKey{SystemID: "order-processing", Version: "2.0.0", Hash: "def456", Env: "prod"}},
	}
	for _, tt := range tests {
		got, err := ParseUserKey(tt.in)
		if err != nil {
			t.Errorf("ParseUserKey(%q) unexpected error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseUserKey(%q) = %+v, want %+v", tt.in, got, tt.want)
		}
	}
}

func TestParseUserKeyErrors(t *testing.T) {
	for _, in := range []string{"", "   ", "@1.0.0"} {
		if _, err := ParseUserKey(in); err == nil {
			t.Errorf("ParseUserKey(%q) expected error, got none", in)
		}
	}
}

func TestParseUserKeyColonStringRoundTrip(t *testing.T) {
	key := InstanceKey{SystemID: "sys", Version: "1.0.0", Hash: "h", Env: "development"}
	parsed, err := ParseUserKey(key.ColonString())
	if err != nil {
		t.Fatalf("parse colon string: %v", err)
	}
	if parsed != key {
		t.Errorf("round trip mismatch: %+v != %+v", parsed, key)
	}

	parsed, err = ParseUserKey(key.String())
	if err != nil {
		t.Fatalf("parse at-form: %v", err)
	}
	if parsed != key {
		t.Errorf("round trip mismatch: %+v != %+v", parsed, key)
	}
}
