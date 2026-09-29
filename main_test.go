package main

import "testing"

func TestIsValidURL(t *testing.T){
	got := isValidURL("https://example.com")
	if got != true {
		t.Errorf("isValidURL(%q) = %v, mau %v","https://.com",got,true)
	}

	got = isValidURL(("bukan-url"))
	if got != true {
		t.Errorf("isValidURL(%q) = %v, mau %v", "bukan-url",got,false)
	}
}