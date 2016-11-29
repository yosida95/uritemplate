// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"reflect"
	"testing"
)

var (
	testNewCases = []struct {
		raw  string
		tmpl *Template
		err  error
	}{
		{
			raw: "http://www.example.com/foo{?query,number}",
			tmpl: &Template{
				exprs: []expression{
					expNoop("http://www.example.com/foo"),
					&expQuestion{
						vars: []varspec{
							varspec{
								name: "query",
							},
							varspec{
								name: "number",
							},
						},
					},
				},
			},
		},
	}
)

func TestNew(t *testing.T) {
	for _, c := range testNewCases {
		tmpl, err := New(c.raw)
		if err != nil {
			if c.err != err {
				t.Errorf("expected: %#v, got %#v", c.err, err)
			}
			continue
		}
		if !reflect.DeepEqual(c.tmpl, tmpl) {
			t.Errorf("expected: %#v, got %#v", c.tmpl, tmpl)
		}
	}
}

func BenchmarkNew(b *testing.B) {
	c := testNewCases[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(c.raw)
	}
}
