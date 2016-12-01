// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"bytes"
)

type template interface {
	expand(*bytes.Buffer, Values) error
}

type literals string

func (l literals) expand(w *bytes.Buffer, _ Values) error {
	w.Write([]byte(l))
	return nil
}

type varspec struct {
	name    string
	maxlen  int
	explode bool
}

type expression struct {
	vars   []varspec
	first  string
	sep    string
	named  bool
	ifemp  string
	escape escapeFunc
}

func (e *expression) expand(w *bytes.Buffer, values Values) error {
	first := true
	for i := 0; i < len(e.vars); i++ {
		varspec := e.vars[i]
		value := values.Get(varspec.name)
		if value == nil || !value.defined() {
			continue
		}
		if first {
			w.WriteString(e.first)
			first = false
		} else {
			w.WriteString(e.sep)
		}

		value.expand(varspec.maxlen, func(k string, v string, first bool) error {
			if !first {
				if varspec.explode {
					w.WriteString(e.sep)
				} else {
					w.Write([]byte{','})
				}
			}
			if k == "" && e.named && (first || varspec.explode) {
				w.WriteString(varspec.name)
				if value.empty() {
					w.WriteString(e.ifemp)
					return nil
				}
				w.Write([]byte{'='})
			}
			if k != "" {
				if e.named && varspec.explode {
					w.WriteString(k)
				} else {
					if err := e.escape(w, k); err != nil {
						return err
					}
				}
				if e.named || varspec.explode {
					w.Write([]byte{'='})
				} else {
					w.Write([]byte{','})
				}
			}

			return e.escape(w, v)
		})
	}
	return nil
}
