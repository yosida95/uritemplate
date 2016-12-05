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
	for _, varspec := range e.vars {
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

		if err := value.expand(w, varspec, e); err != nil {
			return err
		}

	}
	return nil
}
