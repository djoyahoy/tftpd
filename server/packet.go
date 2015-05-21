package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	_ uint16 = iota
	OpRead
	OpWrite
	OpData
	OpAck
	OpErr
)

type Packet interface {
	Bytes() ([]byte, error)
}

type ReadPacket struct {
	File string
	Mode string
}

type WritePacket struct {
	File string
	Mode string
}

type DataPacket struct {
	Block uint16
	Data  []byte
}

type AckPacket struct {
	Block uint16
}

type ErrPacket struct {
	ErrCode uint16
	ErrMsg  string
}

func (dp DataPacket) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, OpData)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.BigEndian, dp.Block)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(dp.Data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ap AckPacket) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, OpAck)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.BigEndian, ap.Block)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ep ErrPacket) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, OpErr)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.BigEndian, ep.ErrCode)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(append([]byte(ep.ErrMsg), []byte{0}...))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MakePacket(buf []byte) (interface{}, error) {
	if len(buf) < 2 {
		return nil, errors.New("No packet op code")
	}
	op, err := readUint16(bytes.NewReader(buf[:2]))
	if err != nil {
		return nil, err
	}

	switch op {
	case OpRead:
		file, mode, err := readRWPacket(buf[2:])
		if err != nil {
			return nil, err
		}
		return ReadPacket{file, mode}, nil
	case OpWrite:
		file, mode, err := readRWPacket(buf[2:])
		if err != nil {
			return nil, err
		}
		return WritePacket{file, mode}, nil
	case OpData:
		if len(buf) < 4 {
			return nil, errors.New("No data block number")
		}
		block, err := readUint16(bytes.NewReader(buf[2:4]))
		if err != nil {
			return nil, err
		}
		data := make([]byte, len(buf[4:]))
		copy(data, buf[4:])
		return DataPacket{block, data}, nil
	case OpAck:
		if len(buf) < 4 {
			return nil, errors.New("No ack block number")
		}
		block, err := readUint16(bytes.NewReader(buf[2:4]))
		if err != nil {
			return nil, err
		}
		return AckPacket{block}, nil
	case OpErr:
		if len(buf) < 4 {
			return nil, errors.New("No error code")
		}
		code, err := readUint16(bytes.NewReader(buf[2:4]))
		if err != nil {
			return nil, err
		}
		return ErrPacket{code, string(bytes.TrimSuffix(buf[4:], []byte{0}))}, nil
	}
	return nil, errors.New("Invalid op code")
}

func readUint16(rdr io.Reader) (uint16, error) {
	var i uint16
	err := binary.Read(rdr, binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func readRWPacket(buf []byte) (string, string, error) {
	s := bytes.SplitN(buf, []byte{0}, 3)
	if len(s) != 3 {
		return "", "", errors.New("Malformed packet")
	}
	return string(s[0]), string(s[1]), nil
}
