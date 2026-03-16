package utils

import (
	"regexp"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	length := 16
	password := GeneratePassword(length)

	if len(password) != length {
		t.Errorf("expected password length %d, got %d", length, len(password))
	}

	match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", password)
	if !match {
		t.Errorf("password contains invalid characters: %s", password)
	}
}

func TestGeneratePasswordDifferent(t *testing.T) {
	p1 := GeneratePassword(16)
	p2 := GeneratePassword(16)
	if p1 == p2 {
		t.Error("passwords should be different")
	}
}
