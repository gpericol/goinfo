package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

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
