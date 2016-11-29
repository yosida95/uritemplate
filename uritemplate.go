// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"bytes"
)

type Value interface {
}

// ValueKV represents a template variable which is typed as a list of
// key-value pairs. Length of the ValueKV instance must be multiples of two.
type ValueKV []string

// ValueList represents a list
type ValueList []string

// Template represents an URI Template.
type Template struct {
	exprs []expression
}

// Expand returns an URI reference corresponding t and vars.
func (t *Template) Expand(vars map[string]Value) (string, error) {
	w := bytes.Buffer{}
	for i := range t.exprs {
		expr := t.exprs[i]
		if err := expr.expand(&w, vars); err != nil {
			return w.String(), err
		}
	}
	return w.String(), nil
}
