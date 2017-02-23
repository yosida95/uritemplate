// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"strconv"
	"strings"
	"testing"
)

func TestMatcher(t *testing.T) {
	for _, c := range testTemplateCases {
		if c.failMatch {
			continue
		}

		tmpl, err := New(c.raw)
		if err != nil {
			t.Errorf("unexpected error on %q: %#v", c.raw, err)
			continue
		}
		m, err := CompileMatcher(tmpl)
		if err != nil {
			t.Errorf("unexpected error on %q: %#v", c.raw, err)
			return
		}

		match := m.Match(c.expected)
		if match == nil {
			t.Errorf("failed to match %q against %q", c.raw, c.expected)
			t.Log(m.prog.String())
			continue
		}
		for name, actual := range match {
			var expected Value
			if semi := strings.Index(name, ":"); semi >= 0 {
				maxlen, _ := strconv.Atoi(name[semi+1:])
				name = name[:semi]
				expected = testExpressionExpandVarMap[name]

				if expected.T != ValueTypeString {
					t.Errorf("failed to match %q against %q", c.raw, c.expected)
					t.Errorf("expected %#q, but got %#q", expected, actual)
					continue
				}
				if v := expected.V[0]; len(v) > maxlen {
					expected.V = []string{v[:maxlen]}
				}
			} else {
				expected = testExpressionExpandVarMap[name]
			}

			if actual.T != expected.T {
				t.Errorf("failed to match %q against %q", c.raw, c.expected)
				t.Errorf("expected %#q, but got %#q", expected, actual)
				continue
			}
			if len(actual.V) != len(expected.V) {
				t.Errorf("failed to match %q against %q", c.raw, c.expected)
				t.Errorf("expected %#q, but got %#q", expected, actual)
				continue
			}
			for i := range actual.V {
				if actual.V[i] != expected.V[i] {
					t.Errorf("failed to match %q against %q", c.raw, c.expected)
					t.Errorf("expected %#q, but got %#q", expected, actual)
					break
				}
			}
		}
	}
}
