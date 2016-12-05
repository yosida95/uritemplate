// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"bytes"
)

type Values map[string]Value

func (v Values) Set(name string, value Value) {
	v[name] = value // TODO(yosida95): canonicalize pct-encoded in the name
}

func (v Values) Get(name string) Value {
	if v == nil {
		return nil
	}
	return v[name] // TODO(yosida95): canonicalize pct-encoded in the name
}

type Value interface {
	defined() bool
	expand(w *bytes.Buffer, spec varspec, e *expression) error
}

// String returns Value that represents string.
func String(v string) Value {
	return valueString(v)
}

type valueString string

func (v valueString) defined() bool {
	return true
}

func (v valueString) expand(w *bytes.Buffer, spec varspec, exp *expression) error {
	var maxlen int
	if max := len(v); spec.maxlen < 1 || spec.maxlen > max {
		maxlen = max
	} else {
		maxlen = spec.maxlen
	}

	if exp.named {
		w.WriteString(spec.name)
		if v == "" {
			w.WriteString(exp.ifemp)
			return nil
		}
		w.WriteByte('=')
	}
	return exp.escape(w, string(v[:maxlen]))
}

// List returns Value that represents list.
func List(v ...string) Value {
	return valueList(v)
}

type valueList []string

func (v valueList) defined() bool {
	return v != nil && len(v) > 0
}

func (v valueList) expand(w *bytes.Buffer, spec varspec, exp *expression) error {
	var sep string
	if spec.explode {
		sep = exp.sep
	} else {
		sep = ","
	}

	var pre string
	var preifemp string
	if spec.explode && exp.named {
		pre = spec.name + "="
		preifemp = spec.name + exp.ifemp
	}

	if !spec.explode && exp.named {
		w.WriteString(spec.name)
		w.WriteByte('=')
	}

	for i := range v {
		if i > 0 {
			w.WriteString(sep)
		}
		if v[i] == "" {
			w.WriteString(preifemp)
			continue
		}
		w.WriteString(pre)

		if err := exp.escape(w, v[i]); err != nil {
			return err
		}
	}
	return nil
}

// KV returns Value that represents associative list.
// KV panics if len(kv) is not even.
func KV(kv ...string) Value {
	if len(kv)%2 != 0 {
		panic("uritemplate.go: count of the kv must be even number")
	}
	return valueKV(kv)
}

type valueKV []string

func (v valueKV) defined() bool {
	return v != nil && len(v) > 0
}

func (v valueKV) expand(w *bytes.Buffer, spec varspec, exp *expression) error {
	var sep string
	var kvsep string
	if spec.explode {
		sep = exp.sep
		kvsep = "="
	} else {
		sep = ","
		kvsep = ","
	}

	var ifemp string
	var kescape escapeFunc
	if spec.explode && exp.named {
		ifemp = exp.ifemp
		kescape = escapeLiteral
	} else {
		ifemp = ","
		kescape = exp.escape
	}

	if !spec.explode && exp.named {
		w.WriteString(spec.name)
		w.WriteByte('=')
	}

	for i := 0; i < len(v); i += 2 {
		if i > 0 {
			w.WriteString(sep)
		}
		if err := kescape(w, v[i]); err != nil {
			return err
		}
		if v[i+1] == "" {
			w.WriteString(ifemp)
			continue
		}
		w.WriteString(kvsep)

		if err := exp.escape(w, v[i+1]); err != nil {
			return err
		}
	}
	return nil
}
