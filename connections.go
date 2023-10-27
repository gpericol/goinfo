// connections.go
package main

import (
	"log"
	"os/user"

	psnet "github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

// ConnectionInfo represents information about a network connection.
type ConnectionInfo struct {
	ProgramName string
	RemoteAddr  string
	RemotePort  uint32
}

// Connections represents a list of network connections.
type Connections struct {
	list []ConnectionInfo
}

// NewConnections creates and returns a new Connections instance.
func NewConnections() *Connections {
	return &Connections{}
}

// RefreshConnections refreshes the list of network connections.
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

// GetConnections returns the list of network connections.
func (c *Connections) GetConnections() []ConnectionInfo {
	if c.list == nil {
		c.RefreshConnections()
	}
	return c.list
}
