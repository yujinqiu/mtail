// Copyright 2011 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package vm

import (
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

type lexerTest struct {
	name   string
	input  string
	tokens []token
}

var lexerTests = []lexerTest{
	{"empty", "", []token{
		token{EOF, "", position{"empty", 0, 0, 0}}}},
	{"spaces", " \t\n", []token{
		token{EOF, "", position{"spaces", 1, 0, 0}}}},
	{"comment", "# comment", []token{
		token{EOF, "", position{"comment", 0, 9, 9}}}},
	{"comment not at col 1", "  # comment", []token{
		token{EOF, "", position{"comment not at col 1", 0, 11, 11}}}},
	{"punctuation", "{}()[],", []token{
		token{LCURLY, "{", position{"punctuation", 0, 0, 0}},
		token{RCURLY, "}", position{"punctuation", 0, 1, 1}},
		token{LPAREN, "(", position{"punctuation", 0, 2, 2}},
		token{RPAREN, ")", position{"punctuation", 0, 3, 3}},
		token{LSQUARE, "[", position{"punctuation", 0, 4, 4}},
		token{RSQUARE, "]", position{"punctuation", 0, 5, 5}},
		token{COMMA, ",", position{"punctuation", 0, 6, 6}},
		token{EOF, "", position{"punctuation", 0, 7, 7}}}},
	{"operators", "- + = ++ += < > <= >= == !=", []token{
		token{MINUS, "-", position{"operators", 0, 0, 0}},
		token{PLUS, "+", position{"operators", 0, 2, 2}},
		token{ASSIGN, "=", position{"operators", 0, 4, 4}},
		token{INC, "++", position{"operators", 0, 6, 7}},
		token{ADD_ASSIGN, "+=", position{"operators", 0, 9, 10}},
		token{LT, "<", position{"operators", 0, 12, 12}},
		token{GT, ">", position{"operators", 0, 14, 14}},
		token{LE, "<=", position{"operators", 0, 16, 17}},
		token{GE, ">=", position{"operators", 0, 19, 20}},
		token{EQ, "==", position{"operators", 0, 22, 23}},
		token{NE, "!=", position{"operators", 0, 25, 26}},
		token{EOF, "", position{"operators", 0, 27, 27}}}},
	{"keywords",
		"counter\ngauge\nas\nby\nhidden\ndef\nnext\nconst\n", []token{
			token{COUNTER, "counter", position{"keywords", 0, 0, 6}},
			token{GAUGE, "gauge", position{"keywords", 1, 0, 4}},
			token{AS, "as", position{"keywords", 2, 0, 1}},
			token{BY, "by", position{"keywords", 3, 0, 1}},
			token{HIDDEN, "hidden", position{"keywords", 4, 0, 5}},
			token{DEF, "def", position{"keywords", 5, 0, 2}},
			token{NEXT, "next", position{"keywords", 6, 0, 3}},
			token{CONST, "const", position{"keywords", 7, 0, 4}},
			token{EOF, "", position{"keywords", 8, 0, 0}}}},
	{"builtins",
		"strptime\ntimestamp\ntolower\nlen\n", []token{
			token{BUILTIN, "strptime", position{"builtins", 0, 0, 7}},
			token{BUILTIN, "timestamp", position{"builtins", 1, 0, 8}},
			token{BUILTIN, "tolower", position{"builtins", 2, 0, 6}},
			token{BUILTIN, "len", position{"builtins", 3, 0, 2}},
			token{EOF, "", position{"builtins", 4, 0, 0}}}},
	{"numeric", "1 23", []token{
		token{NUMERIC, "1", position{"numeric", 0, 0, 0}},
		token{NUMERIC, "23", position{"numeric", 0, 2, 3}},
		token{EOF, "", position{"numeric", 0, 4, 4}}}},
	{"identifier", "a be foo\nquux line-count", []token{
		token{ID, "a", position{"identifier", 0, 0, 0}},
		token{ID, "be", position{"identifier", 0, 2, 3}},
		token{ID, "foo", position{"identifier", 0, 5, 7}},
		token{ID, "quux", position{"identifier", 1, 0, 3}},
		token{ID, "line-count", position{"identifier", 1, 5, 14}},
		token{EOF, "", position{"identifier", 1, 15, 15}}}},
	{"regex", "/asdf/", []token{
		token{REGEX, "asdf", position{"regex", 0, 0, 5}},
		token{EOF, "", position{"regex", 0, 6, 6}}}},
	{"regex with escape", `/asdf\//`, []token{
		token{REGEX, `asdf/`, position{"regex with escape", 0, 0, 7}},
		token{EOF, "", position{"regex with escape", 0, 8, 8}}}},
	{"regex with escape and special char", `/foo\d\//`, []token{
		token{REGEX, `foo\d/`, position{"regex with escape and special char", 0, 0, 8}},
		token{EOF, "", position{"regex with escape and special char", 0, 9, 9}}}},
	{"capref", "$foo", []token{
		token{CAPREF, "foo", position{"capref", 0, 0, 3}},
		token{EOF, "", position{"capref", 0, 4, 4}}}},
	{"numerical capref", "$1", []token{
		token{CAPREF, "1", position{"numerical capref", 0, 0, 1}},
		token{EOF, "", position{"numerical capref", 0, 2, 2}}}},
	{"capref with trailing punc", "$foo,", []token{
		token{CAPREF, "foo", position{"capref with trailing punc", 0, 0, 3}},
		token{COMMA, ",", position{"capref with trailing punc", 0, 4, 4}},
		token{EOF, "", position{"capref with trailing punc", 0, 5, 5}}}},
	{"quoted string", `"asdf"`, []token{
		token{STRING, `asdf`, position{"quoted string", 0, 0, 5}},
		token{EOF, "", position{"quoted string", 0, 6, 6}}}},
	{"escaped quote in quoted string", `"\""`, []token{
		token{STRING, `"`, position{"escaped quote in quoted string", 0, 0, 3}},
		token{EOF, "", position{"escaped quote in quoted string", 0, 4, 4}}}},
	{"decorator", `@foo`, []token{
		token{DECO, "foo", position{"decorator", 0, 0, 3}},
		token{EOF, "", position{"decorator", 0, 4, 4}}}},
	{"large program",
		"/(?P<date>[[:digit:]-\\/ ])/ {\n" +
			"  strptime($date, \"%Y/%m/%d %H:%M:%S\")\n" +
			"  foo++\n" +
			"}", []token{
			token{REGEX, "(?P<date>[[:digit:]-/ ])", position{"large program", 0, 0, 26}},
			token{LCURLY, "{", position{"large program", 0, 28, 28}},
			token{BUILTIN, "strptime", position{"large program", 1, 2, 9}},
			token{LPAREN, "(", position{"large program", 1, 10, 10}},
			token{CAPREF, "date", position{"large program", 1, 11, 15}},
			token{COMMA, ",", position{"large program", 1, 16, 16}},
			token{STRING, "%Y/%m/%d %H:%M:%S", position{"large program", 1, 18, 36}},
			token{RPAREN, ")", position{"large program", 1, 37, 37}},
			token{ID, "foo", position{"large program", 2, 2, 4}},
			token{INC, "++", position{"large program", 2, 5, 6}},
			token{RCURLY, "}", position{"large program", 3, 0, 0}},
			token{EOF, "", position{"large program", 3, 1, 1}}}},
	{"linecount",
		"# comment\n" +
			"# blank line\n" +
			"\n" +
			"foo", []token{
			token{ID, "foo", position{"linecount", 3, 0, 2}},
			token{EOF, "", position{"linecount", 3, 3, 3}}}},
	// errors
	{"unexpected char", "?", []token{
		token{INVALID, "Unexpected input: '?'", position{"unexpected char", 0, 0, 0}}}},
	{"unterminated regex", "/foo\n", []token{
		token{INVALID, "Unterminated regular expression: \"/foo\"", position{"unterminated regex", 0, 0, 3}}}},
	{"unterminated quoted string", "\"foo\n", []token{
		token{INVALID, "Unterminated quoted string: \"\\\"foo\"", position{"unterminated quoted string", 0, 0, 3}}}},
}

// collect gathers the emitted items into a slice.
func collect(t *lexerTest) (tokens []token) {
	l := newLexer(t.name, strings.NewReader(t.input))
	for {
		token := l.nextToken()
		tokens = append(tokens, token)
		if token.kind == EOF || token.kind == INVALID {
			break
		}
	}
	return
}

func TestLex(t *testing.T) {
	for _, test := range lexerTests {
		tokens := collect(&test)
		diff := pretty.Compare(test.tokens, tokens)
		if len(diff) > 0 {
			t.Errorf("%s tokens didn't match:\n%s:", test.name, diff)
		}
	}
}
