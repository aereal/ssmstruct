package ssmstruct

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/google/go-cmp/cmp"
)

func TestUnmarshal(t *testing.T) {
	testCases := []struct {
		name       string
		parameters []types.Parameter
		want       *testStruct
		wantErr    bool
	}{
		{
			"ok/scalars",
			[]types.Parameter{
				{Name: strRef("str"), Type: types.ParameterTypeString, Value: strRef("strValue")},
				{Name: strRef("int"), Type: types.ParameterTypeString, Value: strRef("64")},
				{Name: strRef("uint"), Type: types.ParameterTypeString, Value: strRef("128")},
				{Name: strRef("boolean"), Type: types.ParameterTypeString, Value: strRef("true")},
			},
			&testStruct{Str: "strValue", Int: 64, Uint: 128, Boolean: true},
			false,
		},
		{
			"ok/slice",
			[]types.Parameter{
				{Name: strRef("strSlice"), Type: types.ParameterTypeStringList, Value: strRef("a,b,c")},
				{Name: strRef("intSlice"), Type: types.ParameterTypeStringList, Value: strRef("1,2,3")},
			},
			&testStruct{StrSlice: []string{"a", "b", "c"}, IntSlice: []int{1, 2, 3}},
			false,
		},
		{
			"ok/text unmarshaler",
			[]types.Parameter{
				{Name: strRef("pair"), Type: types.ParameterTypeString, Value: strRef("name=aereal")},
				{Name: strRef("pairs"), Type: types.ParameterTypeStringList, Value: strRef("name=yuno,name=miyako,name=sae,name=hiro")},
			},
			&testStruct{
				Pair: &pair{Key: "name", Value: "aereal"},
				Pairs: []*pair{
					{Key: "name", Value: "yuno"},
					{Key: "name", Value: "miyako"},
					{Key: "name", Value: "sae"},
					{Key: "name", Value: "hiro"},
				}},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got testStruct
			err := Unmarshal(tc.parameters, &got)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Fatalf("wantErr=%v but got=%v", tc.wantErr, gotErr)
			}
			if diff := cmp.Diff(tc.want, &got); diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
		})
	}
}

type testStruct struct {
	Str      string   `ssmp:"str"`
	Int      int      `ssmp:"int"`
	Uint     uint     `ssmp:"uint"`
	Boolean  bool     `ssmp:"boolean"`
	StrSlice []string `ssmp:"strSlice"`
	IntSlice []int    `ssmp:"intSlice"`
	Pair     *pair    `ssmp:"pair"`
	Pairs    []*pair  `ssmp:"pairs"`
}

type pair struct {
	Key   string
	Value string
}

func (x *pair) UnmarshalText(text []byte) error {
	xs := strings.SplitN(string(text), "=", 2)
	y := pair{Key: xs[0], Value: xs[1]}
	*x = y
	return nil
}

func strRef(s string) *string {
	return &s
}
