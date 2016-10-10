package utils

import (
	"os"
	"reflect"
	"testing"
)

func TestGetOpt(t *testing.T) {
	// If EnvVar has a value, GetOpt should return it.
	expected := "bar"
	err := os.Setenv("TEST_FOO", expected)
	if err != nil {
		t.Error(err)
	}

	actual := GetOpt("TEST_FOO", "baz")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}

	// If EnvVar has an empty value, GetOpt should return the default value.
	err = os.Setenv("TEST_EMPTY", "")
	expected = "foobar"
	if err != nil {
		t.Error(err)
	}
	actual = GetOpt("TEST_EMPTY", expected)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}
}
