package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"revolver/cmd/buffer"
	dnspacket "revolver/cmd/packet"
)


func recLookup(qname string, qtype dnspacket.QueryType) {
	// a.root-server.net
	ns := "198.41.0.4:53"

	for {
		fmt.Printf("Attempting to look up %s in ns %s", qname, ns)

		resp, err := lookup(qname, qtype,666 ,ns)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		
	}
}


func lookup(qname string, qtype dnspacket.QueryType, id uint16, server string) ([]byte, error){
    udpAddr, _ := net.ResolveUDPAddr("udp", "8.8.8.8:53")

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
	packet.HDR.QDCOUNT = 0
	packet.HDR.Id = id 
	packet.Questions = qs 

	b := buffer.New()
	dnspacket.MustWritePacket(b, &packet)

    conn, err := net.DialUDP("udp", nil, udpAddr) 

	if err != nil {
		return nil, err
	}
	_, err = conn.Write(b.Inner())
	_, err = conn.Write([]byte("\n"))

    if err != nil {
		return nil, err
    }

    ans := make([]byte, 512) 
	_, err = bufio.NewReader(conn).Read(ans)
	return ans, nil
}



func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	conn, err := net.ListenUDP("udp", udpAddr)

	for {
		buf := make([]byte, 512)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil  {
			fmt.Println(err)
			fmt.Println("request length: ", n)
			os.Exit(1)
		}
		p := dnspacket.FromRaw(buf)
		// Expecting exactly one question.
		if len(p.Questions) == 0 {
			fmt.Println("Nothing in question section.")
			os.Exit(1)
		}
		q := p.Questions[0]
		a, err := lookup(q.Name, q.QType, p.HDR.Id)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		
		conn.WriteToUDP(a, addr)
	}
}
