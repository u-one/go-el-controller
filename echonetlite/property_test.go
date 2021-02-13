package echonetlite

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPropertySerialize(t *testing.T) {
	p := Property{Code: 0xd5, Len: 4, Data: []byte{0x01, 0x05, 0xff, 0x01}}

	got := p.Serialize()
	want := Data{0xd5, 0x04, 0x01, 0x05, 0xff, 0x01}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}

}
