package main

import (
	"log"
	"net"

	"github.com/djoyahoy/tftpd/server"
	"github.com/djoyahoy/tftpd/storage"
)

func main() {
	store := storage.NewMemStore()

	addr, err := net.ResolveUDPAddr("udp", ":60001")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, tid, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}

		raw, err := server.MakePacket(buf[:n])
		if err != nil {
			log.Println("Bad request", err)
			server.Send(conn, tid, server.ErrPacket{0, "Malformed packet"})
			continue
		}

		go dispatch(raw, tid, store)
	}
}

func dispatch(raw interface{}, tid net.Addr, store storage.Storage) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		log.Println("Listen packet failure", err)
		return
	}
	defer conn.Close()

	switch packet := raw.(type) {
	case server.ReadPacket:
		log.Println("Read for file", packet.File)
		server.HandleRead(conn, tid, packet, store)
	case server.WritePacket:
		log.Println("Write for file", packet.File)
		server.HandleWrite(conn, tid, packet, store)
	default:
		log.Println("Bad request packet type", packet)
		server.Send(conn, tid, server.ErrPacket{0, "Malformed packet"})
	}
}
