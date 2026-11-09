package hexcolor_test

import (
	"encoding/json"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/adm87/finch/pkg/hexcolor"
)

func TestParse(t *testing.T) {
	tests := []struct {
		in   string
		want hexcolor.Color
	}{
		{"#6495ED", hexcolor.Color{R: 0x64, G: 0x95, B: 0xed, A: 0xff}},
		{"#6495ED80", hexcolor.Color{R: 0x64, G: 0x95, B: 0xed, A: 0x80}},
		{"6495ed", hexcolor.Color{R: 0x64, G: 0x95, B: 0xed, A: 0xff}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := hexcolor.Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) = %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseRejects(t *testing.T) {
	for _, in := range []string{"", "#fff", "#12345", "#123456789", "#zzzzzz"} {
		if _, err := hexcolor.Parse(in); err == nil {
			t.Errorf("Parse(%q) returned no error", in)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	want := hexcolor.Color{R: 0x64, G: 0x95, B: 0xed, A: 0xff}

	var fromYAML struct {
		C hexcolor.Color `yaml:"c"`
	}
	if err := yaml.Unmarshal([]byte(`c: "#6495ED"`), &fromYAML); err != nil {
		t.Fatalf("yaml.Unmarshal = %v", err)
	}
	if fromYAML.C != want {
		t.Errorf("from YAML = %+v, want %+v", fromYAML.C, want)
	}

	var fromJSON struct{ C hexcolor.Color }
	if err := json.Unmarshal([]byte(`{"C": "#6495ED"}`), &fromJSON); err != nil {
		t.Fatalf("json.Unmarshal = %v", err)
	}
	if fromJSON.C != want {
		t.Errorf("from JSON = %+v, want %+v", fromJSON.C, want)
	}
}
