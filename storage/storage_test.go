package storage

import (
	"bytes"
	"testing"
)

func TestMemStorePut(t *testing.T) {
	m := NewMemStore()

	m.Put("foo", []byte{1, 2, 3, 4, 5})
	data, ok := m.store["foo"]
	if !ok || !bytes.Equal(data, []byte{1, 2, 3, 4, 5}) {
		t.Error("Put failure.")
	}

	m.Put("bar", []byte{6, 7, 8, 9})
	data, ok = m.store["bar"]
	if !ok || !bytes.Equal(data, []byte{6, 7, 8, 9}) {
		t.Error("Put failure.")
	}

	m.Put("foo", []byte{1, 2, 3, 4})
	data, ok = m.store["foo"]
	if !ok || !bytes.Equal(data, []byte{1, 2, 3, 4}) {
		t.Error("Put failure.")
	}

	b := []byte{100, 101, 102}
	m.Put("baz", b)
	b[0] = 1
	data, ok = m.store["baz"]
	if !ok || !bytes.Equal(data, []byte{100, 101, 102}) {
		t.Error("Put failure.")
	}
}

func TestMemStoreGet(t *testing.T) {
	m := NewMemStore()
	m.store["foo"] = []byte{1, 2, 3}
	m.store["bar"] = []byte{4, 5, 6}

	_, err := m.Get("baz")
	if err == nil {
		t.Error("Bad file did not return error.")
	}

	data, err := m.Get("foo")
	if err != nil || !bytes.Equal(data, []byte{1, 2, 3}) {
		t.Error("Get failure.")
	}

	m.store["foo"] = []byte{100, 101, 102}
	data, err = m.Get("foo")
	if err != nil || !bytes.Equal(data, []byte{100, 101, 102}) {
		t.Error("Get failure.")
	}

	data, err = m.Get("bar")
	if err != nil || !bytes.Equal(data, []byte{4, 5, 6}) {
		t.Error("Get failure.")
	}

	data, err = m.Get("bar")
	data[0] = 200
	if err != nil || !bytes.Equal(m.store["bar"], []byte{4, 5, 6}) {
		t.Error("Get failure.")
	}
}
