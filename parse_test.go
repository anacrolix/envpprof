package envpprof

import (
	"reflect"
	"testing"
)

func TestParseGOPPROF(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want []gopprofItem
	}{
		{"", nil},
		{"cpu", []gopprofItem{{"cpu", ""}}},
		{"http,block", []gopprofItem{{"http", ""}, {"block", ""}}},
		{"http=:6060,cpu", []gopprofItem{{"http", ":6060"}, {"cpu", ""}}},
		{"cpu, heap", []gopprofItem{{"cpu", ""}, {"heap", ""}}},
		{" http = 6060 ", []gopprofItem{{"http", "6060"}}},
		{"cpu,", []gopprofItem{{"cpu", ""}}},
		{",,cpu,,heap,", []gopprofItem{{"cpu", ""}, {"heap", ""}}},
		{"http=a=b", []gopprofItem{{"http", "a=b"}}},
		// Left for the caller to report as an unexpected key.
		{"=6060", []gopprofItem{{"", "6060"}}},
	} {
		if got := parseGOPPROF(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parseGOPPROF(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
