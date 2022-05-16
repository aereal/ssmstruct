package ssmstruct

import (
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
	Str     string `ssmp:"str"`
	Int     int    `ssmp:"int"`
	Uint    uint   `ssmp:"uint"`
	Boolean bool   `ssmp:"boolean"`
}

func strRef(s string) *string {
	return &s
}
