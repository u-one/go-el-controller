package echonetlite

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_ClassInfoMap_get(t *testing.T) {
	t.Parallel()

	cInfo0101 := ClassInfo{0x01, 0x01, map[PropertyCode]*PropertyInfo{}, "class info 0101"}
	cInfo0102 := ClassInfo{0x01, 0x02, map[PropertyCode]*PropertyInfo{}, "class info 0102"}
	cInfo0201 := ClassInfo{0x02, 0x01, map[PropertyCode]*PropertyInfo{}, "class info 0201"}

	testcases := []struct {
		name     string
		m        ClassInfoMap
		inputCg  ClassGroupCode
		inputC   ClassCode
		wantInfo ClassInfo
		wantBool bool
	}{
		{
			name: "Hit",
			m: ClassInfoMap{
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
			m:        ClassInfoMap{0x01: map[ClassCode]ClassInfo{0x01: cInfo0101}},
			inputCg:  0x02,
			inputC:   0x01,
			wantInfo: ClassInfo{},
			wantBool: false,
		},
		{
			name:     "Class no hit",
			m:        ClassInfoMap{0x01: map[ClassCode]ClassInfo{0x01: cInfo0101}},
			inputCg:  0x01,
			inputC:   0x02,
			wantInfo: ClassInfo{},
			wantBool: false,
		},
		{
			name:     "No data",
			m:        ClassInfoMap{},
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

			gotInfo, gotBool := tc.m.get(tc.inputCg, tc.inputC)

			if diff := cmp.Diff(tc.wantInfo, gotInfo); diff != "" {
				t.Errorf("ClassInfo differs: (-want +got)\n%s", diff)
			}

			if diff := cmp.Diff(tc.wantBool, gotBool); diff != "" {
				t.Errorf("bool result differs: (-want +got)\n%s", diff)
			}

		})
	}

}

func Test_ClassInfoMap_Get(t *testing.T) {
	t.Parallel()

	cInfo0102 := ClassInfo{0x01, 0x02, map[PropertyCode]*PropertyInfo{}, "class info 0102"}
	cInfo0201 := ClassInfo{0x02, 0x01, map[PropertyCode]*PropertyInfo{}, "class info 0201"}
	m := ClassInfoMap{
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
			wantInfo: ClassInfo{0x03, 0x03, map[PropertyCode]*PropertyInfo{}, "unknown"},
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotInfo := m.Get(tc.inputCg, tc.inputC)

			if diff := cmp.Diff(tc.wantInfo, gotInfo); diff != "" {
				t.Errorf("ClassInfo differs: (-want +got)\n%s", diff)
			}
		})
	}

}

func Test_ClassInfoMap_add(t *testing.T) {
	t.Parallel()

	cInfo0101 := ClassInfo{0x01, 0x01, map[PropertyCode]*PropertyInfo{}, "class info 0101"}
	cInfo0101Mod := ClassInfo{0x01, 0x01, map[PropertyCode]*PropertyInfo{}, "class info 0101 mod"}
	cInfo0102 := ClassInfo{0x01, 0x02, map[PropertyCode]*PropertyInfo{}, "class info 0102"}
	cInfo0201 := ClassInfo{0x02, 0x01, map[PropertyCode]*PropertyInfo{}, "class info 0201"}

	testcases := []struct {
		name      string
		m         ClassInfoMap
		inputCg   ClassGroupCode
		inputC    ClassCode
		inputInfo ClassInfo
		wantMap   ClassInfoMap
	}{
		{
			name:    "Add to empty map",
			m:       ClassInfoMap{},
			inputCg: 0x01, inputC: 0x01, inputInfo: cInfo0101,
			wantMap: ClassInfoMap{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101},
			},
		},
		{
			name: "ClassGroup exists",
			m: ClassInfoMap{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101},
			},
			inputCg: 0x01, inputC: 0x02, inputInfo: cInfo0102,
			wantMap: ClassInfoMap{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101, 0x02: cInfo0102},
			},
		},
		{
			name: "ClassGroup not exists",
			m: ClassInfoMap{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101},
			},
			inputCg: 0x02, inputC: 0x01, inputInfo: cInfo0201,
			wantMap: ClassInfoMap{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101},
				0x02: map[ClassCode]ClassInfo{0x01: cInfo0201},
			},
		},
		{
			name: "Both exists",
			m: ClassInfoMap{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101},
			},
			inputCg: 0x01, inputC: 0x01, inputInfo: cInfo0101Mod,
			wantMap: ClassInfoMap{
				0x01: map[ClassCode]ClassInfo{0x01: cInfo0101Mod},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.m.add(tc.inputCg, tc.inputC, tc.inputInfo)

			if diff := cmp.Diff(tc.wantMap, tc.m); diff != "" {
				t.Errorf("ClassInfo differs: (-want +got)\n%s", diff)
			}

		})
	}

}
