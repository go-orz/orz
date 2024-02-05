package base58

import (
	"bytes"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte("hello golang!"), "9hLF3DC56e6t7k87sn"},
		{[]byte("hello base58!"), "9hLF3DC56SzXCpECpQ"},
	}

	for _, test := range tests {
		if got := string(Encode(test.input)); got != test.expected {
			t.Errorf("Encode(%s) = %s; want %s", test.input, got, test.expected)
		}
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
	}{
		{"9hLF3DC56e6t7k87sn", []byte("hello golang!")},
		{"9hLF3DC56SzXCpECpQ", []byte("hello base58!")},
	}

	for _, test := range tests {
		if got := Decode([]byte(test.input)); !bytes.Equal(got, test.expected) {
			t.Errorf("Decode(%s) = %v; want %v", test.input, got, test.expected)
		}
	}
}

func TestEncodeDecode(t *testing.T) {
	data := []byte("hello golang!")
	encoded := Encode(data)
	decoded := Decode(encoded)
	if !bytes.Equal(data, decoded) {
		t.Errorf("TestEncodeDecode failed: got %v, want %v", decoded, data)
	}
}

func TestReverseBytes(t *testing.T) {
	data := []byte("123456")
	reversed := reverseBytes(data)
	expected := []byte("654321")
	if !bytes.Equal(reversed, expected) {
		t.Errorf("TestReverseBytes failed: got %v, want %v", reversed, expected)
	}
}
