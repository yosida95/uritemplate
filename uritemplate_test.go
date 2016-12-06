// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"fmt"
	"testing"
)

func Example() {
	tmpl := MustNew("https://example.com/dictionary/{term:1}/{term}")

	vars := Values{}
	vars.Set("term", String("cat"))
	ret, err := tmpl.Expand(vars)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret)

	// Output:
	// https://example.com/dictionary/c/cat
}

var (
	testTemplateCases = []struct {
		raw      string
		expected string
	}{
		// below cases are quoted from the RFC
		// http://tools.ietf.org/html/rfc6570
		// § 3.2.1
		{"{count}", "one,two,three"},
		{"{count*}", "one,two,three"},
		{"{/count}", "/one,two,three"},
		{"{/count*}", "/one/two/three"},
		{"{;count}", ";count=one,two,three"},
		{"{;count*}", ";count=one;count=two;count=three"},
		{"{?count}", "?count=one,two,three"},
		{"{?count*}", "?count=one&count=two&count=three"},
		{"{&count*}", "&count=one&count=two&count=three"},
		// § 3.2.2
		{"{var}", "value"},
		{"{hello}", "Hello%20World%21"},
		{"{half}", "50%25"},
		{"O{empty}X", "OX"},
		{"O{undef}X", "OX"},
		{"{x,y}", "1024,768"},
		{"{x,hello,y}", "1024,Hello%20World%21,768"},
		{"?{x,empty}", "?1024,"},
		{"?{x,undef}", "?1024"},
		{"?{undef,y}", "?768"},
		{"{var:3}", "val"},
		{"{var:30}", "value"},
		{"{list}", "red,green,blue"},
		{"{list*}", "red,green,blue"},
		{"{keys}", "semi,%3B,dot,.,comma,%2C"},
		{"{keys*}", "semi=%3B,dot=.,comma=%2C"},
		// § 3.2.3
		{"{+var}", "value"},
		{"{+hello}", "Hello%20World!"},
		{"{+half}", "50%25"},
		{"{base}index", "http%3A%2F%2Fexample.com%2Fhome%2Findex"},
		{"{+base}index", "http://example.com/home/index"},
		{"O{+empty}X", "OX"},
		{"O{+undef}X", "OX"},
		{"{+path}/here", "/foo/bar/here"},
		{"here?ref={+path}", "here?ref=/foo/bar"},
		{"up{+path}{var}/here", "up/foo/barvalue/here"},
		{"{+x,hello,y}", "1024,Hello%20World!,768"},
		{"{+path,x}/here", "/foo/bar,1024/here"},
		{"{+path:6}/here", "/foo/b/here"},
		{"{+list}", "red,green,blue"},
		{"{+list*}", "red,green,blue"},
		{"{+keys}", "semi,;,dot,.,comma,,"},
		{"{+keys*}", "semi=;,dot=.,comma=,"},
		// § 3.2.4
		{"{#var}", "#value"},
		{"{#hello}", "#Hello%20World!"},
		{"{#half}", "#50%25"},
		{"foo{#empty}", "foo#"},
		{"foo{#undef}", "foo"},
		{"{#x,hello,y}", "#1024,Hello%20World!,768"},
		{"{#path,x}/here", "#/foo/bar,1024/here"},
		{"{#path:6}/here", "#/foo/b/here"},
		{"{#list}", "#red,green,blue"},
		{"{#list*}", "#red,green,blue"},
		{"{#keys}", "#semi,;,dot,.,comma,,"},
		{"{#keys*}", "#semi=;,dot=.,comma=,"},
		// § 3.2.5
		{"{.who}", ".fred"},
		{"{.who,who}", ".fred.fred"},
		{"{.half,who}", ".50%25.fred"},
		{"www{.dom*}", "www.example.com"},
		{"X{.var}", "X.value"},
		{"X{.empty}", "X."},
		{"X{.undef}", "X"},
		{"X{.var:3}", "X.val"},
		{"X{.list}", "X.red,green,blue"},
		{"X{.list*}", "X.red.green.blue"},
		{"X{.keys}", "X.semi,%3B,dot,.,comma,%2C"},
		{"X{.keys*}", "X.semi=%3B.dot=..comma=%2C"},
		{"X{.empty_keys}", "X"},
		{"X{.empty_keys*}", "X"},
		// § 3.2.6
		{"{/who}", "/fred"},
		{"{/who,who}", "/fred/fred"},
		{"{/half,who}", "/50%25/fred"},
		{"{/who,dub}", "/fred/me%2Ftoo"},
		{"{/var}", "/value"},
		{"{/var,empty}", "/value/"},
		{"{/var,undef}", "/value"},
		{"{/var,x}/here", "/value/1024/here"},
		{"{/var:1,var}", "/v/value"},
		{"{/list}", "/red,green,blue"},
		{"{/list*}", "/red/green/blue"},
		{"{/list*,path:4}", "/red/green/blue/%2Ffoo"},
		{"{/keys}", "/semi,%3B,dot,.,comma,%2C"},
		{"{/keys*}", "/semi=%3B/dot=./comma=%2C"},
		// § 3.2.7
		{"{;who}", ";who=fred"},
		{"{;half}", ";half=50%25"},
		{"{;empty}", ";empty"},
		{"{;v,empty,who}", ";v=6;empty;who=fred"},
		{"{;v,bar,who}", ";v=6;who=fred"},
		{"{;x,y}", ";x=1024;y=768"},
		{"{;x,y,empty}", ";x=1024;y=768;empty"},
		{"{;x,y,undef}", ";x=1024;y=768"},
		{"{;hello:5}", ";hello=Hello"},
		{"{;list}", ";list=red,green,blue"},
		{"{;list*}", ";list=red;list=green;list=blue"},
		{"{;keys}", ";keys=semi,%3B,dot,.,comma,%2C"},
		{"{;keys*}", ";semi=%3B;dot=.;comma=%2C"},
		// § 3.2.8
		{"{?who}", "?who=fred"},
		{"{?half}", "?half=50%25"},
		{"{?x,y}", "?x=1024&y=768"},
		{"{?x,y,empty}", "?x=1024&y=768&empty="},
		{"{?x,y,undef}", "?x=1024&y=768"},
		{"{?var:3}", "?var=val"},
		{"{?list}", "?list=red,green,blue"},
		{"{?list*}", "?list=red&list=green&list=blue"},
		{"{?keys}", "?keys=semi,%3B,dot,.,comma,%2C"},
		{"{?keys*}", "?semi=%3B&dot=.&comma=%2C"},
		// § 3.2.9
		{"{&who}", "&who=fred"},
		{"{&half}", "&half=50%25"},
		{"?fixed=yes{&x}", "?fixed=yes&x=1024"},
		{"{&x,y,empty}", "&x=1024&y=768&empty="},
		{"{&x,y,undef}", "&x=1024&y=768"},
		{"{&var:3}", "&var=val"},
		{"{&list}", "&list=red,green,blue"},
		{"{&list*}", "&list=red&list=green&list=blue"},
		{"{&keys}", "&keys=semi,%3B,dot,.,comma,%2C"},
		{"{&keys*}", "&semi=%3B&dot=.&comma=%2C"},
	}
	testExpressionExpandVarMap = Values{
		"count":      List("one", "two", "three"),
		"dom":        List("example", "com"),
		"dub":        String("me/too"),
		"hello":      String("Hello World!"),
		"half":       String("50%"),
		"var":        String("value"),
		"who":        String("fred"),
		"base":       String("http://example.com/home/"),
		"path":       String("/foo/bar"),
		"list":       List("red", "green", "blue"),
		"keys":       KV("semi", ";", "dot", ".", "comma", ","),
		"v":          String("6"),
		"x":          String("1024"),
		"y":          String("768"),
		"empty":      String(""),
		"empty_keys": KV(),
		// undef is omitted. uritemplate.go treats variables that could not
		// found in the varmap as null.
	}
)

func TestTemplateExpand(t *testing.T) {
	for _, c := range testTemplateCases {
		tmpl, err := New(c.raw)
		if err != nil {
			t.Errorf("unexpected error on %q: %#v", c.raw, err)
			continue
		}

		got, err := tmpl.Expand(testExpressionExpandVarMap)
		if err != nil {
			t.Errorf("unexpected error on %q: %#v", c.raw, err)
			continue
		}
		if c.expected != got {
			t.Errorf("on %q: expected: %#v, got: %#v", c.raw, c.expected, got)
		}
	}
}

func BenchmarkExpressionExpand(b *testing.B) {
	c := testTemplateCases[0]
	tmpl, err := New(c.raw)
	if err != nil {
		b.Errorf("got unexpected error; %#v", err)
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpl.Expand(testExpressionExpandVarMap)
	}
}
