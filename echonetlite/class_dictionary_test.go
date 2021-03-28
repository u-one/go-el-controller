package echonetlite

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_ClassDictionary_get(t *testing.T) {
	t.Parallel()

	cInfo0101 := ClassInfo{0x01, 0x01, PropertyDictionary{}, "class info 0101"}
	cInfo0102 := ClassInfo{0x01, 0x02, PropertyDictionary{}, "class info 0102"}
	cInfo0201 := ClassInfo{0x02, 0x01, PropertyDictionary{}, "class info 0201"}

	testcases := []struct {
		name     string
		dict     ClassDictionary
		inputCg  ClassGroupCode
		inputC   ClassCode
		wantInfo ClassInfo
		wantBool bool
	}{
		{
			name: "Hit",
			dict: ClassDictionary{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101, 0x02: cInfo0102},
				0x02: map[ClassCode]ClassInfo{0x01: cInfo0201},
			},
			inputCg:  0x01,
			inputC:   0x02,
			wantInfo: cInfo0102,
			wantBool: true,
		},
		{
			name:     "ClassGroup no hit",
			dict:     ClassDictionary{0x01: map[ClassCode]ClassInfo{0x01: cInfo0101}},
			inputCg:  0x02,
			inputC:   0x01,
			wantInfo: ClassInfo{},
			wantBool: false,
		},
		{
			name:     "Class no hit",
			dict:     ClassDictionary{0x01: map[ClassCode]ClassInfo{0x01: cInfo0101}},
			inputCg:  0x01,
			inputC:   0x02,
			wantInfo: ClassInfo{},
			wantBool: false,
		},
		{
			name:     "No data",
			dict:     ClassDictionary{},
			inputCg:  0x01,
			inputC:   0x02,
			wantInfo: ClassInfo{},
			wantBool: false,
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotInfo, gotBool := tc.dict.get(tc.inputCg, tc.inputC)

			if diff := cmp.Diff(tc.wantInfo, gotInfo); diff != "" {
				t.Errorf("ClassInfo differs: (-want +got)\n%s", diff)
			}

			if diff := cmp.Diff(tc.wantBool, gotBool); diff != "" {
				t.Errorf("bool result differs: (-want +got)\n%s", diff)
			}

		})
	}

}

func Test_ClassDictionary_Get(t *testing.T) {
	t.Parallel()

	cInfo0102 := ClassInfo{0x01, 0x02, PropertyDictionary{}, "class info 0102"}
	cInfo0201 := ClassInfo{0x02, 0x01, PropertyDictionary{}, "class info 0201"}
	dict := ClassDictionary{
		0x01: map[ClassCode]ClassInfo{0x02: cInfo0102},
		0x02: map[ClassCode]ClassInfo{0x01: cInfo0201},
	}

	testcases := []struct {
		name     string
		inputCg  ClassGroupCode
		inputC   ClassCode
		wantInfo ClassInfo
		wantBool bool
	}{
		{
			name:    "Hit",
			inputCg: 0x02, inputC: 0x01,
			wantInfo: cInfo0201,
		},
		{
			name:    "No hit",
			inputCg: 0x03, inputC: 0x03,
			wantInfo: ClassInfo{0x03, 0x03, PropertyDictionary{}, "unknown"},
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotInfo := dict.Get(tc.inputCg, tc.inputC)

			if diff := cmp.Diff(tc.wantInfo, gotInfo); diff != "" {
				t.Errorf("ClassInfo differs: (-want +got)\n%s", diff)
			}
		})
	}

}

func Test_ClassDictionary_add(t *testing.T) {
	t.Parallel()

	cInfo0101 := ClassInfo{0x01, 0x01, PropertyDictionary{}, "class info 0101"}
	cInfo0101Mod := ClassInfo{0x01, 0x01, PropertyDictionary{}, "class info 0101 mod"}
	cInfo0102 := ClassInfo{0x01, 0x02, PropertyDictionary{}, "class info 0102"}
	cInfo0201 := ClassInfo{0x02, 0x01, PropertyDictionary{}, "class info 0201"}

	testcases := []struct {
		name      string
		dict      ClassDictionary
		inputCg   ClassGroupCode
		inputC    ClassCode
		inputInfo ClassInfo
		wantMap   ClassDictionary
	}{
		{
			name:    "Add to empty map",
			dict:    ClassDictionary{},
			inputCg: 0x01, inputC: 0x01, inputInfo: cInfo0101,
			wantMap: ClassDictionary{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101},
			},
		},
		{
			name: "ClassGroup exists",
			dict: ClassDictionary{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101},
			},
			inputCg: 0x01, inputC: 0x02, inputInfo: cInfo0102,
			wantMap: ClassDictionary{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101, 0x02: cInfo0102},
			},
		},
		{
			name: "ClassGroup not exists",
			dict: ClassDictionary{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101},
			},
			inputCg: 0x02, inputC: 0x01, inputInfo: cInfo0201,
			wantMap: ClassDictionary{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101},
				0x02: map[ClassCode]ClassInfo{0x01: cInfo0201},
			},
		},
		{
			name: "Both exists",
			dict: ClassDictionary{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101},
			},
			inputCg: 0x01, inputC: 0x01, inputInfo: cInfo0101Mod,
			wantMap: ClassDictionary{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101Mod},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.dict.add(tc.inputCg, tc.inputC, tc.inputInfo)

			if diff := cmp.Diff(tc.wantMap, tc.dict); diff != "" {
				t.Errorf("ClassInfo differs: (-want +got)\n%s", diff)
			}

		})
	}

}
