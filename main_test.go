package main

import "testing"

func TestisValidURL(t *testing.T){
	got := isValidURL("https://example.com")
	if got != true {
		t.Errorf("isValidURL(%q) = %v, mau %v","https://example.com",got,true)
	}

	got = isValidURL("bukan-url")
	if got != false {
		t.Errorf("isValidURL(%q) = %v, mau %v", "bukan-url",got,false)
	}
}