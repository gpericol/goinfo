// collector.go
package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

type Collector struct {
	Programs []string
	Data     map[byte]map[uint16]map[uint32][]uint32
}

func NewCollector() *Collector {
	return &Collector{
		Programs: make([]string, 0),
		Data:     make(map[byte]map[uint16]map[uint32][]uint32),
	}
}

func (c *Collector) AddConnectionInfo(info ConnectionInfo) {
	timestamp := combineDayHour(time.Now().Unix())
	programId := addStringToList(&c.Programs, info.ProgramName)

	if c.Data[timestamp] == nil {
		c.Data[timestamp] = make(map[uint16]map[uint32][]uint32)
	}

	if c.Data[timestamp][programId] == nil {
		c.Data[timestamp][programId] = make(map[uint32][]uint32)
	}

	if c.Data[timestamp][programId][info.RemotePort] == nil {
		c.Data[timestamp][programId][info.RemotePort] = make([]uint32, 0)
	}

	ips := c.Data[timestamp][programId][info.RemotePort]
	addIpToList(&ips, ip2Integer(info.RemoteAddr))
	c.Data[timestamp][programId][info.RemotePort] = ips
}

func (c *Collector) EncodeToBinary() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(c)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeCollectorFromBinary(data []byte) (*Collector, error) {
	var c Collector
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Collector) Print() {
	for timestamp, programs := range c.Data {
		hour, day := splitDate(timestamp)
		fmt.Printf("%d - %d\n", hour, day)
		for programId, ports := range programs {
			fmt.Printf("\tProgram: %s\n", c.Programs[programId])
			for port, ips := range ports {
				fmt.Printf("\t\tPort: %d\n", port)
				for _, ip := range ips {
					fmt.Printf("\t\t\tIP: %s\n", Integer2Ip(ip))
				}
			}
		}
	}
}
