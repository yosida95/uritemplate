// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"bytes"
	"log"
	"regexp"
	"sync"
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

	// protects the rest of fields
	mu       sync.Mutex
	varnames []string
	re       *regexp.Regexp
	prog     *prog
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

// Varnames returns variable names used in the t
func (t *Template) Varnames() []string {
	t.mu.Lock()
	if t.varnames != nil {
		t.mu.Unlock()
		return t.varnames
	}

	reg := map[string]struct{}{}
	t.varnames = []string{}
	for i := range t.exprs {
		expr, ok := t.exprs[i].(*expression)
		if !ok {
			continue
		}
		for i := range expr.vars {
			spec := expr.vars[i]
			if _, ok := reg[spec.name]; ok {
				continue
			}
			reg[spec.name] = struct{}{}
			t.varnames = append(t.varnames, spec.name)
		}
	}
	t.mu.Unlock()

	return t.varnames
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

// Regexp converts the t to regexp and returns compiled *regexp.Regexp.
func (t *Template) Regexp() *regexp.Regexp {
	t.mu.Lock()
	if t.re != nil {
		t.mu.Unlock()
		return t.re
	}

	b := bytes.Buffer{}
	b.WriteByte('^')
	for i := range t.exprs {
		expr := t.exprs[i]
		expr.regexp(&b)
	}
	b.WriteByte('$')
	t.re = regexp.MustCompile(b.String())
	t.mu.Unlock()

	return t.re
}
