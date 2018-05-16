package config

import (
	"errors"
	"testing"
)

func ReaderStub(filename string) ([]byte, error) {
	return []byte("success"), nil
}

func ReaderStub2(filename string) ([]byte, error) {
	return nil, errors.New("fake")
}

func TestConfig_Load(t *testing.T) {
	cfg := New()
	cfg.StubReader(ReaderStub)
	data, _ := cfg.Load("yeah")
	if string(data) != "success" {
		t.Errorf("format string")
	}
	cfg2 := New()
	cfg2.StubReader(ReaderStub2)
	data2, _ := cfg2.Load("yes")
	if data2 != "" {
		t.Errorf("format string2")
	}
}
