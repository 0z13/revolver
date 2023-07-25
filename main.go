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


func lookup(qname string, qtype dnspacket.QueryType) ([]byte, error){
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
	packet.HDR.Id = 6666 
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
	port := flag.String("l", "1234", "Provide listen port")

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + *port)
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
		q := p.Questions[0]
		// 
		a, err := lookup(q.Name, q.QType)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		
		conn.WriteToUDP(a, addr)
	}


    //serve()
}

func serve() {
	qname := "google.com"
	qtype := dnspacket.QueryType(1)

	// bind udp to arbritrary port
    // Try without port..
    udpAddr, err := net.ResolveUDPAddr("udp", "8.8.8.8:53")
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
	packet.HDR.Id = 6666 
	packet.Questions = qs 

	b := buffer.New()
	dnspacket.MustWritePacket(b, &packet)
	// Next just write this packet to the buffer.
	pInner := b.Inner()

    conn, err := net.DialUDP("udp", nil, udpAddr) 

	if err != nil {
		fmt.Println(err)
		panic("couldn't dail")
	}
	_, err = conn.Write(pInner[:])
	_, err = conn.Write([]byte("\n"))

    if err != nil {
        fmt.Println(err)
    }

    answerGoogleWrongType := make([]byte, 512) 
	_, err = bufio.NewReader(conn).Read(answerGoogleWrongType)
    
    
    fmt.Println("recieved something :)")

    // fix this in post :P 
    answerGoogle := [512]byte{}
    copy(answerGoogle[:], answerGoogleWrongType)
    bb := buffer.New()

    bb.SetInner(answerGoogle)
	returnPacket := dnspacket.New()
    returnPacket.FromPacketBuffer(bb)

    fmt.Printf("%v",returnPacket.HDR)

    for _, q := range returnPacket.Questions {
        fmt.Printf("question: %+v\n",q)
    }

    for _, a := range returnPacket.Answers {
        fmt.Printf("answer: %+v\n",a)
    }
}
