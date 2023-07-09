package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"revolver/cmd/buffer"
	dnspacket "revolver/cmd/packet"
)

// Recursive lookup.
// Lookup NS records, grab a glue record.
// Keep going until we find query name.
func RecurOnGlueRecord(qname string, qtype dnspacket.QueryType) (dnspacket.DNSPacket, error) {
	// a.root-server.net
    fmt.Println("Recurring on glue records.")
	ns := "198.41.0.4:53"
    id := uint16(777)
	for {
		fmt.Printf("Attempting to look up %s in ns %s\n", qname, ns)

		r, err := lookup(qname, qtype, id, ns)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

        response := dnspacket.FromRaw(r)
        fmt.Println("Am i hanging here... 1")
        if (len(response.Answers) != 0) && response.HDR.RCODE == dnspacket.NOERROR {
            // We are done.
            fmt.Printf("SUCCESS: %q", qname)
            printRecord(&response)

            return response, nil
        }

        if response.HDR.RCODE == dnspacket.NXDOMAIN {
            // Also valid, domain name doesn't exist.
            fmt.Printf("NXDOMAIN: %q", qname)
            return response, nil
        }
        glue, err := getSomeGlueRecord(&response)
        var noGlueRecords dnspacket.DNSPacket 
        if err != nil && glue == "" {
            fmt.Println(err)
            fmt.Println("zz:", noGlueRecords)
            // TODO: Handle this gracefully.
            os.Exit(1)
        } else if err != nil && glue != "" {
            // Naming is the hardest thing in computer science obviously
            fmt.Println("NO glue record. Gotta resolve the nameserver :')")
            fmt.Println("NO glue record. Gotta resolve the nameserver :')")
            // Start with the root..
            getNS, _ := RecurOnGlueRecord(glue, qtype)
            noGlueRecords = getNS
        } else {
            ns = glue
            continue
        }
        var nextNS string
        for _, a := range noGlueRecords.Resources{
            switch rec := a.(type) {
            case *dnspacket.ARecord:
                nextNS = rec.Addr.String()
            default: 
                continue
            }
        }
        ns = nextNS
	}
}
func printRecord(p *dnspacket.DNSPacket) {
    fmt.Println(p.HDR)

    for _, blah := range p.Resources {
        fmt.Printf("in resource: %+v\n", blah)
    }
    for _, blah := range p.Answers {
        fmt.Printf("in ans: %+v\n", blah)
    }
    for _, blah := range p.Authorities {
        fmt.Printf("in auth: %+v\n", blah)
    }
    for _, blah := range p.Questions{
        fmt.Printf("in qs: %+v\n", blah)
    }
}


func getSomeGlueRecord(p *dnspacket.DNSPacket) (string, error) {
        fmt.Println("Fetching glue record.")
        printRecord(p) 
        nservers := map[string]bool{}
        for _, a := range p.Authorities {
            switch rec := a.(type) {
            case *dnspacket.NSRecord:
                nservers[rec.Host] = true
            default:
                continue
            }

        }
        fmt.Println(nservers)
        for _, a := range p.Resources{
            switch rec := a.(type) {
            case *dnspacket.ARecord:
                fmt.Println("Found A record for: ", rec.Domain())
                _, ok := nservers[rec.Domain()]
                if ok {
                    fmt.Println("Found A record: ", rec.Addr.String())
                    return rec.Addr.String() + ":53", nil
                }
            }
        }
        // No glue record.
        var noGlueValue string
        for k := range nservers {
            noGlueValue = k
            break
        }
        return noGlueValue, fmt.Errorf("Nameserver doesn't play nice. No glue record.")
}

func lookup(qname string, qtype dnspacket.QueryType, id uint16, server string) ([]byte, error){
    udpAddr, _ := net.ResolveUDPAddr("udp", server)

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
	packet.HDR.QDCOUNT = 1
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
		a, err := RecurOnGlueRecord(q.Name, q.QType)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
        buff := buffer.New()
        dnspacket.MustWritePacket(buff, &a)
	    	
		conn.WriteToUDP(buff.Inner(), addr)
	}
}
