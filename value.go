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
}

// String returns Value that represents string.
func String(v string) Value {
	return valueString(v)
}

type valueString string

// List returns Value that represents list.
func List(v ...string) Value {
	return valueList(v)
}

type valueList []string

// KV returns Value that represents associative list.
// KV panics if len(kv) is not even.
func KV(kv ...string) Value {
	if len(kv)%2 != 0 {
		panic("uritemplate.go: count of the kv must be even number")
	}
	return valueKV(kv)
}

type valueKV []string
