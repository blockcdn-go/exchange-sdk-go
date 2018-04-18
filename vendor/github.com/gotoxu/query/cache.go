package query

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var errInvalidPath = errors.New("url: invalid path")

func newCache() *cache {
	return &cache{
		m:       make(map[reflect.Type]*structInfo),
		regconv: make(map[reflect.Type]Converter),
		tag:     "url",
	}
}

type cache struct {
	l       sync.RWMutex
	m       map[reflect.Type]*structInfo
	regconv map[reflect.Type]Converter
	tag     string
}

// registerConverter 为自定义类型注册一个转换函数
func (c *cache) registerConverter(value interface{}, converterFunc Converter) {
	c.regconv[reflect.TypeOf(value)] = converterFunc
}

func (c *cache) parsePath(p string, t reflect.Type) ([]pathPart, error) {
	var struc *structInfo
	var field *fieldInfo
	var index64 int64
	var err error

	parts := make([]pathPart, 0)
	path := make([]string, 0)
	keys := strings.Split(p, ".")

	for i := 0; i < len(keys); i++ {
		if t.Kind() != reflect.Struct {
			return nil, errInvalidPath
		}
		if struc = c.get(t); struc == nil {
			return nil, errInvalidPath
		}
		if field = struc.get(keys[i]); field == nil {
			return nil, errInvalidPath
		}

		path = append(path, field.name)
		if field.ss {
			i++
			if i+1 > len(keys) {
				return nil, errInvalidPath
			}
			if index64, err = strconv.ParseInt(keys[i], 10, 0); err != nil {
				return nil, errInvalidPath
			}
			parts = append(parts, pathPart{
				path:  path,
				field: field,
				index: int(index64),
			})
			path = make([]string, 0)

			if field.typ.Kind() == reflect.Ptr {
				t = field.typ.Elem()
			} else {
				t = field.typ
			}

			if t.Kind() == reflect.Slice {
				t = t.Elem()
				if t.Kind() == reflect.Ptr {
					t = t.Elem()
				}
			}
		} else if field.typ.Kind() == reflect.Ptr {
			t = field.typ.Elem()
		} else {
			t = field.typ
		}
	}

	parts = append(parts, pathPart{
		path:  path,
		field: field,
		index: -1,
	})

	return parts, nil
}

func (c *cache) get(t reflect.Type) *structInfo {
	c.l.RLock()
	info := c.m[t]
	c.l.RUnlock()
	if info == nil {
		info = c.create(t, nil)
		c.l.Lock()
		c.m[t] = info
		c.l.Unlock()
	}
	return info
}

func (c *cache) create(t reflect.Type, info *structInfo) *structInfo {
	if info == nil {
		info = &structInfo{fields: []*fieldInfo{}}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			ft := field.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				bef := len(info.fields)
				c.create(ft, info)
				for _, fi := range info.fields[bef:len(info.fields)] {
					fi.required = false
				}
			}
		}
		c.createField(field, info)
	}

	return info
}

func (c *cache) createField(field reflect.StructField, info *structInfo) {
	alias, options := fieldAlias(field, c.tag)
	if alias == "-" {
		return
	}

	isSlice, isStruct := false, false
	ft := field.Type
	if ft.Kind() == reflect.Ptr {
		ft = ft.Elem()
	}
	if isSlice = ft.Kind() == reflect.Slice; isSlice {
		ft = ft.Elem()
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
	}
	if ft.Kind() == reflect.Array {
		ft = ft.Elem()
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
	}
	if isStruct = ft.Kind() == reflect.Struct; !isStruct {
		if c.converter(ft) == nil && builtinConverters[ft.Kind()] == nil {
			return
		}
	}

	info.fields = append(info.fields, &fieldInfo{
		typ:      field.Type,
		name:     field.Name,
		ss:       isSlice && isStruct,
		alias:    alias,
		anon:     field.Anonymous,
		required: options.Contains("required"),
	})
}

// converter 返回类型t的转换函数
func (c *cache) converter(t reflect.Type) Converter {
	return c.regconv[t]
}

type structInfo struct {
	fields []*fieldInfo
}

func (i *structInfo) get(alias string) *fieldInfo {
	for _, field := range i.fields {
		if strings.EqualFold(field.alias, alias) {
			return field
		}
	}
	return nil
}

type pathPart struct {
	field *fieldInfo
	path  []string // path to the field: walks structs using field names.
	index int      // struct index in slices of structs.
}

type fieldInfo struct {
	typ      reflect.Type
	name     string // field name in the struct.
	ss       bool   // true if this is a slice of structs.
	alias    string
	anon     bool // is an embedded field
	required bool // tag option
}

// fieldAlias parses a field tag to get a field alias.
func fieldAlias(field reflect.StructField, tagName string) (alias string, options tagOptions) {
	if tag := field.Tag.Get(tagName); tag != "" {
		alias, options = parseTag(tag)
	}
	if alias == "" {
		alias = field.Name
	}
	return alias, options
}

// tagOptions is the string following a comma in a struct field's tag, or
// the empty string. It does not include the leading comma.
type tagOptions []string

func parseTag(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	return s[0], s[1:]
}

// Contains checks whether the tagOptions contains the specified option.
func (o tagOptions) Contains(option string) bool {
	for _, s := range o {
		if s == option {
			return true
		}
	}

	return false
}
