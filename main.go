package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"time"

	psnet "github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

/* ConnectionInfo */
type ConnectionInfo struct {
	ProgramName string
	RemoteAddr  string
	RemotePort  uint32
}

/* Connections */
type Connections struct {
	list []ConnectionInfo
}

func NewConnections() *Connections {
	return &Connections{}
}

func (c *Connections) RefreshConnections() {
	c.list = nil

	currUser, err := user.Current()
	if err != nil {
		log.Fatalf("no user: %v", err)
	}

	conns, err := psnet.Connections("all")
	if err != nil {
		log.Fatalf("error getting connections: %v", err)
	}

	for _, conn := range conns {
		if conn.Status == "ESTABLISHED" {
			proc, err := process.NewProcess(conn.Pid)
			if err != nil {
				log.Printf("error on getting PID %d: %v", conn.Pid, err)
				continue
			}

			procUser, err := proc.Username()
			if err != nil {
				continue
			}

			if procUser == currUser.Username {
				procName, _ := proc.Name()
				connectionInfo := ConnectionInfo{
					ProgramName: procName,
					RemoteAddr:  conn.Raddr.IP,
					RemotePort:  conn.Raddr.Port,
				}
				c.list = append(c.list, connectionInfo)
			}
		}
	}
}

func (c *Connections) GetConnections() []ConnectionInfo {
	if c.list == nil {
		c.RefreshConnections()
	}
	return c.list
}

/* Collector */
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

func ip2Integer(ipAddress string) uint32 {
	ip := net.ParseIP(ipAddress)
	ip1 := uint32(ip[12])
	ip2 := uint32(ip[13])
	ip3 := uint32(ip[14])
	ip4 := uint32(ip[15])
	return ip1 + ip2<<8 + ip3<<16 + ip4<<24
}

func Integer2Ip(ipAddress uint32) string {
	ip1 := byte(ipAddress & 0xFF)
	ip2 := byte((ipAddress >> 8) & 0xFF)
	ip3 := byte((ipAddress >> 16) & 0xFF)
	ip4 := byte((ipAddress >> 24) & 0xFF)
	ip := net.IPv4(ip1, ip2, ip3, ip4)
	return ip.String()
}

func combineDayHour(timestamp int64) byte {
	t := time.Unix(timestamp, 0)
	dayOfWeek := int(t.Weekday())
	hour := t.Hour()
	combined := byte(hour<<3) | byte(dayOfWeek)
	return combined
}

func addStringToList(list *[]string, str string) uint16 {
	for i, s := range *list {
		if s == str {
			return uint16(i)
		}
	}
	*list = append(*list, str)
	return uint16(len(*list) - 1)
}

func addIpToList(list *[]uint32, ip uint32) {
	for _, v := range *list {
		if v == ip {
			return
		}
	}
	*list = append(*list, ip)
}

func splitDate(b byte) (int, int) {
	hour := int(b >> 3)
	day := int(b & 0x07)
	return hour, day
}

func SaveToBinaryFile(data []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	const numRuns = 100
	const interval = time.Minute
	connections := NewConnections()
	collector := NewCollector()
	for i := 0; i < numRuns; i++ {
		connections.GetConnections()
		for _, info := range connections.GetConnections() {
			collector.AddConnectionInfo(info)
		}
		collector.Print()
		binaryData, err := collector.EncodeToBinary()
		if err != nil {
			log.Fatalf("Errore durante la codifica gob: %v", err)
		}
		SaveToBinaryFile(binaryData, "data.bin")
		fmt.Printf("Dati codificati in formato binario:\n")
		fmt.Printf("%x\n", binaryData)
		time.Sleep(interval)
	}
}
