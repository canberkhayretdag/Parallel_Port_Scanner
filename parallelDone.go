package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

// Critical ports
var ports = []int{21, 22, 23, 80, 135, 443, 1433, 3306, 8080}

var timeout = 100 * time.Millisecond

func findIPs(pcdir string) []string {
	var hosts []string
	// convert string to IPNet struct
	_, ipv4Net, err := net.ParseCIDR(pcdir)
	if err != nil {
		log.Fatal(err)
	}

	// convert IPNet struct mask and address to uint32
	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)

	// find the final address
	finish := (start & mask) | (mask ^ 0xffffffff)

	// loop through addresses as uint32
	for i := start; i <= finish; i++ {
		// convert back to net.IP
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		hosts = append(hosts, fmt.Sprint(ip))
	}

	return hosts

}

func portScanWorker(IPchan chan string, results chan int) {

	for ip := range IPchan {
		for i := 0; i < len(ports); i++ {
			address := fmt.Sprintf(ip+":%d", ports[i])
			conn, err := net.DialTimeout("tcp", address, timeout)
			if err != nil {
				//fmt.Println(address)
				results <- 0
				continue
			}
			fmt.Println(address)
			conn.Close()
			results <- 1
		}

	}

}

func main() {

	pcdir := flag.String("cdir", "-", "CDIR")
	flag.Parse()

	ips := findIPs(*pcdir)

	IPchan := make(chan string, cap(ips))
	results := make(chan int, cap(ips))
	for i := 0; i < 25; i++ {
		go portScanWorker(IPchan, results)
	}

	for i := 0; i < cap(ips); i++ {
		IPchan <- ips[i]
	}

	close(IPchan)

	for a := 0; a < cap(results); a++ {
		<-results

	}
}
