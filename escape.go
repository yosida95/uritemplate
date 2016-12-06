// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	hex = []byte("0123456789ABCDEF")
	// reserved   = gen-delims / sub-delims
	// gen-delims =  ":" / "/" / "?" / "#" / "[" / "]" / "@"
	// sub-delims =  "!" / "$" / "&" / "’" / "(" / ")"
	//            /  "*" / "+" / "," / ";" / "="
	rangeReserved = &unicode.RangeTable{
		R16: []unicode.Range16{
			unicode.Range16{Lo: 0x21, Hi: 0x21, Stride: 1}, // '!'
			unicode.Range16{Lo: 0x23, Hi: 0x24, Stride: 1}, // '#' - '$'
			unicode.Range16{Lo: 0x26, Hi: 0x2C, Stride: 1}, // '&' - ','
			unicode.Range16{Lo: 0x2F, Hi: 0x2F, Stride: 1}, // '/'
			unicode.Range16{Lo: 0x3A, Hi: 0x3B, Stride: 1}, // ':' - ';'
			unicode.Range16{Lo: 0x3D, Hi: 0x3D, Stride: 1}, // '='
			unicode.Range16{Lo: 0x3F, Hi: 0x40, Stride: 1}, // '?' - '@'
			unicode.Range16{Lo: 0x5B, Hi: 0x5B, Stride: 1}, // '['
			unicode.Range16{Lo: 0x5D, Hi: 0x5D, Stride: 1}, // ']'
		},
		LatinOffset: 9,
	}
	reReserved = `\x21\x23\x24\x26-\x2c\x2f\x3a\x3b\x3d\x3f\x40\x5b\x5d`
	// ALPHA      = %x41-5A / %x61-7A
	// DIGIT      = %x30-39
	// unreserved = ALPHA / DIGIT / "-" / "." / "_" / "~"
	rangeUnreserved = &unicode.RangeTable{
		R16: []unicode.Range16{
			unicode.Range16{Lo: 0x2D, Hi: 0x2E, Stride: 1}, // '-' - '.'
			unicode.Range16{Lo: 0x30, Hi: 0x39, Stride: 1}, // '0' - '9'
			unicode.Range16{Lo: 0x41, Hi: 0x5A, Stride: 1}, // 'A' - 'Z'
			unicode.Range16{Lo: 0x5F, Hi: 0x5F, Stride: 1}, // '_'
			unicode.Range16{Lo: 0x61, Hi: 0x7A, Stride: 1}, // 'a' - 'z'
			unicode.Range16{Lo: 0x7E, Hi: 0x7E, Stride: 1}, // '~'
		},
	}
	reUnreserved = `\x2d\x2e\x30-\x39\x41-\x5a\x5f\x61-\x7a\x7e`
)

type runeClass uint8

const (
	runeClassU runeClass = 1 << iota
	runeClassR
	runeClassLast

	runeClassUR = runeClassU | runeClassR
)

var runeClassNames = []string{
	"U",
	"R",
}

func (rc runeClass) String() string {
	ret := make([]string, len(runeClassNames))
	var i int
	for j := uint(0); ; j++ {
		test := runeClass(1) << j
		if test >= runeClassLast {
			break
		}
		if rc&test == test {
			ret[i] = runeClassNames[j]
			i++
		}
	}
	return strings.Join(ret[:i], "+")
}

func pctEncode(w *bytes.Buffer, r rune) {
	if s := r >> 24 & 0xff; s > 0 {
		w.Write([]byte{'%', hex[s/16], hex[s%16]})
	}
	if s := r >> 16 & 0xff; s > 0 {
		w.Write([]byte{'%', hex[s/16], hex[s%16]})
	}
	if s := r >> 8 & 0xff; s > 0 {
		w.Write([]byte{'%', hex[s/16], hex[s%16]})
	}
	if s := r & 0xff; s > 0 {
		w.Write([]byte{'%', hex[s/16], hex[s%16]})
	}
}

type escapeFunc func(*bytes.Buffer, string) error

func escapeLiteral(w *bytes.Buffer, v string) error {
	w.WriteString(v)
	return nil
}

func escapeExceptU(w *bytes.Buffer, v string) error {
	for i := 0; i < len(v); {
		r, size := utf8.DecodeRuneInString(v[i:])
		if r == utf8.RuneError {
			return errorf(i, "invalid encoding")
		}
		if unicode.Is(rangeUnreserved, r) {
			w.WriteRune(r)
		} else {
			pctEncode(w, r)
		}
		i += size
	}
	return nil
}

func escapeExceptUR(w *bytes.Buffer, v string) error {
	for i := 0; i < len(v); {
		r, size := utf8.DecodeRuneInString(v[i:])
		if r == utf8.RuneError {
			return errorf(i, "invalid encoding")
		}
		// TODO(yosida95): is pct-encoded triplets allowed here?
		if unicode.In(r, rangeUnreserved, rangeReserved) {
			w.WriteRune(r)
		} else {
			pctEncode(w, r)
		}
		i += size
	}
	return nil
}
