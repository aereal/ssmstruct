package ssmstruct

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

func ExampleDecoder_Decode() {
	params := []types.Parameter{
		{Name: strRef("/str"), Type: types.ParameterTypeString, Value: strRef("strValue")},
		{Name: strRef("/int"), Type: types.ParameterTypeString, Value: strRef("64")},
		{Name: strRef("/strSlice"), Type: types.ParameterTypeStringList, Value: strRef("a,b,c")},
	}
	type x struct {
		Str      string   `ssmp:"/str"`
		Int      int      `ssmp:"/int"`
		StrSlice []string `ssmp:"/strSlice"`
	}
	var v x
	if err := NewDecoder(params).Decode(&v); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", v)
	// Output:
	// ssmstruct.x{Str:"strValue", Int:64, StrSlice:[]string{"a", "b", "c"}}
}

func ExampleDecoder_Decode_withPathPrefix() {
	params := []types.Parameter{
		{Name: strRef("/my/str"), Type: types.ParameterTypeString, Value: strRef("strValue")},
		{Name: strRef("/my/int"), Type: types.ParameterTypeString, Value: strRef("64")},
		{Name: strRef("/my/strSlice"), Type: types.ParameterTypeStringList, Value: strRef("a,b,c")},
	}
	type x struct {
		Str      string   `ssmp:"/str"`
		Int      int      `ssmp:"/int"`
		StrSlice []string `ssmp:"/strSlice"`
	}
	var v x
	if err := NewDecoder(params, WithPathPrefix("/my")).Decode(&v); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", v)
	// Output:
	// ssmstruct.x{Str:"strValue", Int:64, StrSlice:[]string{"a", "b", "c"}}
}
