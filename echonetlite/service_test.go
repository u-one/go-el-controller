package echonetlite

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestESVTypeisRequest(t *testing.T) {

	testcases := []struct {
		input ESVType
		want  bool
	}{
		{SetI, true},
		{SetC, true},
		{Get, true},
		{InfReq, true},
		{SetGet, true},
		{SetRes, false},
		{GetRes, false},
		{Inf, false},
		{InfC, false},
		{InfCRes, false},
		{SetGetRes, false},
		{SetISNA, false},
		{SetCSNA, false},
		{GetSNA, false},
		{InfSNA, false},
		{SetGetSNA, false},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.input.String(), func(t *testing.T) {
			t.Parallel()

			got := tc.input.isRequest()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("result differs: (-want +got)\n%s", diff)
			}
		})
	}
}

func TestESVTypeisResponseOrNotification(t *testing.T) {
	testcases := []struct {
		input ESVType
		want  bool
	}{
		{SetI, false},
		{SetC, false},
		{Get, false},
		{InfReq, false},
		{SetGet, false},
		{SetRes, true},
		{GetRes, true},
		{Inf, true},
		{InfC, true},
		{InfCRes, true},
		{SetGetRes, true},
		{SetISNA, true},
		{SetCSNA, true},
		{GetSNA, true},
		{InfSNA, true},
		{SetGetSNA, true},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.input.String(), func(t *testing.T) {
			t.Parallel()

			got := tc.input.isResponseOrNotification()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("result differs: (-want +got)\n%s", diff)
			}
		})
	}
}
