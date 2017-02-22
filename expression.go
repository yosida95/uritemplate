// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"bytes"
	"regexp"
	"strconv"
)

type template interface {
	expand(*bytes.Buffer, Values) error
	regexp(*bytes.Buffer)
}

type literals string

func (l literals) expand(w *bytes.Buffer, _ Values) error {
	w.Write([]byte(l))
	return nil
}

func (l literals) regexp(b *bytes.Buffer) {
	b.WriteByte('(')
	b.WriteString(regexp.QuoteMeta(string(l)))
	b.WriteByte(')')
}

type varspec struct {
	name    string
	maxlen  int
	explode bool
}

type expression struct {
	vars   []varspec
	op     int
	first  string
	sep    string
	named  bool
	ifemp  string
	escape escapeFunc
	allow  runeClass
}

func (e *expression) init() {
	switch e.op {
	case parseOpSimple:
		e.sep = ","
		e.escape = escapeExceptU
		e.allow = runeClassU
	case parseOpPlus:
		e.sep = ","
		e.escape = escapeExceptUR
		e.allow = runeClassUR
	case parseOpCrosshatch:
		e.first = "#"
		e.sep = ","
		e.escape = escapeExceptUR
		e.allow = runeClassUR
	case parseOpDot:
		e.first = "."
		e.sep = "."
		e.escape = escapeExceptU
		e.allow = runeClassU
	case parseOpSlash:
		e.first = "/"
		e.sep = "/"
		e.escape = escapeExceptU
		e.allow = runeClassU
	case parseOpSemicolon:
		e.first = ";"
		e.sep = ";"
		e.named = true
		e.escape = escapeExceptU
		e.allow = runeClassU
	case parseOpQuestion:
		e.first = "?"
		e.sep = "&"
		e.named = true
		e.ifemp = "="
		e.escape = escapeExceptU
		e.allow = runeClassU
	case parseOpAmpersand:
		e.first = "&"
		e.sep = "&"
		e.named = true
		e.ifemp = "="
		e.escape = escapeExceptU
		e.allow = runeClassU
	}
}

func (e *expression) expand(w *bytes.Buffer, values Values) error {
	first := true
	for _, varspec := range e.vars {
		value := values.Get(varspec.name)
		if !value.Valid() {
			continue
		}

		if first {
			w.WriteString(e.first)
			first = false
		} else {
			w.WriteString(e.sep)
		}

		if err := value.expand(w, varspec, e); err != nil {
			return err
		}

	}
	return nil
}

func (e *expression) regexp(b *bytes.Buffer) {
	if e.first != "" {
		b.WriteString("(?:") // $1
		b.WriteString(regexp.QuoteMeta(e.first))
	}
	b.WriteByte('(') // $2
	runeClassToRegexp(b, e.allow, e.named || e.vars[0].explode)
	if len(e.vars) > 1 || e.vars[0].explode {
		max := len(e.vars) - 1
		for i := 0; i < len(e.vars); i++ {
			if e.vars[i].explode {
				max = -1
				break
			}
		}

		b.WriteString("(?:") // $3
		b.WriteString(regexp.QuoteMeta(e.sep))
		runeClassToRegexp(b, e.allow, e.named || max < 0)
		b.WriteByte(')') // $3
		if max > 0 {
			b.WriteString("{0,")
			b.WriteString(strconv.Itoa(max))
			b.WriteByte('}')
		} else {
			b.WriteByte('*')
		}
	}
	b.WriteByte(')') // $2
	if e.first != "" {
		b.WriteByte(')') // $1
	}
	b.WriteByte('?')
}

func runeClassToRegexp(b *bytes.Buffer, class runeClass, named bool) {
	b.WriteString("(?:(?:[")
	if class&runeClassR == 0 {
		b.WriteString(`\x2c`)
		if named {
			b.WriteString(`\x3d`)
		}
	}
	if class&runeClassU == runeClassU {
		b.WriteString(reUnreserved)
	}
	if class&runeClassR == runeClassR {
		b.WriteString(reReserved)
	}
	b.WriteString("]")
	b.WriteString("|%[[:xdigit:]][[:xdigit:]]")
	b.WriteString(")*)")
}
