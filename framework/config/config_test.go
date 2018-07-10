package config

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type noFlags struct {
	Val  int
	Val2 int
}

func TestEachSubField(t *testing.T) {
	tests := []struct {
		testStruct  interface{}
		shouldPanic bool
		shouldError bool
		f           func(reflect.Value, string, []string) error
	}{
		// Struct must be settable
		{
			testStruct:  struct{}{},
			shouldPanic: true,
			shouldError: false,
			f:           func(reflect.Value, string, []string) error { return nil },
		},
		// Must be a struct of structs
		{
			testStruct:  1,
			shouldPanic: true,
			shouldError: false,
			f:           func(reflect.Value, string, []string) error { return nil },
		},
		{
			testStruct:  &struct{}{},
			shouldPanic: false,
			shouldError: false,
			f:           func(reflect.Value, string, []string) error { return nil },
		},
		{
			testStruct: &struct {
				Test      int
				TestSlice []string
			}{1, []string{"Hi", "Mom"}},
			shouldPanic: false,
			shouldError: false,
			f:           func(reflect.Value, string, []string) error { return nil },
		},
		{
			testStruct: &struct {
				Test struct{ Val int }
			}{
				Test: struct{ Val int }{1},
			},
			shouldPanic: false,
			shouldError: false,
			f: func(p reflect.Value, f string, bc []string) error {
				if strings.Join(bc, "") != "Test" || f != "Val" {
					t.Errorf("Expected single call for Test.Val. Got %s.%s", strings.Join(bc, ""), f)
					return fmt.Errorf("Expected single call for Test.Val. Got %s.%s", strings.Join(bc, ""), f)
				}
				return nil
			},
		},
		{
			testStruct: &struct {
				Test struct {
					Val   int `flag:"false"`
					Test2 struct {
						Val2     int
						ValSlice []string
					}
					unexported string
					SkipMe     string `flag:"false"`
				}
			}{
				Test: struct {
					Val   int `flag:"false"`
					Test2 struct {
						Val2     int
						ValSlice []string
					}
					unexported string
					SkipMe     string `flag:"false"`
				}{1, struct {
					Val2     int
					ValSlice []string
				}{2, []string{"hi", "mom"}}, "hello, you failed!", "BadFlag"},
			},
			shouldPanic: false,
			shouldError: false,
			f: func(p reflect.Value, f string, bc []string) error {
				if f != "Val2" && f != "IgnoreThis" && f != "SkipMe" && f != "ValSlice" {
					t.Errorf("Expected single call for Test.Val. Got %s.%s", strings.Join(bc, ""), f)
					return fmt.Errorf("Expected single call for Test.Val. Got %s.%s", strings.Join(bc, ""), f)
				}
				if f == "SkipMe" || f == "Val" {
					t.Error("Should not have gotten to SkipMe, since flag:false")
				}
				return nil
			},
		},

		{
			testStruct: &struct {
				NoFlags        noFlags `flag:"false"`
				DontIgnoreThis string
			}{
				NoFlags: noFlags{
					1, 2},
			},
			shouldPanic: false,
			shouldError: false,
			f: func(p reflect.Value, f string, bc []string) error {
				if f == "Val1" || f == "Val2" {
					t.Error("should not have parsed NoFlags.Val1 or NoFlags.Val2 since `flags:'false'` is set.")
				}
				return nil
			},
		},
		{
			testStruct: &struct {
				Test struct{ Val int }
			}{
				Test: struct{ Val int }{1},
			},
			shouldPanic: false,
			shouldError: false,
			f:           func(reflect.Value, string, []string) error { return fmt.Errorf("error") },
		},
	}

	//func eachSubField(i interface{}, fn func(reflect.Value, string, []string))

	for i, test := range tests {
		assertPanic(i, t, test.testStruct, test.f, test.shouldPanic, test.shouldError)
	}
}

// Wrapper function to catch panics
func assertPanic(index int, t *testing.T, i interface{}, f func(reflect.Value, string, []string) error, shouldPanic bool, shouldError bool) {
	defer func() {
		if r := recover(); r == nil && shouldPanic {
			t.Errorf("The code did not panic")
		} else if r != nil && !shouldPanic {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()
	err := eachSubField(i, f)
	if err != nil {
		if !shouldError {
			t.Errorf("Test: %d Received Err : %v, expected none.", index, err)
		}
	} else {
		if shouldError {
			t.Errorf("Test %d No Err received when one was expected", index)
		}
	}
}

func TestFlagString(t *testing.T) {
	tests := []struct {
		parent   string
		field    string
		expected string
	}{
		{"Docker", "Foo", "docker-foo"},
		{"Foo", "Bar", "foo-bar"},
		{"FOO", "FooBAR", "foo-foo-bar"},
		{"FoOo", "FooBARBaZ", "fo-oo-foo-bar-baz"},
		{"FOO", "FOoBARBaZa", "foo-f-oo-bar-ba-za"},
		{"", "bar", "bar"},
		{"bar", "", "bar"},
		{"", "", ""},
	}

	for _, test := range tests {
		if flagString(test.parent, test.field) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, flagString(test.parent, test.field))
		}
	}
}

func TestEnvString(t *testing.T) {
	tests := []struct {
		parent   string
		field    string
		expected string
	}{
		{"Docker", "Foo", "DOCKER_FOO"},
		{"Foo", "Bar", "FOO_BAR"},
		{"FOO", "FooBAR", "FOO_FOO_BAR"},
		{"FoOo", "FooBARBaZ", "FO_OO_FOO_BAR_BAZ"},
		{"FOO", "FOoBARBaZa", "FOO_F_OO_BAR_BA_ZA"},
		{"", "bar", "BAR"},
		{"bar", "", "BAR"},
		{"", "", ""},
	}

	for _, test := range tests {
		if envString(test.parent, test.field) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, flagString(test.parent, test.field))
		}
	}
}
