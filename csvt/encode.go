package csvt

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Rafael24595/go-collections/collection"
)

const (
	HEADER_ROOT       = string(TBL_HEAD_BASE) + string(TBL_HEAD_ROOT) + string(TBL_HEAD_ROOT)
	HEADER_REGULAR    = string(TBL_HEAD_BASE) + string(TBL_HEAD_BASE) + string(TBL_HEAD_BASE)
	POINTER_INDEX_FIX = 2
)

// MarshalOptions defines the configuration for the CSV serialization process.
// Currently it includes:
//   - Compact: when set to true, dentical serialized rows should be
//              deduplicated by caching and referenced via pointers.
type MarshalOptions struct {
	Compact bool
}

var defaultMarshalOpts = MarshalOptions{
	Compact: true,
}

type csvtSerializer struct {
	opts        MarshalOptions
	tables      map[string][]string
	cache       map[string]string
	nilPointers map[string]string
}

// Marshal encodes the provided value into CSVT using default
// serialization options. The value parameter must be a struct
// or a slice of structs.
//
// Parameters:
//   - value: a struct or slice of structs to serialize
//
// Returns an error if serialization fails at any stage.
//
// Example:
//   var item MyStruct
//   bytes, err := csvt.Marshal(item)
//
//   var items []MyStruct
//   bytes, err := csvt.Marshal(items...)
func Marshal(v ...any) ([]byte, error) {
	return MarshalOpts(defaultMarshalOpts, v...)
}

// MarshalOpts encodes the provided value into CSVT using the given
// serialization options. It behaves the same as Marshal, but allows
// configuring the process via MarshalOptions.
//
// Parameters:
//   - value: a struct or slice of structs to serialize
//   - opts: serialization options (e.g., compact mode)
//
// Returns an error if serialization fails at any stage.
//
// Example:
//   var item MyStruct
//   opts := csvt.MarshalOptions{ Compact: false }
//   bytes, err := csvt.MarshalOpts(opts, item)
func MarshalOpts(opts MarshalOptions, v ...any) ([]byte, error) {
	instance := &csvtSerializer{
		opts:        opts,
		tables:      make(map[string][]string),
		cache:       make(map[string]string),
		nilPointers: make(map[string]string),
	}

	if len(v) == 0 {
		return make([]byte, 0), nil
	}

	rootKey := instance.key(reflect.ValueOf(v[0]))
	if rootKey == "common-array" || rootKey == "common-map" {
		return make([]byte, 0), errors.New("common structures cannot be root")
	}

	for _, e := range v {
		if reflect.ValueOf(e).Kind() == reflect.Pointer {
			return make([]byte, 0), errors.New("not supported yet")
		}
		_, err := instance.serialize(e)
		if err != nil {
			return make([]byte, 0), err
		}
	}

	return []byte(instance.formatTables(rootKey)), nil
}

func (s *csvtSerializer) formatTables(rootKey string) string {
	keys := collection.DictionaryFromMap(s.tables).
		KeysVector().
		Sort(func(a, b string) bool {
			return a == rootKey
		})

	buffer := ""
	for _, k := range keys.Collect() {
		rows := s.tables[k]

		pattern := HEADER_REGULAR
		if k == rootKey {
			pattern = HEADER_ROOT
		}

		buffer += fmt.Sprintf("\n%s %s\n", pattern, k)
		buffer += s.formatRows(rows)
	}

	return buffer
}

func (s *csvtSerializer) formatRows(rows []string) string {
	buffer := ""
	for i, r := range rows {
		index := strconv.FormatInt(int64(i-1), 10)
		if i == 0 {
			index = string(TBL_INDEX_HEAD)
		}
		buffer += fmt.Sprintf("%s%s\n", s.formatIndexArrow(index), r)
	}

	return buffer
}

func (s *csvtSerializer) serialize(entity any) (string, error) {
	rEntity := reflect.ValueOf(entity)

	key := s.key(rEntity)

	if _, exists := s.tables[key]; !exists {
		headers, _ := s.headers(entity)
		s.tables[key] = append(s.tables[key], headers)
		if s.canEmpty(rEntity) {
			item, err := s.makeEmpty(rEntity)
			if err != nil {
				return "", err
			}

			s.tables[key] = append(s.tables[key], item)
			s.nilPointers[key] = s.formatPointerReference(key, len(s.tables[key]))
		}
	}

	if pointer, ok := s.nilPointers[key]; ok && s.isEmpty(rEntity) {
		return pointer, nil
	}

	row, err := s.serializeEntity(entity, rEntity)
	if err != nil {
		return "", err
	}

	if s.opts.Compact {
		if pointer, ok := s.cache[row]; ok {
			return pointer, nil
		}
	}

	s.tables[key] = append(s.tables[key], row)
	pointer := s.formatPointerReference(key, len(s.tables[key]))

	if s.opts.Compact {
		s.cache[row] = pointer
	}

	return pointer, nil
}

func (s *csvtSerializer) canEmpty(entity reflect.Value) bool {
	kind := entity.Kind()
	return kind == reflect.Array || kind == reflect.Chan ||
		kind == reflect.Map || kind == reflect.Slice ||
		kind == reflect.String
}

func (s *csvtSerializer) isEmpty(entity reflect.Value) bool {
	return s.canEmpty(entity) && entity.Len() == 0
}

func (s *csvtSerializer) makeEmpty(entity reflect.Value) (string, error) {
	switch entity.Kind() {
	case reflect.String:
		return "\"\"", nil
	case reflect.Map:
		return string(MAP_CLOSING), nil
	case reflect.Slice, reflect.Array:
		return string(ARR_CLOSING), nil
	default:
		return "", errors.New("the current structure cannot be empty")
	}
}

func (s *csvtSerializer) serializeEntity(entity any, rEntity reflect.Value) (string, error) {
	switch rEntity.Kind() {
	case reflect.Struct:
		return s.serializeStruct(rEntity)
	case reflect.Map:
		return s.serializeMap(rEntity)
	case reflect.Slice, reflect.Array:
		return s.serializeArray(rEntity)
	default:
		return s.serializeObject(entity, rEntity), nil
	}
}

func (s *csvtSerializer) serializeStruct(entity reflect.Value) (string, error) {
	var err error
	strRow := []string{}

	for i := 0; i < entity.NumField(); i++ {
		value := entity.Field(i).Interface()
		if !isCommonType(value) {
			value, err = s.serialize(value)
			if err != nil {
				return "", err
			}
		} else {
			value = sprintf("%v", value)
		}

		strRow = append(strRow, fmt.Sprintf("%v", value))
	}
	return fmt.Sprintf("%v%c", strings.Join(strRow, string(STR_SEPARATOR)), STR_CLOSING), nil
}

func (s *csvtSerializer) serializeMap(entity reflect.Value) (string, error) {
	var err error
	mapRow := []string{}

	for _, k := range entity.MapKeys() {
		key := k.Interface()
		if !isCommonType(key) {
			key, err = s.serialize(key)
			if err != nil {
				return "", err
			}
		} else {
			key = sprintf("%v", key)
		}

		value := entity.MapIndex(k).Interface()
		if !isCommonType(value) {
			value, err = s.serialize(value)
			if err != nil {
				return "", err
			}
		} else {
			value = sprintf("%v", value)
		}

		mapRow = append(mapRow, fmt.Sprintf("%v%c%v", key, MAP_LINKER, value))
	}

	return fmt.Sprintf("%v%c", strings.Join(mapRow, string(MAP_SEPARATOR)), MAP_CLOSING), nil
}

func (s *csvtSerializer) serializeArray(entity reflect.Value) (string, error) {
	var err error
	arrayRow := []string{}

	for i := 0; i < entity.Len(); i++ {
		value := entity.Index(i).Interface()
		if !isCommonType(value) {
			value, err = s.serialize(value)
			if err != nil {
				return "", err
			}
		} else {
			value = sprintf("%v", value)
		}

		arrayRow = append(arrayRow, fmt.Sprintf("%v", value))
	}

	return fmt.Sprintf("%v%c", strings.Join(arrayRow, string(ARR_SEPARATOR)), ARR_CLOSING), nil
}

func (s *csvtSerializer) serializeObject(entity any, rEntity reflect.Value) string {
	if rEntity.Kind() == reflect.String {
		return sprintf("%s", fmt.Sprintf("%v", entity))
	}
	return sprintf("%v", entity)
}

func (s *csvtSerializer) key(val reflect.Value) string {
	switch val.Kind() {
	case reflect.Map:
		return "common-map"
	case reflect.Slice, reflect.Array:
		return "common-array"
	default:
		typ := val.Type()
		return fmt.Sprintf("%s&%s", typ.Name(), s.sha1Identifier(typ.PkgPath()))
	}
}

func (s *csvtSerializer) headers(value any) (string, bool) {
	val := reflect.ValueOf(value)
	typ := val.Type()

	headers := []string{}

	if val.Kind() != reflect.Struct {
		return "", false
	}

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i).Name
		csvTag := typ.Field(i).Tag.Get("csv")
		if csvTag != "" {
			field = csvTag
		}

		headers = append(headers, field)
	}

	return strings.Join(headers, string(HEA_SEPARATOR)), true
}

func (s *csvtSerializer) formatIndexArrow(index string) string {
	return fmt.Sprintf("%v-> ", index)
}

func (s *csvtSerializer) formatPointerReference(key string, position int) string {
	return fmt.Sprintf("$%s_%v", key, position-POINTER_INDEX_FIX)
}

func (s csvtSerializer) sha1Identifier(input string) string {
	hash := sha1.New()
	hash.Write([]byte(input))
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}

func sprintf(pattern string, values ...any) string {
	for i, v := range values {
		switch v := v.(type) {
		case string:
			fixed := strings.ReplaceAll(v, "\"", "\\'")
			fixed = strings.ReplaceAll(fixed, "\\n", "\\\\n")
			fixed = strings.ReplaceAll(fixed, "\n", "\\n")
			values[i] = fmt.Sprintf("\"%v\"", fixed)
		}
	}
	return fmt.Sprintf(pattern, values...)
}

func isCommonType(value interface{}) bool {
	switch value.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return true
	default:
		return false
	}
}
