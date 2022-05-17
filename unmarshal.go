package ssmstruct

import (
	"encoding"
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

var (
	textUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// NewDecoder returns a new decoder that reads from params.
func NewDecoder(params []types.Parameter, opts ...Option) *Decoder {
	dec := &Decoder{params: params}
	for _, opt := range opts {
		opt(dec)
	}
	return dec
}

// Decoder decodes values from fetched SSM parameters.
type Decoder struct {
	params     []types.Parameter
	pathPrefix string
}

// Option is a function changes the decoder's behavior
type Option func(d *Decoder)

// WithPathPrefix returns an Option that indicates the decoder to treat pathPrefix as common prefix.
//
// The decoder searches corresponding parameter using its name without pathPrefix.
func WithPathPrefix(pathPrefix string) Option {
	return func(d *Decoder) {
		d.pathPrefix = pathPrefix
	}
}

// Decode decodes SSM parameters that previously given into Go values.
//
// Supported Go value types are:
// - string
// - int family
// - slices
//   - the element type must be also Decode()'s supported type
//   - only string list parameter can be decoded
// - the type implements encoding.TextUnmarshaler
func (d *Decoder) Decode(v interface{}) error {
	vt := reflect.ValueOf(v)
	if vt.IsNil() || !(vt.Kind() == reflect.Pointer && vt.Elem().Kind() == reflect.Struct) {
		return errors.New("v must be a pointer to the struct type")
	}

	byName := map[string]types.Parameter{}
	for _, p := range d.params {
		key := *p.Name
		if d.pathPrefix != "" {
			key = strings.TrimPrefix(key, d.pathPrefix)
		}
		byName[key] = p
	}

	structType := vt.Elem()
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

func decodeScalar(val string, fieldValue reflect.Value) error {
	if fieldValue.Type().Implements(textUnmarshaler) {
		fv := fieldValue
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		if v, ok := fv.Interface().(encoding.TextUnmarshaler); ok {
			return v.UnmarshalText([]byte(val))
		}
	}

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
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, fieldValue.Type().Bits())
		if err != nil {
			return fmt.Errorf("cannot convert value %q to %s: %w", val, kind, err)
		}
		if fieldValue.OverflowFloat(n) {
			return fmt.Errorf("cannot conver value %q: overflow float size", val)
		}
		fieldValue.SetFloat(n)
	case reflect.Bool:
		switch val {
		case "true":
			fieldValue.SetBool(true)
		case "false":
			fieldValue.SetBool(false)
		default:
			return fmt.Errorf("invalid boolean: %s", val)
		}
	default:
		return fmt.Errorf("%T is cannot be unmarshaled", kind)
	}
	return nil
}

func Unmarshal(params []types.Parameter, v interface{}) error {
	return NewDecoder(params).Decode(v)
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
