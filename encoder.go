// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"reflect"
)

// An Encoder writes config values to an output stream.
type Encoder struct {
	w              io.Writer
	EncodeFallback func(val interface{}) error
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the config encoding of cto the stream,
func (enc *Encoder) Encode(c interface{}) error {
	var err error
	for i := range reflect.TypeOf(c).NumField() {
		cKey, ok := reflect.StructTag(reflect.TypeOf(c).Field(i).Tag).Lookup("config")
		if !ok {
			continue
		}
		field := reflect.ValueOf(c).Elem().Field(i)
		switch field.Kind() {
		case reflect.Int:
			_, err = fmt.Fprintf(enc.w, "%s = %d\n", cKey, field.Int())
			if err != nil {
				return fmt.Errorf("write %s int: %w", cKey, err)
			}
		case reflect.String:
			_, err = fmt.Fprintf(enc.w, "%s = %s\n", cKey, field.String())
			if err != nil {
				return fmt.Errorf("write %s string: %w", cKey, err)
			}
		case reflect.Bool:
			_, err = fmt.Fprintf(enc.w, "%s = %t\n", cKey, field.Bool())
			if err != nil {
				return fmt.Errorf("write %s bool: %w", cKey, err)
			}
		case reflect.Float64:
			_, err = fmt.Fprintf(enc.w, "%s = %f\n", cKey, field.Float())
			if err != nil {
				return fmt.Errorf("write %s float64: %w", cKey, err)
			}
		case reflect.Uint:
			_, err = fmt.Fprintf(enc.w, "%s = %d\n", cKey, field.Uint())
			if err != nil {
				return fmt.Errorf("write %s uint: %w", cKey, err)
			}

		case reflect.Struct:
			switch val := field.Interface().(type) {
			case image.Rectangle:
				_, err = fmt.Fprintf(enc.w, "%s = %d,%d,%d,%d\n", cKey, val.Min.X, val.Min.Y, val.Max.X, val.Max.Y)
				if err != nil {
					return fmt.Errorf("write %s rectangle: %w", cKey, err)
				}

			case color.RGBA:
				_, err = fmt.Fprintf(enc.w, "%s = %d,%d,%d,%d\n", cKey, val.R, val.G, val.B, val.A)
				if err != nil {
					return fmt.Errorf("write %s color.RGBA: %w", cKey, err)
				}
			default:
				if enc.EncodeFallback != nil {
					err = enc.EncodeFallback(field.Kind())
					if err != nil {
						return fmt.Errorf("write %s struct: %w", cKey, err)
					}
				}

				return fmt.Errorf("unknown struct type %s", field.Kind())
			}
		default:
			if enc.EncodeFallback != nil {
				err = enc.EncodeFallback(field.Interface())
				if err != nil {
					return fmt.Errorf("write %s type: %w", cKey, err)
				}
			}
			return fmt.Errorf("unknown type %s", field.Kind())
		}
	}

	return nil
}
