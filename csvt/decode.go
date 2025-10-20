package csvt

import (
	"errors"
	"fmt"
	"reflect"
)

// UnmarshalOptions defines the configuration for the CSV deserialization process.
// Currently it includes:
//   - strict: when set to true, deserialization will return an error if a field
//             in the target struct does not exist in the CSV tables.
type UnmarshalOptions struct {
	Strict bool
}

var defaultUnmarshalOpts = UnmarshalOptions{
	Strict: false,
}

type csvtDeserializer struct {
	opts   UnmarshalOptions
	tables table
}

// Unmarshal decodes the CSVT data into the provided value using default
// deserialization options. The value parameter must be a pointer to a struct
// or a slice of structs.
//
// Parameters:
//   - data: the CSV-formatted input as a byte slice
//   - value: a pointer to a struct or slice of structs to populate
//
// Returns an error if deserialization fails at any stage.
//
// Example:
//   var result MyStruct
//   err := csvt.Unmarshal(data, &result)
//
//   var list []MyStruct
//   err := csvt.Unmarshal(data, &list)
func Unmarshal[T any](data []byte, value *T) error {
	return UnmarshalOpts(data, value, defaultUnmarshalOpts)
}

// UnmarshalOpts decodes the CSVT data into the provided value using the given
// deserialization options. It behaves the same as Unmarshal, but allows
// configuring the process via UnmarshalOptions.
//
// Parameters:
//   - data: the CSVT-formatted input as a byte slice
//   - value: a pointer to a struct or slice of structs to populate
//   - opts: deserialization options (e.g., strict mode)
//
// Returns an error if deserialization fails at any stage.
//
// Example:
//   var result MyStruct
//   opts := UnmarshalOptions{ Strict: true }
//   err := csvt.UnmarshalOpts(data, &result, opts)
func UnmarshalOpts[T any](data []byte, value *T, opts UnmarshalOptions) error {
	tables, err := newReader().read(data)
	if err != nil {
		return err
	}

	instance := &csvtDeserializer{
		opts:   opts,
		tables: *tables,
	}

	rv := reflect.ValueOf(value).Elem()
	if rv.Kind() != reflect.Slice {
		_, err := instance.deserialize(value, 0)
		return err
	}

	elemType := rv.Type().Elem()

	root, ok := instance.tables.root()
	if !ok {
		return errors.New("root struct is not defined")
	}

	for i := 0; i < root.nodes.Size(); i++ {
		itemPtr := reflect.New(elemType)
		_, err := instance.deserialize(itemPtr.Interface(), i)
		if err != nil {
			return err
		}

		rv.Set(reflect.Append(rv, itemPtr.Elem()))
	}

	return nil
}

func (d *csvtDeserializer) deserialize(value any, index int) (any, error) {
	valPtr := reflect.ValueOf(value)

	if valPtr.Kind() != reflect.Ptr || valPtr.Elem().Kind() != reflect.Struct {
		return nil, errors.New("root struct must be a pointer")
	}

	root, ok := d.tables.root()
	if !ok {
		return nil, errors.New("root struct is not defined")
	}

	group, ok := root.get(index)
	if !ok {
		return nil, errors.New("index does not exists")
	}

	result, err := d.makeElement(value, group)
	if err != nil {
		return nil, err
	}

	return result.Interface(), nil
}

func (d *csvtDeserializer) makeElement(template any, root *group) (reflect.Value, error) {
	element := reflect.ValueOf(template)
	switch element.Kind() {
	case reflect.Struct, reflect.Ptr:
		return d.makeStr(template, root)
	case reflect.Map:
		return d.makeMap(template, root)
	case reflect.Slice, reflect.Array:
		return d.makeArr(template, root)
	default:
		return makeObj(template, root)
	}
}

func (d *csvtDeserializer) makeStr(template any, root *group) (reflect.Value, error) {
	structure := fixStr(template)

	for i := 0; i < structure.NumField(); i++ {
		name := structure.Type().Field(i).Name
		field := structure.FieldByName(name)
		fieldTemplate := field.Interface()

		node, ok := root.findField(name)
		if !ok {
			if d.opts.Strict {
				return reflect.Value{}, MissingField(name)
			}
			continue
		}

		if !isCommonType(fieldTemplate) {
			reference, ok := d.tables.Find(node)
			if !ok {
				return reflect.Value{}, fmt.Errorf("field \"%s\" reference \"%s\" not found", name, node.key())
			}

			element, err := d.makeElement(fieldTemplate, reference)
			if err != nil {
				return reflect.Value{}, err
			}
			
			field.Set(element)

			continue
		}

		if !field.IsValid() {
			return reflect.Value{}, fmt.Errorf("field \"%s\" is not valid", name)
		}
		if !field.CanSet() {
			return reflect.Value{}, fmt.Errorf("field \"%s\" cannot set", name)
		}

		valueRef := reflect.ValueOf(node.value)
		if field.Type() != valueRef.Type() {
			if !valueRef.Type().ConvertibleTo(field.Type()) {
				err := TypeMismatchf(field.Type().Name(), valueRef.Type().Name(), "field \"%s\"", name)
				return reflect.Value{}, err
			}

			valueRef = valueRef.Convert(field.Type())
		}

		field.Set(valueRef)
	}
	return structure, nil
}

func fixStr(value any) reflect.Value {
	element := reflect.ValueOf(value)
	if element.Kind() != reflect.Ptr {
		structureType := reflect.TypeOf(value)
		return reflect.New(structureType).Elem()
	}

	return element.Elem()
}

func (d *csvtDeserializer) makeMap(template any, root *group) (reflect.Value, error) {
	mapType := reflect.TypeOf(template)
	mapElement := reflect.New(mapType).Elem()
	mapKeysType := reflect.TypeOf(mapElement.Interface()).Key()
	mapValuesType := reflect.TypeOf(mapElement.Interface()).Elem()
	mapValuesElement := reflect.New(mapValuesType).Elem()

	mapp := reflect.MakeMap(mapType)

	for _, p := range root.findFields() {
		k := p.Key()
		v := p.Value()

		kv := reflect.ValueOf(k)

		if v.index == -1 {
			vv := reflect.ValueOf(v.value)
			mapp.SetMapIndex(kv.Convert(mapKeysType), vv.Convert(mapValuesType))
			continue
		}

		reference, ok := d.tables.Find(&v)
		if !ok {
			return reflect.Value{}, fmt.Errorf("field \"%s\" is not valid", k)
		}

		value, err := d.makeElement(mapValuesElement.Interface(), reference)
		if err != nil {
			return reflect.Value{}, err
		}

		mapp.SetMapIndex(kv.Convert(mapKeysType), value)
	}
	return mapp, nil
}

func (d *csvtDeserializer) makeArr(template any, root *group) (reflect.Value, error) {
	arrType := reflect.TypeOf(template)
	arrElement := reflect.New(arrType).Elem()
	arrValuesType := reflect.TypeOf(arrElement.Interface()).Elem()
	arrValuesElement := reflect.New(arrValuesType).Elem()

	fields := root.findFields()
	len := len(fields)

	arr := reflect.MakeSlice(arrType, len, len)

	for i, p := range fields {
		v := p.Value()
		if v.index == -1 {
			elem := reflect.ValueOf(v.value)
			if elem.Type() != arrValuesType && !elem.CanConvert(arrValuesType) {
				err := TypeMismatchf(elem.Type().Name(), arrValuesType, "array position \"%d\"", i)
				return reflect.Value{}, err
			}

			arr.Index(i).Set(elem.Convert(arrValuesType))

			continue
		}

		reference, ok := d.tables.Find(&v)
		if !ok {
			return reflect.Value{}, fmt.Errorf("array position \"%d\" reference \"%s\" not found", i, v.key())
		}

		value, err := d.makeElement(arrValuesElement.Interface(), reference)
		if err != nil {
			return reflect.Value{}, err
		}

		if value.Type() != arrValuesType && !value.CanConvert(arrValuesType) {
			err := TypeMismatchf(value.Type().Name(), arrValuesType, "array position \"%d\"", i)
			return reflect.Value{}, err
		}

		arr.Index(i).Set(value.Convert(arrValuesType))
	}
	return arr, nil
}

func makeObj(template any, root *group) (reflect.Value, error) {
	element := reflect.ValueOf(template)

	node, ok := root.findValue()
	if !ok {
		return reflect.Value{}, fmt.Errorf("field category \"%s\" not found", root.category)
	}

	valueRef := reflect.ValueOf(node.value)
	if element.Type().Kind() != valueRef.Type().Kind() {
		err := TypeMismatchf(element.Type().Name(), valueRef.Type().Name(), "field category \"%s\" type", root.category)
		return reflect.Value{}, err
	}

	return valueRef.Convert(element.Type()), nil
}
