package ssmstruct

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

const (
	sliceDelim = ","
)

func decodeScalar(val string, fieldValue reflect.Value) error {
	switch kind := fieldValue.Kind(); kind {
	case reflect.String:
		fieldValue.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert value %q to %s: %w", val, kind, err)
		}
		if fieldValue.OverflowInt(n) {
			return fmt.Errorf("cannot convert value %q: overflow int size", val)
		}
		fieldValue.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert value %q to %s: %w", val, kind, err)
		}
		if fieldValue.OverflowUint(n) {
			return fmt.Errorf("cannot convert value %q: overflow uint size", val)
		}
		fieldValue.SetUint(n)
	case reflect.Bool:
		switch val {
		case "true":
			fieldValue.SetBool(true)
		case "false":
			fieldValue.SetBool(false)
		default:
			return fmt.Errorf("invalid boolean: %s", val)
		}
	}
	return nil
}

func decode(structType reflect.Value, params []types.Parameter) error {
	byName := map[string]types.Parameter{}
	for _, p := range params {
		byName[*p.Name] = p
	}

	typ := structType.Type()
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		fieldValue := structType.Field(i)
		tag := parseTag(field.Tag)
		if tag == nil {
			continue
		}
		param, ok := byName[tag.name]
		if !ok {
			continue
		}
		val := *param.Value
		switch kind := fieldValue.Kind(); kind {
		case reflect.Slice:
			if param.Type != types.ParameterTypeStringList {
				return fmt.Errorf("cannot convert %s to slice", param.Type)
			}
			els := strings.Split(val, sliceDelim)

			size := len(els)
			// grow slice size
			if size >= fieldValue.Cap() {
				nv := reflect.MakeSlice(fieldValue.Type(), fieldValue.Len(), size)
				reflect.Copy(nv, fieldValue)
				fieldValue.Set(nv)
				fieldValue.SetLen(size)
			}

			for i, el := range els {
				if err := decodeScalar(el, fieldValue.Index(i)); err != nil {
					return err
				}
			}
		default:
			if err := decodeScalar(val, fieldValue); err != nil {
				return err
			}
		}
	}
	return nil
}

func Unmarshal(params []types.Parameter, v interface{}) error {
	vt := reflect.ValueOf(v)
	if vt.IsNil() || !(vt.Kind() == reflect.Pointer && vt.Elem().Kind() == reflect.Struct) {
		return errors.New("v must be a pointer to the struct type")
	}

	return decode(vt.Elem(), params)
}

var (
	tagKey = "ssmp"
)

type option struct {
	name string
}

func parseTag(tag reflect.StructTag) *option {
	val, ok := tag.Lookup(tagKey)
	if !ok {
		return nil
	}
	return &option{name: val}
}
