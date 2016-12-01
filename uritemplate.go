// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"bytes"
	"log"
)

var (
	debug = debugT(false)
)

type debugT bool

func (t debugT) Printf(format string, v ...interface{}) {
	if t {
		log.Printf(format, v...)
	}
}

// Template represents an URI Template.
type Template struct {
	exprs []template
}

// New parse and construct new Template instance based on the template.
// New returns an error if the template cannot be recognized.
func New(template string) (*Template, error) {
	return (&parser{r: template}).parseURITemplate()
}

// MustNew panics if the template cannot be recognized.
func MustNew(template string) *Template {
	ret, err := New(template)
	if err != nil {
		panic(err)
	}
	return ret
}

// Expand returns an URI reference corresponding t and vars.
func (t *Template) Expand(vars Values) (string, error) {
	w := bytes.Buffer{}
	for i := range t.exprs {
		expr := t.exprs[i]
		if err := expr.expand(&w, vars); err != nil {
			return w.String(), err
		}
	}
	return w.String(), nil
}
