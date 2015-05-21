package server

import (
	"bytes"
	"testing"
)

func TestAckPacketBytes(t *testing.T) {
	p := AckPacket{3}
	b, err := p.Bytes()
	if err != nil {
		t.Fatal("Error writing bytes")
	} else if !bytes.Equal(b, []byte{0, 4, 0, 3}) {
		t.Fatal("Invalid bytes")
	}
}

func TestDataPacketBytes(t *testing.T) {
	p := DataPacket{4, []byte{65, 66, 67}}
	b, err := p.Bytes()
	if err != nil {
		t.Fatal("Error writing bytes")
	} else if !bytes.Equal(b, []byte{0, 3, 0, 4, 65, 66, 67}) {
		t.Fatal("Invalid bytes")
	}
}

func TestErrPacketBytes(t *testing.T) {
	p := ErrPacket{1, "ABC"}
	b, err := p.Bytes()
	if err != nil {
		t.Fatal("Error writing bytes")
	} else if !bytes.Equal(b, []byte{0, 5, 0, 1, 65, 66, 67, 0}) {
		t.Fatal("Invalid bytes")
	}
}

func TestReadUint16(t *testing.T) {
	b := []byte{}
	v, err := readUint16(bytes.NewReader(b))
	if err == nil {
		t.Error("Empty buffer did not return error")
	}

	b = []byte{0, 2}
	v, err = readUint16(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	} else if v != 2 {
		t.Error("Expected 2, got", v)
	}

	b = []byte{0, 4, 0, 2}
	v, err = readUint16(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	} else if v != 4 {
		t.Error("Expected 4, got", v)
	}
}

func TestReadRWPacket(t *testing.T) {
	b := []byte{}
	_, _, err := readRWPacket(b)
	if err == nil {
		t.Error("Empty buffer did not return error")
	}

	b = []byte{65, 66, 67}
	_, _, err = readRWPacket(b)
	if err == nil {
		t.Error("Malformed bytes did not return error")
	}

	b = []byte{65, 66, 67, 0}
	f, m, err := readRWPacket(b)
	if err == nil {
		t.Error("Malformed bytes did not return error")
	}

	b = []byte{65, 66, 67, 0, 68, 69, 70}
	f, m, err = readRWPacket(b)
	if err == nil {
		t.Error("Malformed bytes did not return error")
	}

	b = []byte{65, 66, 67, 0, 68, 69, 70, 0}
	f, m, err = readRWPacket(b)
	if err != nil {
		t.Error(err)
	} else if f != "ABC" || m != "DEF" {
		t.Error("Expected ABC and DEF, got", f, m)
	}
}

func TestMakeReadPacket(t *testing.T) {
	b := []byte{0, 1, 65, 66, 67, 0, 68, 69, 70, 0}
	p, err := MakePacket(b)
	if err != nil {
		t.Fatal(err)
	}
	r := p.(ReadPacket)
	if r.File != "ABC" || r.Mode != "DEF" {
		t.Fatal("Bad read packet", p)
	}
}

func TestMakeWritePacket(t *testing.T) {
	b := []byte{0, 2, 65, 66, 67, 0, 68, 69, 70, 0}
	p, err := MakePacket(b)
	if err != nil {
		t.Fatal(err)
	}
	w := p.(WritePacket)
	if w.File != "ABC" || w.Mode != "DEF" {
		t.Fatal("Bad write packet", p)
	}
}

func TestMakeDataPacket(t *testing.T) {
	b := []byte{0, 3, 0, 1, 65, 66, 67, 68, 69, 70}
	p, err := MakePacket(b)
	if err != nil {
		t.Fatal(err)
	}
	d := p.(DataPacket)
	if d.Block != 1 || !bytes.Equal(d.Data, []byte{65, 66, 67, 68, 69, 70}) {
		t.Fatal("Bad data packet", d)
	}
}

func TestMakeAckPacket(t *testing.T) {
	b := []byte{0, 4, 0, 2}
	p, err := MakePacket(b)
	if err != nil {
		t.Fatal(err)
	}
	a := p.(AckPacket)
	if a.Block != 2 {
		t.Fatal("Bad ack packet", a)
	}
}

func TestMakeErrPacket(t *testing.T) {
	b := []byte{0, 5, 0, 6, 65, 66, 67, 0}
	p, err := MakePacket(b)
	if err != nil {
		t.Fatal(err)
	}
	e := p.(ErrPacket)
	if e.ErrCode != 6 || e.ErrMsg != "ABC" {
		t.Fatal("Bar err packet", e)
	}
}

func TestMakeBadPacket(t *testing.T) {
	b := []byte{0, 9, 65, 66, 67, 0, 68, 69, 70, 0}
	_, err := MakePacket(b)
	if err == nil {
		t.Fatal("Expected error, go nil")
	}
}
