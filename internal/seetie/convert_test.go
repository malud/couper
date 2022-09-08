package seetie

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

func Test_stringListToValue(t *testing.T) {
	tests := []struct {
		slice []string
	}{
		{[]string{"a", "b"}},
		{[]string{}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.slice), func(subT *testing.T) {
			val := stringListToValue(tt.slice)
			valType := val.Type()
			if !valType.IsListType() {
				t.Error("Expected value type to be list")
			}
			if *valType.ListElementType() != cty.String {
				t.Error("Expected list to contain string values")
			}
			sl := val.AsValueSlice()
			if len(sl) != len(tt.slice) {
				t.Errorf("Wrong number of items; want: %d, got: %d", len(tt.slice), len(sl))
			}
			for i, v := range tt.slice {
				if sl[i].AsString() != v {
					t.Errorf("Wrong item at position %d; want %q, got %q", i, v, sl[i])
				}
			}
		})
	}
}

func Test_ValueToLogFields(t *testing.T) {
	type testCase struct {
		name   string
		val    cty.Value
		expLog logrus.Fields
	}
	for _, tc := range []testCase{
		{
			name:   "form body",
			val:    ValuesMapToValue(map[string][]string{"a": []string{"b"}}),
			expLog: logrus.Fields{"v": logrus.Fields{"a": []interface{}{"b"}}},
		},
		{
			name:   "cookies",
			val:    CookiesToMapValue([]*http.Cookie{&http.Cookie{Name: "c", Value: "d"}}),
			expLog: logrus.Fields{"v": logrus.Fields{"c": "d"}},
		},
		{
			name:   "headers",
			val:    HeaderToMapValue(http.Header{"c": []string{"d"}}),
			expLog: logrus.Fields{"v": logrus.Fields{"c": "d"}},
		},
	} {
		t.Run(tc.name, func(subT *testing.T) {
			logs := cty.MapVal(map[string]cty.Value{"v": tc.val})
			lf := ValueToLogFields(logs)
			if !cmp.Equal(tc.expLog, lf) {
				t.Errorf("Expected\n%#v, got:\n%#v", tc.expLog, lf)
			}
		})
	}
}
