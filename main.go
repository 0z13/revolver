package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"revolver/cmd/buffer"
	dnspacket "revolver/cmd/packet"
)

func main() {
	serve()
}

func test() {
	fileName := "./response_packet.txt"

	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read the file contents
	fileContents, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Create an array of type [512]byte
	var innerBuffer [512]byte

	// Just figured i can do this
	// s = innerBuffer[:]
	copy(innerBuffer[:], fileContents)

	buffer := buffer.New()
	buffer.SetInner(innerBuffer)

	packet := dnspacket.New()
	packet.FromPacketBuffer(buffer)

	fmt.Printf("%+v\n", packet.HDR)
	for _, q := range packet.Questions {
		fmt.Printf("%+v\n", q)
	}

	for _, r := range packet.Answers {
		fmt.Printf("%+v\n", r)
	}

	for _, blah := range packet.Authorities {
		fmt.Printf("%+v\n", blah)
	}

	for _, rec := range packet.Resources {
		fmt.Printf("%+v\n", rec)
	}
}

func serve() {
	qname := "google.com"
	qtype := dnspacket.QueryType(1)

	// bind udp to arbritrary port
	addr, err := net.ResolveUDPAddr("udp", ":53033")

	if err != nil {

		fmt.Println("zzzzz")
		panic("zzz")
	}

	sock, err := net.ListenUDP("udp", addr)
	defer sock.Close()

	if err != nil {
		panic("something went wrongk")
	}

	// build query
	qs := []dnspacket.DNSQuestion{
		{
			Name:  qname,
			QType: qtype,
		},
	}
	packet := dnspacket.DNSPacket{
		HDR:         &dnspacket.DnsHeader{},
		Questions:   qs,
	}
	packet.HDR.RD = true
	packet.HDR.QR = true
	packet.HDR.QDCOUNT = 1
	packet.HDR.Id = 6666 
	packet.Questions = qs 

	b := buffer.New()
	dnspacket.MustWritePacket(b, &packet)
	// Next just write this packet to the buffer.
	pInner := b.Inner()
	fmt.Println(pInner)

	googleDNS := &net.UDPAddr{
		Port: 53,
		IP:   net.ParseIP("8.8.8.8"),
	}
	_, err = sock.WriteToUDP(pInner[:], googleDNS) 
	if err != nil {
		fmt.Println(err)
		panic("couldn't write")
	}

	resultBuffer := buffer.New().Inner()
	x,conn ,err := sock.ReadFromUDP(resultBuffer[:])
	fmt.Println("bla")
	fmt.Println(conn)
	fmt.Println(x)
	fmt.Println(err)

	fmt.Printf("%+v\n", packet.HDR)
	for _, q := range packet.Questions {
		fmt.Printf("%+v\n", q)
	}

	for _, r := range packet.Answers {
		fmt.Printf("%+v\n", r)
	}

	for _, blah := range packet.Authorities {
		fmt.Printf("%+v\n", blah)
	}

	for _, rec := range packet.Resources {
		fmt.Printf("%+v\n", rec)
	}
}
