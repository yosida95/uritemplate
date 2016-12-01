// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

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
	empty() bool
	expand(int, func(k string, v string, first bool) error) error
}

// String returns Value that represents string.
func String(v string) Value {
	return valueString(v)
}

type valueString string

func (v valueString) defined() bool {
	return true
}

func (v valueString) empty() bool {
	return len(v) == 0
}

func (v valueString) expand(maxlen int, found func(string, string, bool) error) error {
	if maxlen < 1 {
		maxlen = len(v)
	}
	if max := len(v); maxlen > max {
		maxlen = max
	}
	return found("", string(v)[:maxlen], true)
}

// List returns Value that represents list.
func List(v ...string) Value {
	return valueList(v)
}

type valueList []string

func (v valueList) defined() bool {
	return len(v) > 0
}

func (v valueList) empty() bool {
	return !v.defined()
}

func (v valueList) expand(_ int, found func(string, string, bool) error) error {
	for i := 0; i < len(v); i++ {
		if err := found("", v[i], i == 0); err != nil {
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
	return len(v) > 0
}

func (v valueKV) empty() bool {
	return !v.defined()
}

func (v valueKV) expand(_ int, found func(string, string, bool) error) error {
	for i := 0; i < len(v); i += 2 {
		if err := found(v[i], v[i+1], i == 0); err != nil {
			return err
		}
	}
	return nil
}
