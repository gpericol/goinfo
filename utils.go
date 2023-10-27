package main

import (
	"net"
	"os"
	"time"
)

// ip2Integer converts an IPv4 address in string format to an unsigned 32-bit integer.
func ip2Integer(ipAddress string) uint32 {
	ip := net.ParseIP(ipAddress)
	ip1 := uint32(ip[12])
	ip2 := uint32(ip[13])
	ip3 := uint32(ip[14])
	ip4 := uint32(ip[15])
	return ip1 + ip2<<8 + ip3<<16 + ip4<<24
}

// Integer2Ip converts an unsigned 32-bit integer to an IPv4 address in string format.
func Integer2Ip(ipAddress uint32) string {
	ip1 := byte(ipAddress & 0xFF)
	ip2 := byte((ipAddress >> 8) & 0xFF)
	ip3 := byte((ipAddress >> 16) & 0xFF)
	ip4 := byte((ipAddress >> 24) & 0xFF)
	ip := net.IPv4(ip1, ip2, ip3, ip4)
	return ip.String()
}

// combineDayHour combines the hour of the day and the day of the week into a single byte.
func combineDayHour(timestamp int64) byte {
	t := time.Unix(timestamp, 0)
	dayOfWeek := int(t.Weekday())
	hour := t.Hour()
	combined := byte(hour<<3) | byte(dayOfWeek)
	return combined
}

// addStringToList adds a string to a list and returns its index.
func addStringToList(list *[]string, str string) uint16 {
	for i, s := range *list {
		if s == str {
			return uint16(i)
		}
	}
	*list = append(*list, str)
	return uint16(len(*list) - 1)
}

// addIpToList adds an unsigned 32-bit integer to a list if it doesn't already exist.
func addIpToList(list *[]uint32, ip uint32) {
	for _, v := range *list {
		if v == ip {
			return
		}
	}
	*list = append(*list, ip)
}

// splitDate splits a combined byte into an hour of the day and day of the week.
func splitDate(b byte) (int, int) {
	hour := int(b >> 3)
	day := int(b & 0x07)
	return hour, day
}

// SaveToBinaryFile saves binary data to a file with the specified filename.
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

// ReadFromBinaryFile reads binary data from a file with the specified filename.
func ReadFromBinaryFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	data := make([]byte, fileInfo.Size())
	_, err = file.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
