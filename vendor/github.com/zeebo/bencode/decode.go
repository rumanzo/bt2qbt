package bencode

import (
	"bufio"
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

var (
	reflectByteSliceType = reflect.TypeOf([]byte(nil))
	reflectStringType    = reflect.TypeOf("")
)

// Unmarshaler is the interface implemented by types that can unmarshal
// a bencode description of themselves.
// The input can be assumed to be a valid encoding of a bencode value.
// UnmarshalBencode must copy the bencode data if it wishes to retain the data after returning.
type Unmarshaler interface {
	UnmarshalBencode([]byte) error
}

// A Decoder reads and decodes bencoded data from an input stream.
type Decoder struct {
	r             *bufio.Reader
	raw           bool
	buf           []byte
	n             int
	failUnordered bool
}

// SetFailOnUnorderedKeys will cause the decoder to fail when encountering
// unordered keys. The default is to not fail.
func (d *Decoder) SetFailOnUnorderedKeys(fail bool) {
	d.failUnordered = fail
}

// BytesParsed returns the number of bytes that have actually been parsed
func (d *Decoder) BytesParsed() int {
	return d.n
}

// read also writes into the buffer when d.raw is set.
func (d *Decoder) read(p []byte) (n int, err error) {
	n, err = d.r.Read(p)
	if d.raw {
		d.buf = append(d.buf, p[:n]...)
	}
	d.n += n
	return
}

// readBytes also writes into the buffer when d.raw is set.
func (d *Decoder) readBytes(delim byte) (line []byte, err error) {
	line, err = d.r.ReadBytes(delim)
	if d.raw {
		d.buf = append(d.buf, line...)
	}
	d.n += len(line)
	return
}

// readByte also writes into the buffer when d.raw is set.
func (d *Decoder) readByte() (b byte, err error) {
	b, err = d.r.ReadByte()
	if d.raw {
		d.buf = append(d.buf, b)
	}
	d.n++
	return
}

// readFull also writes into the buffer when d.raw is set.
func (d *Decoder) readFull(p []byte) (n int, err error) {
	n, err = io.ReadFull(d.r, p)
	if d.raw {
		d.buf = append(d.buf, p[:n]...)
	}
	d.n += n
	return
}

func (d *Decoder) peekByte() (b byte, err error) {
	ch, err := d.r.Peek(1)
	if err != nil {
		return
	}
	b = ch[0]
	return
}

// NewDecoder returns a new decoder that reads from r
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: bufio.NewReader(r)}
}

// Decode reads the bencoded value from its input and stores it in the value pointed to by val.
// Decode allocates maps/slices as necessary with the following additional rules:
// To decode a bencoded value into a nil interface value, the type stored in the interface value is one of:
// 	int64 for bencoded integers
// 	string for bencoded strings
// 	[]interface{} for bencoded lists
// 	map[string]interface{} for bencoded dicts
// To unmarshal bencode into a value implementing the Unmarshaler interface,
// Unmarshal calls that value's UnmarshalBencode method.
// Otherwise, if the value implements encoding.TextUnmarshaler
// and the input is a bencode string, Unmarshal calls that value's
// UnmarshalText method with the decoded form of the string.
func (d *Decoder) Decode(val interface{}) error {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("Unwritable type passed into decode")
	}

	return d.decodeInto(rv)
}

// DecodeString reads the data in the string and stores it into the value pointed to by val.
// Read the docs for Decode for more information.
func DecodeString(in string, val interface{}) error {
	buf := strings.NewReader(in)
	d := NewDecoder(buf)
	return d.Decode(val)
}

// DecodeBytes reads the data in b and stores it into the value pointed to by val.
// Read the docs for Decode for more information.
func DecodeBytes(b []byte, val interface{}) error {
	r := bytes.NewReader(b)
	d := NewDecoder(r)
	return d.Decode(val)
}

func indirect(v reflect.Value, alloc bool) reflect.Value {
	for {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				if !alloc {
					return reflect.Value{}
				}
				return v
			}

		case reflect.Ptr:
			if v.IsNil() {
				if !alloc {
					return reflect.Value{}
				}
				v.Set(reflect.New(v.Type().Elem()))
			}

		default:
			return v
		}

		v = v.Elem()
	}
}

func (d *Decoder) decodeInto(val reflect.Value) (err error) {
	var v reflect.Value
	if d.raw {
		v = val
	} else {
		var (
			unmarshaler     Unmarshaler
			textUnmarshaler encoding.TextUnmarshaler
		)
		unmarshaler, textUnmarshaler, v = d.indirect(val)

		// if we're decoding into an Unmarshaler,
		// we pass on the next bencode value to this value instead,
		// so it can decide what to do with it.
		if unmarshaler != nil {
			var x RawMessage
			if err := d.decodeInto(reflect.ValueOf(&x)); err != nil {
				return err
			}
			return unmarshaler.UnmarshalBencode([]byte(x))
		}

		// if we're decoding into an TextUnmarshaler,
		// we'll assume that the bencode value is a string,
		// we decode it as such and pass the result onto the unmarshaler.
		if textUnmarshaler != nil {
			var b []byte
			ref := reflect.ValueOf(&b)
			if err := d.decodeString(reflect.Indirect(ref)); err != nil {
				return err
			}
			return textUnmarshaler.UnmarshalText(b)
		}

		// if we're decoding into a RawMessage set raw to true for the rest of
		// the call stack, and switch out the value with an interface{}.
		if _, ok := v.Interface().(RawMessage); ok {
			v = reflect.Value{} // explicitly make v invalid

			// set d.raw for the lifetime of this function call, and set the raw
			// message when the function is exiting.
			d.buf = d.buf[:0]
			d.raw = true
			defer func() {
				d.raw = false
				v := indirect(val, true)
				v.SetBytes(append([]byte(nil), d.buf...))
			}()
		}
	}

	next, err := d.peekByte()
	if err != nil {
		return
	}

	switch next {
	case 'i':
		err = d.decodeInt(v)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		err = d.decodeString(v)
	case 'l':
		err = d.decodeList(v)
	case 'd':
		err = d.decodeDict(v)
	default:
		err = errors.New("Invalid input")
	}

	return
}

func (d *Decoder) decodeInt(v reflect.Value) error {
	// we need to read an i, some digits, and an e.
	ch, err := d.readByte()
	if err != nil {
		return err
	}
	if ch != 'i' {
		panic("got not an i when peek returned an i")
	}

	line, err := d.readBytes('e')
	if err != nil || d.raw {
		return err
	}

	digits := string(line[:len(line)-1])

	switch v.Kind() {
	default:
		return fmt.Errorf("Cannot store int64 into %s", v.Type())
	case reflect.Interface:
		n, err := strconv.ParseInt(digits, 10, 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(n))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(digits, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(digits, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(n)
	case reflect.Bool:
		n, err := strconv.ParseUint(digits, 10, 64)
		if err != nil {
			return err
		}
		v.SetBool(n != 0)
	}

	return nil
}

func (d *Decoder) decodeString(v reflect.Value) error {
	// read until a colon to get the number of digits to read after
	line, err := d.readBytes(':')
	if err != nil {
		return err
	}

	// parse it into an int for making a slice
	l32, err := strconv.ParseInt(string(line[:len(line)-1]), 10, 32)
	l := int(l32)
	if err != nil {
		return err
	}
	if l < 0 {
		return fmt.Errorf("invalid negative string length: %d", l)
	}

	// read exactly l bytes out and make our string
	buf := make([]byte, l)
	_, err = d.readFull(buf)
	if err != nil || d.raw {
		return err
	}

	switch v.Kind() {
	default:
		return fmt.Errorf("Cannot store string into %s", v.Type())
	case reflect.Slice:
		if v.Type() != reflectByteSliceType {
			return fmt.Errorf("Cannot store string into %s", v.Type())
		}
		v.SetBytes(buf)
	case reflect.String:
		v.SetString(string(buf))
	case reflect.Interface:
		v.Set(reflect.ValueOf(string(buf)))
	}
	return nil
}

func (d *Decoder) decodeList(v reflect.Value) error {
	if !d.raw {
		// if we have an interface, just put a []interface{} in it!
		if v.Kind() == reflect.Interface {
			var x []interface{}
			defer func(p reflect.Value) { p.Set(v) }(v)
			v = reflect.ValueOf(&x).Elem()
		}

		if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
			return fmt.Errorf("Cant store a []interface{} into %s", v.Type())
		}
	}

	// read out the l that prefixes the list
	ch, err := d.readByte()
	if err != nil {
		return err
	}
	if ch != 'l' {
		panic("got something other than a list head after a peek")
	}

	// if we're decoding in raw mode,
	// we only want to read into the buffer,
	// without actually parsing any values
	if d.raw {
		var ch byte
		for {
			// peek for the end token and read it out
			ch, err = d.peekByte()
			if err != nil {
				return err
			}
			if ch == 'e' {
				_, err = d.readByte() // consume the end
				return err
			}

			// decode the next value
			err = d.decodeInto(v)
			if err != nil {
				return err
			}
		}
	}

	for i := 0; ; i++ {
		// peek for the end token and read it out
		ch, err := d.peekByte()
		if err != nil {
			return err
		}
		switch ch {
		case 'e':
			_, err := d.readByte() // consume the end
			return err
		}

		// grow it if required
		if i >= v.Cap() && v.IsValid() {
			newcap := v.Cap() + v.Cap()/2
			if newcap < 4 {
				newcap = 4
			}
			newv := reflect.MakeSlice(v.Type(), v.Len(), newcap)
			reflect.Copy(newv, v)
			v.Set(newv)
		}

		// reslice into cap (its a slice now since it had to have grown)
		if i >= v.Len() && v.IsValid() {
			v.SetLen(i + 1)
		}

		// decode a value into the index
		if err := d.decodeInto(v.Index(i)); err != nil {
			return err
		}
	}
}

func (d *Decoder) decodeDict(v reflect.Value) error {
	// if we have an interface{}, just put a map[string]interface{} in it!
	if !d.raw && v.Kind() == reflect.Interface {
		var x map[string]interface{}
		defer func(p reflect.Value) { p.Set(v) }(v)
		v = reflect.ValueOf(&x).Elem()
	}

	// consume the head token
	ch, err := d.readByte()
	if err != nil {
		return err
	}
	if ch != 'd' {
		panic("got an incorrect token when it was checked already")
	}

	if d.raw {
		// if we're decoding in raw mode,
		// we only want to read into the buffer,
		// without actually parsing any values
		for {
			// peek the next value type
			ch, err := d.peekByte()
			if err != nil {
				return err
			}
			if ch == 'e' {
				_, err = d.readByte() // consume the end token
				return err
			}

			err = d.decodeString(v)
			if err != nil {
				return err
			}

			err = d.decodeInto(v)
			if err != nil {
				return err
			}
		}
	}

	// check for correct type
	var (
		mapElem reflect.Value
		isMap   bool
		vals    map[string]reflect.Value
	)

	switch v.Kind() {
	case reflect.Map:
		t := v.Type()
		if t.Key() != reflectStringType {
			return fmt.Errorf("Can't store a map[string]interface{} into %s", v.Type())
		}
		if v.IsNil() {
			v.Set(reflect.MakeMap(t))
		}

		isMap = true
		mapElem = reflect.New(t.Elem()).Elem()
	case reflect.Struct:
		vals = make(map[string]reflect.Value)
		setStructValues(vals, v)
	default:
		return fmt.Errorf("Can't store a map[string]interface{} into %s", v.Type())
	}

	var (
		lastKey string
		first   bool = true
	)

	for {
		var subv reflect.Value

		// peek the next value type
		ch, err := d.peekByte()
		if err != nil {
			return err
		}
		if ch == 'e' {
			_, err = d.readByte() // consume the end token
			return err
		}

		// peek the next value we're suppsed to read
		var key string
		if err := d.decodeString(reflect.ValueOf(&key).Elem()); err != nil {
			return err
		}

		// check for unordered keys
		if !first && d.failUnordered && lastKey > key {
			return fmt.Errorf("unordered dictionary: %q appears before %q",
				lastKey, key)
		}
		lastKey, first = key, false

		if isMap {
			mapElem.Set(reflect.Zero(v.Type().Elem()))
			subv = mapElem
		} else {
			subv = vals[key]
		}

		if !subv.IsValid() {
			// if it's invalid, grab but ignore the next value
			var x interface{}
			err := d.decodeInto(reflect.ValueOf(&x).Elem())
			if err != nil {
				return err
			}

			continue
		}

		// subv now contains what we load into
		if err := d.decodeInto(subv); err != nil {
			return err
		}

		if isMap {
			v.SetMapIndex(reflect.ValueOf(key), subv)
		}
	}
}

// indirect walks down v allocating pointers as needed,
// until it gets to a non-pointer.
// if it encounters an (Text)Unmarshaler, indirect stops and returns that.
func (d *Decoder) indirect(v reflect.Value) (Unmarshaler, encoding.TextUnmarshaler, reflect.Value) {
	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() {
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr || v.IsNil() {
			break
		}

		vi := v.Interface()
		if u, ok := vi.(Unmarshaler); ok {
			return u, nil, reflect.Value{}
		}
		if u, ok := vi.(encoding.TextUnmarshaler); ok {
			return nil, u, reflect.Value{}
		}

		v = v.Elem()
	}
	return nil, nil, indirect(v, true)
}

func setStructValues(m map[string]reflect.Value, v reflect.Value) {
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return
	}

	// do embedded fields first
	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		v := v.FieldByIndex(f.Index)
		if f.Anonymous && f.Tag == "" {
			setStructValues(m, v)
		}
	}

	// overwrite embedded struct tags and names
	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		v := v.FieldByIndex(f.Index)
		name, _ := parseTag(f.Tag.Get("bencode"))
		if name == "" {
			if f.Anonymous {
				// it's a struct and its fields have already been added to the map
				continue
			}
			name = f.Name
		}
		if isValidTag(name) {
			m[name] = v
		}
	}
}
