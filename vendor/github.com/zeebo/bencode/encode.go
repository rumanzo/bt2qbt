package bencode

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"reflect"
	"sort"
)

type sortValues []reflect.Value

func (p sortValues) Len() int           { return len(p) }
func (p sortValues) Less(i, j int) bool { return p[i].String() < p[j].String() }
func (p sortValues) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Marshaler is the interface implemented by types
// that can marshal themselves into valid bencode.
type Marshaler interface {
	MarshalBencode() ([]byte, error)
}

// An Encoder writes bencoded objects to an output stream.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

// Encode writes the bencoded data of val to its output stream.
// If an encountered value implements the Marshaler interface,
// its MarshalBencode method is called to produce the bencode output for this value.
// If no MarshalBencode method is present but the value implements encoding.TextMarshaler instead,
// its MarshalText method is called, which encodes the result as a bencode string.
// See the documentation for Decode about the conversion of Go values to
// bencoded data.
func (e *Encoder) Encode(val interface{}) error {
	return encodeValue(e.w, reflect.ValueOf(val))
}

// EncodeString returns the bencoded data of val as a string.
func EncodeString(val interface{}) (string, error) {
	buf := new(bytes.Buffer)
	e := NewEncoder(buf)
	if err := e.Encode(val); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// EncodeBytes returns the bencoded data of val as a slice of bytes.
func EncodeBytes(val interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	e := NewEncoder(buf)
	if err := e.Encode(val); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func isNilValue(v reflect.Value) bool {
	return (v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr) &&
		v.IsNil()
}

func encodeValue(w io.Writer, val reflect.Value) error {
	marshaler, textMarshaler, v := indirectEncodeValue(val)

	// marshal a type using the Marshaler type
	// if it implements that interface.
	if marshaler != nil {
		bytes, err := marshaler.MarshalBencode()
		if err != nil {
			return err
		}

		_, err = w.Write(bytes)
		return err
	}

	// marshal a type using the TextMarshaler type
	// if it implements that interface.
	if textMarshaler != nil {
		bytes, err := textMarshaler.MarshalText()
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(w, "%d:%s", len(bytes), bytes)
		return err
	}

	// if indirection returns us an invalid value that means there was a nil
	// pointer in the path somewhere.
	if !v.IsValid() {
		return nil
	}

	// send in a raw message if we have that type
	if rm, ok := v.Interface().(RawMessage); ok {
		_, err := io.Copy(w, bytes.NewReader(rm))
		return err
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, err := fmt.Fprintf(w, "i%de", v.Int())
		return err

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		_, err := fmt.Fprintf(w, "i%de", v.Uint())
		return err

	case reflect.Bool:
		i := 0
		if v.Bool() {
			i = 1
		}
		_, err := fmt.Fprintf(w, "i%de", i)
		return err

	case reflect.String:
		_, err := fmt.Fprintf(w, "%d:%s", len(v.String()), v.String())
		return err

	case reflect.Slice, reflect.Array:
		// handle byte slices like strings
		if byteSlice, ok := val.Interface().([]byte); ok {
			_, err := fmt.Fprintf(w, "%d:", len(byteSlice))

			if err == nil {
				_, err = w.Write(byteSlice)
			}

			return err
		}

		if _, err := fmt.Fprint(w, "l"); err != nil {
			return err
		}

		for i := 0; i < v.Len(); i++ {
			if err := encodeValue(w, v.Index(i)); err != nil {
				return err
			}
		}

		_, err := fmt.Fprint(w, "e")
		return err

	case reflect.Map:
		if _, err := fmt.Fprint(w, "d"); err != nil {
			return err
		}
		var (
			keys sortValues = v.MapKeys()
			mval reflect.Value
		)
		sort.Sort(keys)
		for i := range keys {
			mval = v.MapIndex(keys[i])
			if isNilValue(mval) {
				continue
			}
			if err := encodeValue(w, keys[i]); err != nil {
				return err
			}
			if err := encodeValue(w, mval); err != nil {
				return err
			}
		}
		_, err := fmt.Fprint(w, "e")
		return err

	case reflect.Struct:
		if _, err := fmt.Fprint(w, "d"); err != nil {
			return err
		}

		// add embedded structs to the dictionary
		dict := make(dictionary, 0, v.NumField())
		dict, err := readStruct(dict, v)
		if err != nil {
			return err
		}

		// sort the dictionary by keys
		sort.Sort(dict)

		// encode the dictionary in order
		for _, def := range dict {
			// encode the key
			err := encodeValue(w, reflect.ValueOf(def.key))
			if err != nil {
				return err
			}

			// encode the value
			err = encodeValue(w, def.value)
			if err != nil {
				return err
			}
		}

		_, err = fmt.Fprint(w, "e")
		return err
	}

	return fmt.Errorf("Can't encode type: %s", v.Type())
}

// indirectEncodeValue walks down v allocating pointers as needed,
// until it gets to a non-pointer.
// if it encounters an (Text)Marshaler, indirect stops and returns that.
func indirectEncodeValue(v reflect.Value) (Marshaler, encoding.TextMarshaler, reflect.Value) {
	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		if v.Kind() == reflect.Ptr && v.IsNil() {
			break
		}

		vi := v.Interface()
		if m, ok := vi.(Marshaler); ok {
			return m, nil, reflect.Value{}
		}
		if m, ok := vi.(encoding.TextMarshaler); ok {
			return nil, m, reflect.Value{}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		v = v.Elem()
	}
	return nil, nil, indirect(v, false)
}

type definition struct {
	key   string
	value reflect.Value
}

type dictionary []definition

func (d dictionary) Len() int           { return len(d) }
func (d dictionary) Less(i, j int) bool { return d[i].key < d[j].key }
func (d dictionary) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }

func readStruct(dict dictionary, v reflect.Value) (dictionary, error) {
	t := v.Type()
	var (
		fieldValue reflect.Value
		rkey       string
	)
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i)
		rkey = key.Name
		fieldValue = v.FieldByIndex(key.Index)

		// filter out unexported values etc.
		if !fieldValue.CanInterface() {
			continue
		}

		// filter out nil pointer values
		if isNilValue(fieldValue) {
			continue
		}

		// * Near identical to usage in JSON except with key 'bencode'
		//
		// * Struct values encode as BEncode dictionaries. Each exported
		//   struct field becomes a set in the dictionary unless
		//   - the field's tag is "-", or
		//   - the field is empty and its tag specifies the "omitempty"
		//     option.
		//
		// * The default key string is the struct field name but can be
		//   specified in the struct field's tag value.  The "bencode"
		//   key in struct field's tag value is the key name, followed
		//   by an optional comma and options.
		tagValue := key.Tag.Get("bencode")
		if tagValue != "" {
			// Keys with '-' are omit from output
			if tagValue == "-" {
				continue
			}

			name, options := parseTag(tagValue)
			// Keys with 'omitempty' are omitted if the field is empty
			if options.Contains("omitempty") && isEmptyValue(fieldValue) {
				continue
			}

			// All other values are treated as the key string
			if isValidTag(name) {
				rkey = name
			}
		}

		if key.Anonymous && key.Type.Kind() == reflect.Struct && tagValue == "" {
			var err error
			dict, err = readStruct(dict, fieldValue)
			if err != nil {
				return nil, err
			}
		} else {
			dict = append(dict, definition{rkey, fieldValue})
		}
	}
	return dict, nil
}
