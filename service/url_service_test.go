package service

import (
	"strings"
	"testing"
)



func TestIsValidURL(t *testing.T){
	got := IsValidURL("https://example.com")
	if got != true {
		t.Errorf("isValidURL(%q) = %v, mau %v","https://example.com",got,true)
	}

	got = IsValidURL("bukan-url")
	if got != false {
		t.Errorf("isValidURL(%q) = %v, mau %v", "bukan-url",got,false)
	}
}


func TestGenerateShortCode(t *testing.T){
	length := 6

	got, err := GenerateShortCode(length)
	// cek main error fail all fatal
	if err!=nil {
		t.Fatalf("generateShortCode(%d) error: %v",length,err)
	}
	// cek panjang
	if len(got) != length {
		t.Errorf("len(generateShortCode(%d)) = %d, want %d", length,len(got),length)
	}
	// cel tiap karakter ad di charset
	for _, c:= range got {
		if !strings.ContainsRune(charset,c){
			t.Errorf("karakter %q di kode %q tidak ad di charset", c, got)
		}
	}
}


func TestGenerateShortCodeUnique(t *testing.T){
	a, err := GenerateShortCode(6)
	if err != nil{
		t.Fatalf("error: %v", err)
	}
	b, err := GenerateShortCode(6)
	if err!= nil {
		t.Fatalf("error: %v", err)
	}

	if a==b{
		t.Errorf("dua kode sama: %q, harusnya beda", a)
	}

}

