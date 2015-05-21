package server

import (
	"log"
	"net"
	"time"

	"github.com/djoyahoy/tftpd/storage"
)

const blockSize int = 512
const maxTimeout int = 5

type TidError struct {
	Tid net.Addr
	Msg string
}

func (t TidError) Error() string {
	return t.Msg
}

func HandleWrite(conn net.PacketConn, tid net.Addr, wp WritePacket, store storage.Storage) {
	blk := uint16(0)
	dat := make([]byte, 0)
	for reading := true; reading == true; {
		raw, err := Transmit(conn, tid, AckPacket{blk})
		if err != nil {
			log.Println("Transmit failure", err)
			Send(conn, tid, raw.(ErrPacket))
			return
		}

		switch packet := raw.(type) {
		case DataPacket:
			if packet.Block == blk + 1 {
				blk = packet.Block
				dat = append(dat, packet.Data...)
			}
			if len(packet.Data) < blockSize {
				reading = false
			}
		case ErrPacket:
			log.Println("Got an error packet", packet.ErrCode, packet.ErrMsg)
			return
		default:
			log.Println("Invalid packet", packet)
			Send(conn, tid, ErrPacket{0, "Invalid packet op"})
			return
		}
	}

	err := store.Put(wp.File, dat)
	if err != nil {
		log.Println("Storage put failure", err)
		Send(conn, tid, ErrPacket{0, "Unable to put file"})
		return
	}
	log.Printf("Put file %s with %d bytes\n", wp.File, len(dat))

	_, err = Send(conn, tid, AckPacket{blk})
	if err != nil {
		log.Println("Send final ACK failure", err)
		Send(conn, tid, ErrPacket{0, "Failed to write final ACK"})
		return
	}
}

func HandleRead(conn net.PacketConn, tid net.Addr, rp ReadPacket, store storage.Storage) {
	dat, err := store.Get(rp.File)
	if err != nil {
		log.Println("Storage get failure", err)
		Send(conn, tid, ErrPacket{1, "File does not exist"})
		return
	}

	pos := 0
	blk := uint16(1)
	for pos < len(dat) {
		chunk := min(pos + blockSize, len(dat))

		raw, err := Transmit(conn, tid, DataPacket{blk, dat[pos : chunk]})
		if err != nil {
			log.Println("Transmit failure", err)
			Send(conn, tid, raw.(ErrPacket))
			return
		}

		switch packet := raw.(type) {
		case AckPacket:
			if blk == packet.Block {
				blk += 1
				pos = chunk
			}
		case ErrPacket:
			log.Println("Got an error packet", packet.ErrCode, packet.ErrMsg)
			return
		default:
			log.Println("Invalid packet", packet)
			Send(conn, tid, ErrPacket{0, "Incorrect packet op"})
			return
		}
	}
	log.Printf("Got file %s with %d bytes\n", rp.File, len(dat))
}

func Transmit(conn net.PacketConn, tid net.Addr, packet Packet) (interface{}, error) {
	for try := 0;; {
		_, err := Send(conn, tid, packet)
		if err != nil {
			return ErrPacket{0, "Failed to send packet"}, err
		}

		// Attempt to recv a packet until max timeout or success.
		// On a transfer ID error, simply send an err packet to the host.
		raw, err := Recv(conn, tid, time.Now().Add(time.Second * 5))
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			log.Println("Recv timed out", try)
			if try++; try >= maxTimeout {
				log.Println("Reached timeout limit")
				return ErrPacket{0, "Connection timed out"}, nerr
			}
			continue
		} else if terr, ok := err.(TidError); ok {
			log.Println("Unknown tansfer ID", err)
			Send(conn, terr.Tid, ErrPacket{5, "Unknown transfer ID"})
			continue
		} else if err != nil {
			log.Println("Recv error", err)
			return ErrPacket{0, err.Error()}, err
		}

		// A packet was read without error. Exit the try loop.
		return raw, nil
	}
}

func Recv(conn net.PacketConn, tid net.Addr, timeout time.Time) (interface{}, error) {
	buf := make([]byte, 1024)
	conn.SetReadDeadline(timeout)
	n, addr, err := conn.ReadFrom(buf)
	if err != nil {
		return nil, err
	}

	// Check the source of the packet and return an error if the transfer ID is wrong.
	if addr.String() != tid.String() {
		return nil, TidError{addr, "Unknown transfer ID"}
	}

	raw, err := MakePacket(buf[:n])
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func Send(conn net.PacketConn, tid net.Addr, packet Packet) (int, error) {
	buf, err := packet.Bytes()
	if err != nil {
		return 0, err
	}

	n, err := conn.WriteTo(buf, tid)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func min(a int, b int) (int) {
	if a < b {
		return a
	}
	return b
}
