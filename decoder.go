// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Decoder reads config values from an input stream.
type Decoder struct {
	r                     io.Reader
	DecodeFallback        func(val interface{}) error
	DefaultDecodeFallback func(val interface{}) error
	FailOnUnknownKey      bool
	FailOnMissingKey      bool
}

// NewDecoder returns a new decoder that writes to w.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode writes the config decode of c to the stream,
func (dec *Decoder) Decode(c interface{}) error {

	var err error

	err = dec.decodeDefault(c)
	if err != nil {
		return fmt.Errorf("decode default: %w", err)
	}

	err = dec.decode(c)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

func (dec *Decoder) decodeDefault(c interface{}) error {
	var err error
	for i := range reflect.TypeOf(c).NumField() {
		cKey, ok := reflect.StructTag(reflect.TypeOf(c).Field(i).Tag).Lookup("config")
		if !ok {
			continue
		}

		defaultValue := reflect.TypeOf(c).Field(i).Tag.Get("config_default")
		if defaultValue == "" {
			continue
		}

		field := reflect.ValueOf(c).Elem().Field(i)
		switch field.Kind() {
		case reflect.Int:
			val, err := strconv.Atoi(defaultValue)
			if err != nil {
				return fmt.Errorf("parse %s to int: %w", defaultValue, err)
			}
			field.SetInt(int64(val))
		case reflect.String:
			field.SetString(defaultValue)
		case reflect.Bool:
			val, err := strconv.ParseBool(defaultValue)
			if err != nil {
				return fmt.Errorf("parse %s to bool: %w", defaultValue, err)
			}
			field.SetBool(val)
		case reflect.Float64:
			val, err := strconv.ParseFloat(defaultValue, 64)
			if err != nil {
				return fmt.Errorf("parse %s to float64: %w", defaultValue, err)
			}
			field.SetFloat(val)
		case reflect.Float32:
			val, err := strconv.ParseFloat(defaultValue, 32)
			if err != nil {
				return fmt.Errorf("parse %s to float32: %w", defaultValue, err)
			}
			field.SetFloat(val)
		case reflect.Uint:
			val, err := strconv.ParseUint(defaultValue, 10, 0)
			if err != nil {
				return fmt.Errorf("parse %s to uint: %w", defaultValue, err)
			}
			field.SetUint(val)
		case reflect.Uint8:
			val, err := strconv.ParseUint(defaultValue, 10, 8)
			if err != nil {
				return fmt.Errorf("parse %s to uint8: %w", defaultValue, err)
			}
			field.SetUint(val)
		case reflect.Uint16:
			val, err := strconv.ParseUint(defaultValue, 10, 16)
			if err != nil {
				return fmt.Errorf("parse %s to uint16: %w", defaultValue, err)
			}
			field.SetUint(val)
		case reflect.Uint32:
			val, err := strconv.ParseUint(defaultValue, 10, 32)
			if err != nil {
				return fmt.Errorf("parse %s to uint32: %w", defaultValue, err)
			}
			field.SetUint(val)
		case reflect.Uint64:
			val, err := strconv.ParseUint(defaultValue, 10, 64)
			if err != nil {
				return fmt.Errorf("parse %s to uint64: %w", defaultValue, err)
			}
			field.SetUint(val)
		case reflect.Uintptr:
			val, err := strconv.ParseUint(defaultValue, 10, 64)
			if err != nil {
				return fmt.Errorf("parse %s to uintptr: %w", defaultValue, err)
			}
			field.SetUint(val)
		case reflect.Complex64:
			val, err := strconv.ParseComplex(defaultValue, 64)
			if err != nil {
				return fmt.Errorf("parse %s to complex64: %w", defaultValue, err)
			}
			field.SetComplex(val)
		case reflect.Complex128:
			val, err := strconv.ParseComplex(defaultValue, 128)
			if err != nil {
				return fmt.Errorf("parse %s to complex128: %w", defaultValue, err)
			}
			field.SetComplex(val)
		case reflect.Struct:
			switch field.Interface().(type) {
			case image.Rectangle:
				parts := strings.Split(defaultValue, ",")
				if len(parts) != 4 {
					return fmt.Errorf("parse %s to image.Rectangle: invalid number of parts", defaultValue)
				}

				var rect image.Rectangle
				for i := range parts {
					val, err := strconv.Atoi(parts[i])
					if err != nil {
						return fmt.Errorf("parse %s to image.Rectangle: %w", defaultValue, err)
					}

					switch i {
					case 0:
						rect.Min.X = val
					case 1:
						rect.Min.Y = val
					case 2:
						rect.Max.X = val
					case 3:
						rect.Max.Y = val
					}
				}

				field.Set(reflect.ValueOf(rect))
			case color.RGBA:
				parts := strings.Split(defaultValue, ",")
				if len(parts) != 4 {
					return fmt.Errorf("parse %s to color.RGBA: invalid number of parts", defaultValue)
				}

				var rgba color.RGBA
				for i := range parts {
					val, err := strconv.Atoi(parts[i])
					if err != nil {
						return fmt.Errorf("parse %s to color.RGBA: %w", defaultValue, err)
					}

					switch i {
					case 0:
						rgba.R = uint8(val)
					case 1:
						rgba.G = uint8(val)
					case 2:
						rgba.B = uint8(val)
					case 3:
						rgba.A = uint8(val)
					}
				}

				field.Set(reflect.ValueOf(rgba))

			default:
				if dec.DefaultDecodeFallback != nil {
					err = dec.DefaultDecodeFallback(field.Interface())
					if err != nil {
						return fmt.Errorf("read %s struct: %w", cKey, err)
					}
				}
				return fmt.Errorf("unknown struct type %s", field.Kind())
			}
		default:
			if dec.DefaultDecodeFallback != nil {
				err = dec.DefaultDecodeFallback(field.Interface())
				if err != nil {
					return fmt.Errorf("read %s type: %w", cKey, err)
				}
			}

			return fmt.Errorf("unknown type %s", field.Kind())
		}
	}
	return nil
}

func (dec *Decoder) decode(c interface{}) error {

	foundKeys := map[string]bool{}
	knownKeys := []string{}

	for i := range reflect.TypeOf(c).NumField() {
		sKey, ok := reflect.StructTag(reflect.TypeOf(c).Field(i).Tag).Lookup("config")
		if !ok {
			continue
		}

		knownKeys = append(knownKeys, sKey)
	}

	lineNumber := 0
	reader := bufio.NewScanner(dec.r)
	for reader.Scan() {
		line := reader.Text()
		lineNumber++
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) != 2 {
				continue
			}
			cKey := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])

			isKnown := false
			for i := range reflect.TypeOf(c).NumField() {
				sKey, ok := reflect.StructTag(reflect.TypeOf(c).Field(i).Tag).Lookup("config")
				if !ok {
					continue
				}

				if sKey != cKey {
					continue
				}

				field := reflect.ValueOf(c).Elem().Field(i)
				switch field.Kind() {
				case reflect.Bool:
					val, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("line %d parse %s=%s to bool: %w", lineNumber, cKey, value, err)
					}

					field.SetBool(val)
				case reflect.Int:
					val, err := strconv.Atoi(value)
					if err != nil {
						return fmt.Errorf("line %d parse %s=%s to int: %w", lineNumber, cKey, value, err)
					}

					field.SetInt(int64(val))
				case reflect.String:
					field.SetString(value)
				case reflect.Struct:
					switch field.Interface().(type) {
					case image.Rectangle:
						parts := strings.Split(value, ",")
						if len(parts) != 4 {
							return fmt.Errorf("line %d parse %s=%s to image.Rectangle: invalid number of parts", lineNumber, cKey, value)
						}

						var rect image.Rectangle
						for i := range parts {
							val, err := strconv.Atoi(parts[i])
							if err != nil {
								return fmt.Errorf("line %d parse %s=%s to image.Rectangle: %w", lineNumber, cKey, value, err)
							}

							switch i {
							case 0:
								rect.Min.X = val
							case 1:
								rect.Min.Y = val
							case 2:
								rect.Max.X = val
							case 3:
								rect.Max.Y = val
							}
						}

						field.Set(reflect.ValueOf(rect))

					default:
						return fmt.Errorf("line %d unknown struct type %s", lineNumber, field.Kind())
					}
				default:
					return fmt.Errorf("line %d unknown type %s", lineNumber, field.Kind())
				}
				isKnown = true
				foundKeys[cKey] = true
			}

			if !isKnown {
				if dec.FailOnUnknownKey {
					return fmt.Errorf("line %d unknown key %s", lineNumber, cKey)
				}
			}
		}
	}

	if len(foundKeys) != len(knownKeys) {
		for i := range knownKeys {
			if !foundKeys[knownKeys[i]] {
				if dec.FailOnMissingKey {
					return fmt.Errorf("missing key %s", knownKeys[i])
				}
			}
		}
	}
	return nil

}
