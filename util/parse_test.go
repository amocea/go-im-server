package util

import "testing"

func TestArgName2propertyName(t *testing.T) {
	name := ArgName2propertyName("createat")
	if name != "createat" {
		t.Fatalf("Not true，actual=%s, expect = %s", name, "Createat")
	}
}

func TestIsPointer(t *testing.T) {
	var x struct{}
	ip := IsPointer(&x)
	if !ip {
		t.Fatalf("Not is pointer, failed")
	}
}
