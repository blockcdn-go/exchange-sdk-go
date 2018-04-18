package query

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// NewDecoder 返回一个新的解码器
func NewDecoder() *Decoder {
	return &Decoder{cache: newCache()}
}

// Decoder 将map[string][]string反编码为一个struct
type Decoder struct {
	cache             *cache
	zeroEmpty         bool
	ignoreUnknownKeys bool
}

// SetAliasTag 修改用来定位自定义字段别名的tag
func (d *Decoder) SetAliasTag(tag string) {
	d.cache.tag = tag
}

// ZeroEmpty 控制解码器遇到空值时的行为
func (d *Decoder) ZeroEmpty(z bool) {
	d.zeroEmpty = z
}

// IgnoreUnknownKeys 控制解码器遇到未知key时的行为
func (d *Decoder) IgnoreUnknownKeys(i bool) {
	d.ignoreUnknownKeys = i
}

// RegisterConverter 为自定义类型注册一个转换函数
func (d *Decoder) RegisterConverter(value interface{}, converterFunc Converter) {
	d.cache.registerConverter(value, converterFunc)
}

// Decode 将一个map[string][]string解码为一个struct
func (d *Decoder) Decode(src map[string][]string, dst interface{}) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("url: interface must be a pointer to struct")
	}

	v = v.Elem()
	t := v.Type()
	errors := MultiError{}
	for path, values := range src {
		if parts, err := d.cache.parsePath(path, t); err == nil {
			if err = d.decode(v, path, parts, values); err != nil {
				errors[path] = err
			}
		} else if !d.ignoreUnknownKeys {
			errors[path] = fmt.Errorf("url: invalid path %q", path)
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return d.checkRequired(t, src, "")
}

func (d *Decoder) checkRequired(t reflect.Type, src map[string][]string, prefix string) error {
	struc := d.cache.get(t)
	if struc == nil {
		return errors.New("cache fail")
	}

	for _, f := range struc.fields {
		if f.typ.Kind() == reflect.Struct {
			err := d.checkRequired(f.typ, src, prefix+f.alias+".")
			if err != nil {
				if !f.anon {
					return err
				}

				err2 := d.checkRequired(f.typ, src, prefix)
				if err2 != nil {
					return err
				}
			}
		}

		if f.required {
			key := f.alias
			if prefix != "" {
				key = prefix + key
			}
			if isEmpty(f.typ, src[key]) {
				return fmt.Errorf("%v is empty", key)
			}
		}
	}

	return nil
}

func isEmpty(t reflect.Type, value []string) bool {
	if len(value) == 0 {
		return true
	}
	switch t.Kind() {
	case boolType, float32Type, float64Type, intType, int8Type, int16Type, int32Type, int64Type, stringType, uint8Type, uint16Type, uint32Type, uint64Type:
		return len(value[0]) == 0
	}
	return false
}

func (d *Decoder) decode(v reflect.Value, path string, parts []pathPart, values []string) error {
	for _, name := range parts[0].path {
		if v.Type().Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.FieldByName(name)
	}

	if !v.CanSet() {
		return nil
	}

	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if v.IsNil() {
			v.Set(reflect.New(t))
		}
		v = v.Elem()
	}

	if len(parts) > 1 {
		idx := parts[0].index
		if v.IsNil() || v.Len() < idx+1 {
			value := reflect.MakeSlice(t, idx+1, idx+1)
			if v.Len() < idx+1 {
				reflect.Copy(value, v)
			}
			v.Set(value)
		}
		return d.decode(v.Index(idx), path, parts[1:], values)
	}

	conv := d.cache.converter(t)
	if conv == nil && t.Kind() == reflect.Slice {
		var items []reflect.Value
		elemT := t.Elem()
		isPtrElem := elemT.Kind() == reflect.Ptr
		if isPtrElem {
			elemT = elemT.Elem()
		}

		conv := d.cache.converter(elemT)
		if conv == nil {
			conv = builtinConverters[elemT.Kind()]
			if conv == nil {
				return fmt.Errorf("url: converter not found for %v", elemT)
			}
		}

		for key, value := range values {
			if value == "" {
				if d.zeroEmpty {
					items = append(items, reflect.Zero(elemT))
				}
			} else if m := isTextUnmarshaler(v); m.IsValid {
				u := reflect.New(elemT)
				if m.IsPtr {
					u = reflect.New(reflect.PtrTo(elemT).Elem())
				}
				if err := u.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)); err != nil {
					return ConversionError{
						Key:   path,
						Type:  t,
						Index: key,
						Err:   err,
					}
				}
				if m.IsPtr {
					items = append(items, u.Elem().Addr())
				} else if u.Kind() == reflect.Ptr {
					items = append(items, u.Elem())
				} else {
					items = append(items, u)
				}
			} else if item := conv(value); item.IsValid() {
				if isPtrElem {
					ptr := reflect.New(elemT)
					ptr.Elem().Set(item)
					item = ptr
				}
				if item.Type() != elemT && !isPtrElem {
					item = item.Convert(elemT)
				}
				items = append(items, item)
			} else {
				if strings.Contains(value, ",") {
					values := strings.Split(value, ",")
					for _, value := range values {
						if value == "" {
							if d.zeroEmpty {
								items = append(items, reflect.Zero(elemT))
							}
						} else if item := conv(value); item.IsValid() {
							if isPtrElem {
								ptr := reflect.New(elemT)
								ptr.Elem().Set(item)
								item = ptr
							}
							if item.Type() != elemT && !isPtrElem {
								item = item.Convert(elemT)
							}
							items = append(items, item)
						} else {
							return ConversionError{
								Key:   path,
								Type:  elemT,
								Index: key,
							}
						}
					}
				} else {
					return ConversionError{
						Key:   path,
						Type:  elemT,
						Index: key,
					}
				}
			}
		}
		value := reflect.Append(reflect.MakeSlice(t, 0, 0), items...)
		v.Set(value)
	} else {
		val := ""
		if len(values) > 0 {
			val = values[len(values)-1]
		}

		if val == "" {
			if d.zeroEmpty {
				v.Set(reflect.Zero(t))
			}
		} else if conv != nil {
			if value := conv(val); value.IsValid() {
				v.Set(value.Convert(t))
			} else {
				return ConversionError{
					Key:   path,
					Type:  t,
					Index: -1,
				}
			}
		} else if m := isTextUnmarshaler(v); m.IsValid {
			if err := m.Unmarshaler.UnmarshalText([]byte(val)); err != nil {
				return ConversionError{
					Key:   path,
					Type:  t,
					Index: -1,
					Err:   err,
				}
			}
		} else if conv := builtinConverters[t.Kind()]; conv != nil {
			if value := conv(val); value.IsValid() {
				v.Set(value.Convert(t))
			} else {
				return ConversionError{
					Key:   path,
					Type:  t,
					Index: -1,
				}
			}
		} else {
			return fmt.Errorf("url: converter not found for %v", t)
		}
	}
	return nil
}

func isTextUnmarshaler(v reflect.Value) unmarshaler {
	m := unmarshaler{}

	if v.CanAddr() {
		v = v.Addr()
	}
	if m.Unmarshaler, m.IsValid = v.Interface().(encoding.TextUnmarshaler); m.IsValid {
		return m
	}

	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		if t = t.Elem(); t.Kind() == reflect.Ptr {
			t = reflect.PtrTo(t.Elem())
			v = reflect.Zero(t)
			m.IsPtr = true
			m.Unmarshaler, m.IsValid = v.Interface().(encoding.TextUnmarshaler)
			return m
		}
	}

	v = reflect.New(t)
	m.Unmarshaler, m.IsValid = v.Interface().(encoding.TextUnmarshaler)
	return m
}

type unmarshaler struct {
	Unmarshaler encoding.TextUnmarshaler
	IsPtr       bool
	IsValid     bool
}
