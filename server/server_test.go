package server

import (
	"bytes"
	"errors"
	"net"
	"testing"
	"time"
)

type MockAddr struct {
	Net string
	Host string
}

func (m MockAddr) Network() (string) {
	return m.Net
}

func (m MockAddr) String() (string) {
	return m.Host
}

type MockTimeout struct {
	Msg string
}

func (m MockTimeout) Error() string {
	return m.Msg
}

func (m MockTimeout) Timeout() bool {
	return true
}

func (m MockTimeout) Temporary() bool {
	return false
}

type MockPacketConn struct {
	Timeout bool
	WriteFail bool
	WriteBuf []byte
	ReadFail bool
	ReadAddr MockAddr
	ReadBuf []byte
}

func (m *MockPacketConn) Close() (error) {
	return nil
}

func (m *MockPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	if m.ReadFail {
		return 0, nil, errors.New("Read Fail")
	} else if m.Timeout {
		return 0, nil, MockTimeout{"foo"}
	} else {
		copy(b, m.ReadBuf)
		return len(m.ReadBuf), m.ReadAddr, nil
	}
}

func (m *MockPacketConn) LocalAddr() (net.Addr) {
	return MockAddr{"udp", ":69"}
}

func (m *MockPacketConn) SetDeadline(t time.Time) (error) {
	return nil
}

func (m *MockPacketConn) SetReadDeadline(t time.Time) (error) {
	return nil
}

func (m *MockPacketConn) SetWriteDeadline(t time.Time) (error) {
	return nil
}

func (m *MockPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	if m.WriteFail {
		return 0, errors.New("Write fail")
	} else {
		m.WriteBuf = b
		return len(m.WriteBuf), nil
	}
}

func TestSendAck(t *testing.T) {
	conn := MockPacketConn{}
	n, err := Send(&conn, MockAddr{"udp", ":70"}, AckPacket{1})
	if err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal("Expected 4, got", n)
	}
	if !bytes.Equal(conn.WriteBuf, []byte{0, 4, 0, 1}) {
		t.Fatal("Invalid write buffer", conn.WriteBuf)
	}
}

func TestSendData(t *testing.T) {
	conn := MockPacketConn{}
	n, err := Send(&conn, MockAddr{"udp", ":70"}, DataPacket{1, []byte{65, 66, 67}})
	if err != nil {
		t.Fatal(err)
	} else if n != 7 {
		t.Fatal("Expected 7, got", n)
	}
	if !bytes.Equal(conn.WriteBuf, []byte{0, 3, 0, 1, 65, 66, 67}) {
		t.Fatal("Invalid write buffer", conn.WriteBuf)
	}
}

func TestSendFail(t *testing.T) {
	conn := MockPacketConn{WriteFail: true}
	n, err := Send(&conn, MockAddr{"udp", ":70"}, AckPacket{1})
	if err == nil {
		t.Fatal("Expected nil, got", err)
	} else if n != 0 {
		t.Fatal("Expected 0, got", n)
	}
}

func TestRecvAck(t *testing.T) {
	conn := MockPacketConn{ReadAddr: MockAddr{"udp", ":70"},
		ReadBuf: []byte{0, 4, 0, 2}}
	raw, err := Recv(&conn, MockAddr{"udp", ":70"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	switch packet := raw.(type) {
	case AckPacket:
		if packet.Block != 2 {
			t.Fatal("Expected block 2, got", packet.Block)
		}
	default:
		t.Fatal("Expected ack, got", packet)
	}
}

func TestRecvData(t *testing.T) {
	conn := MockPacketConn{ReadAddr: MockAddr{"udp", ":70"},
		ReadBuf: []byte{0, 3, 0, 4, 65, 66, 67}}
	raw, err := Recv(&conn, MockAddr{"udp", ":70"}, time.Now())
	if err != nil {
		t.Error(err)
	}
	switch packet := raw.(type) {
	case DataPacket:
		if packet.Block != 4 {
			t.Fatal("Expected block 4, got", packet.Block)
		}
		if !bytes.Equal(packet.Data, []byte{65, 66, 67}) {
			t.Fatal("Inavlid data block", packet.Data)
		}
	default:
		t.Fatal("Expected ack, got", packet)
	}
}

func TestRecvFail(t *testing.T) {
	conn := MockPacketConn{ReadFail: true}
	_, err := Recv(&conn, MockAddr{"udp", ":70"}, time.Now())
	if err == nil {
		t.Error("Expected err, got nil")
	}
}

func TestRecvTimeout(t *testing.T) {
	conn := MockPacketConn{Timeout: true}
	_, err := Recv(&conn, MockAddr{"udp", ":70"}, time.Now())
	if err == nil {
		t.Fatal("Expected timeout, got nil")
	}
}

func TestRecvBadTid(t *testing.T) {
	conn := MockPacketConn{ReadAddr: MockAddr{"udp", ":70"}}
	_, err := Recv(&conn, MockAddr{"udp", ":71"}, time.Now())
	if err == nil {
		t.Fatal("Expected err, got nil")
	}
	switch et := err.(type) {
	case TidError:
		if et.Tid.Network() != "udp"  || et.Tid.String() != ":70" {
			t.Fatal("Incorret address")
		}
	default:
		t.Fatal("Expected tid error")
	}
}

func TestTransmitAck(t *testing.T) {
	conn := MockPacketConn{ReadAddr: MockAddr{"udp", ":70"},
		ReadBuf: []byte{0, 4, 0, 2}}
	raw, err := Transmit(&conn, MockAddr{"udp", ":70"}, AckPacket{1})
	if err != nil {
		t.Error(err)
	}
	switch packet := raw.(type) {
	case AckPacket:
		if packet.Block != 2 {
			t.Fatal("Expected block 2, got", packet.Block)
		}
	default:
		t.Fatal("Expected ack, got", packet)
	}
}

func TestTransmitData(t *testing.T) {
	conn := MockPacketConn{ReadAddr: MockAddr{"udp", ":70"},
		ReadBuf: []byte{0, 3, 0, 4, 65, 66, 67}}
	raw, err := Transmit(&conn, MockAddr{"udp", ":70"}, AckPacket{1})
	if err != nil {
		t.Error(err)
	}
	switch packet := raw.(type) {
	case DataPacket:
		if packet.Block != 4 {
			t.Fatal("Expected block 4, got", packet.Block)
		}
		if !bytes.Equal(packet.Data, []byte{65, 66, 67}) {
			t.Fatal("Inavlid data block", packet.Data)
		}
	default:
		t.Fatal("Expected ack, got", packet)
	}
}

func TestTransmitTimeout(t *testing.T) {
	conn := MockPacketConn{Timeout: true}
	raw, err := Transmit(&conn, MockAddr{"udp", ":70"}, AckPacket{1})
	if err == nil {
		t.Error("Expected timeout, got nil")
	}
	switch packet := raw.(type) {
	case ErrPacket:
		if packet.ErrCode != 0 {
			t.Fatal("Expected code 0, got", packet.ErrCode)
		}
	default:
		t.Fatal("Expected err, got", packet)
	}
}

func TestTransmitFail(t *testing.T) {
	conn := MockPacketConn{ReadFail: true}
	raw, err := Transmit(&conn, MockAddr{"udp", ":70"}, AckPacket{1})
	if err == nil {
		t.Error("Expected err, got nil")
	}
	switch packet := raw.(type) {
	case ErrPacket:
		if packet.ErrCode != 0 {
			t.Fatal("Expected code 0, got", packet.ErrCode)
		}
	default:
		t.Fatal("Expected err, got", packet)
	}
}
