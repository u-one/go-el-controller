package echonetlite

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_NewObject(t *testing.T) {
	got := NewObject(ProfileGroup, Profile, 0x01)
	want := Object{0x0e, 0xf0, 0x01}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Object differs: (-want +got)\n%s", diff)
	}

}

func Test_NewObjectFromData(t *testing.T) {
	got := NewObjectFromData([]byte{0x0e, 0xf0, 0x01})
	want := Object{0x0e, 0xf0, 0x01}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Object differs: (-want +got)\n%s", diff)
	}

}

func TestObject_ClassGroupCode(t *testing.T) {
	o := NewObject(ProfileGroup, Profile, 0x01)
	got := o.classGroupCode()
	want := ProfileGroup

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ClassGroupCode differs: (-want +got)\n%s", diff)
	}
}

func TestObject_ClassCode(t *testing.T) {
	o := NewObject(ProfileGroup, Profile, 0x01)
	got := o.classCode()
	want := Profile

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ClassCode differs: (-want +got)\n%s", diff)
	}
}
