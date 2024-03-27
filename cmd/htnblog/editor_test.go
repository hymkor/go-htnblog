package main

import (
	"testing"
)

func testFields(t *testing.T, source string, expect []string) {
	t.Helper()
	result := fields(source)
	if len(result) != len(expect) {
		t.Fatalf("expect %#v,but %#v", expect, result)
	}
	for i := range expect {
		if result[i] != expect[i] {
			t.Fatalf("expect %#v,but %#v", expect, result)
		}
	}

}

func TestFields(t *testing.T) {
	testFields(t, `a "b c"  d`, []string{"a", "b c", "d"})
	testFields(t, `a "b c" d `, []string{"a", "b c", "d"})
}
