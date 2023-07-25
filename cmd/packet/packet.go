package dnspacket

import (
	"fmt"
	"net"
	"revolver/cmd/buffer"
)

type RCODE byte
const (
	NOERROR RCODE = iota
	FORMERR
	SERVFAIL
	NXDOMAIN
	NOTIMP
	REFUSED
)

type QueryType uint16
const (
    UNKNOWN QueryType = iota
    A
	NS
	CNAME = 5
	AAAA = 25

    // AAAA
)

type DNSQuestion struct {
    Name string
    QType QueryType
}

func MustReadDNSQuestion(b *buffer.PacketBuffer) DNSQuestion {
    name := b.MustReadQualifiedName() 
    qtype := QueryType(b.MustReadUInt16())
    b.MustReadUInt16() // Class name.
    return DNSQuestion{
        Name: name,
        QType: qtype,
    }
}
 func (q *DNSQuestion) MustWriteDNSQuestion(b *buffer.PacketBuffer) {
	b.MustWriteQName(q.Name)
	t := uint16(q.QType)
	b.MustWriteU16(t)
	b.MustWriteU16(1)
 }

// Boiler plate records spam comes next.

type DNSRecord interface {
    Domain() string
    TTL() uint32
}

type UnknownRecord struct {
    domain string
    qtype QueryType 
    dataLen uint16
    ttl uint32
}

func (r *UnknownRecord) Domain() string {
    if r.domain == "" {
		return "err? empty domain"
	} else {
		return r.domain
	}
}

func (r *UnknownRecord) TTL() uint32 {
    return r.ttl
}

type ARecord struct {
    domain string
    Addr net.IP 
    ttl  uint32
}

func (r *ARecord) Domain() string {
    return r.domain
}

func (r *ARecord) TTL() uint32 {
    return r.ttl
}

type AAAARecord struct {
    domain string
    Addr net.IP 
    ttl  uint32
}

func (r *AAAARecord) Domain() string {
    return r.domain
}

func (r *AAAARecord) TTL() uint32 {
    return r.ttl
}

type NSRecord struct {
    domain string
    Host string 
    ttl  uint32
}

func (r *NSRecord) Domain() string {
    return r.domain
}

func (r *NSRecord) TTL() uint32 {
    return r.ttl
}


type CNameRecord struct {
    domain string
    Host string 
    ttl  uint32
}

func (r *CNameRecord) Domain() string {
    return r.domain
}

func (r *CNameRecord) TTL() uint32 {
    return r.ttl
}

// TODO rename +  move to utils or something
// also IDK 
func whyIsThisSoPainfulV6(a, b, c, d uint32) net.IP {
	return net.IP{
                uint8((a >> 24) & 0xFF),
                uint8((a >> 16) & 0xFF),
                uint8((a >>  8) & 0xFF),
                uint8((a >> 0) & 0xFF),
                uint8((b >> 24) & 0xFF),
                uint8((b >> 16) & 0xFF),
                uint8((b >>  8) & 0xFF),
                uint8((b >> 0) & 0xFF),
                uint8((c >> 24) & 0xFF),
                uint8((c >> 16) & 0xFF),
                uint8((c >>  8) & 0xFF),
                uint8((c >> 0) & 0xFF),
                uint8((d >> 24) & 0xFF),
                uint8((d >> 16) & 0xFF),
                uint8((d >>  8) & 0xFF),
                uint8((d >> 0) & 0xFF),
	}
}


func ReadDNSRecord(b *buffer.PacketBuffer) DNSRecord {
    domainName := b.MustReadQualifiedName()
    qtype := QueryType(b.MustReadUInt16())
    b.MustReadUInt16() // Class I think? can't remember. Whatever. dismissed.
    ttl := b.MustReadUInt32()
    dataLen := b.MustReadUInt16()

    switch qtype {
        case A:
            addressRaw := b.MustReadUInt32()
            addr := net.IPv4(
                uint8((addressRaw >> 24) & 0xFF),
                uint8((addressRaw >> 16) & 0xFF),
                uint8((addressRaw >>  8) & 0xFF),
                uint8((addressRaw >> 0) & 0xFF),
            )
            return &ARecord{
            	domain: domainName,
            	Addr:   addr,
            	ttl:    ttl,
            }
		case AAAA:
			addr1 := b.MustReadUInt32()
			addr2 := b.MustReadUInt32()
			addr3 := b.MustReadUInt32()
			addr4 := b.MustReadUInt32()
			Addr := whyIsThisSoPainfulV6(addr1,addr2,addr3,addr4) 
			return &AAAARecord{
				domain: domainName,
				Addr: Addr,
				ttl: ttl,
			}
		case NS:
			namespace := b.MustReadQualifiedName()
			return &NSRecord{
				domain: domainName,
				Host:   namespace,
				ttl:    ttl,
			}
		case CNAME:
			cname := b.MustReadQualifiedName()
			return &CNameRecord{
				domain: domainName,
				Host:   cname,
				ttl:    ttl,
			}
        case UNKNOWN:
            b.Step(int(dataLen))
            return &UnknownRecord{
            	domain:  domainName,
            	qtype:   qtype,
            	dataLen: dataLen,
            	ttl:     ttl,
            }
        default:
	    fmt.Println(qtype)
            panic("Only supports A/AAAA/CNAME/NS records atm :')")
    }
}

func MustWriteDNSRecord(b *buffer.PacketBuffer, d DNSRecord) int {
	start := b.Pos()
	switch d.(type) {
	case *ARecord:
		b.MustWriteQName(d.Domain())
		b.MustWriteU16(uint16(A))
		b.MustWriteU16(1)
		b.MustWriteU32(d.TTL())
		b.MustWriteU16(4)
		octets := d.(*ARecord).Addr
		b.MustWriteU8(octets[0])
		b.MustWriteU8(octets[1])
		b.MustWriteU8(octets[2])
		b.MustWriteU8(octets[3])
	case *NSRecord:
		b.MustWriteQName(d.Domain())
		b.MustWriteU16(uint16(NS))
		b.MustWriteU16(1)
		b.MustWriteU32(d.TTL())

		pos := b.Pos()
		b.MustWriteU16(0)

		host := d.(*CNameRecord).Host
		b.MustWriteQName(host)
		size := b.Pos() - (pos + 2)
		b.SetU16(pos, uint16(size))
	case *CNameRecord:
		b.MustWriteQName(d.Domain())
		b.MustWriteU16((uint16(CNAME)))
		b.MustWriteU16(1)
		b.MustWriteU32(d.TTL())

		pos := b.Pos()
		b.MustWriteU16(0)

		host := d.(*CNameRecord).Host
		b.MustWriteQName(host)
		size := b.Pos() - (pos + 2)
		b.SetU16(pos, uint16(size))
	case *AAAARecord:
		b.MustWriteQName(d.Domain())
		b.MustWriteU16(uint16(AAAA))
		b.MustWriteU16(1)
		b.MustWriteU32(d.TTL())
		b.MustWriteU16(16)
		addr := d.(*AAAARecord).Addr

		for _, segment := range addr {
			b.MustWriteU8(segment)
		}
	case *UnknownRecord:
		fmt.Printf("skipping unknown record %+v\n", d)
	}
	return (b.Pos() - start)
} 

func MustWritePacket(b *buffer.PacketBuffer, d *DNSPacket) {	
	d.HDR.QDCOUNT = uint16(len(d.Questions))
	d.HDR.ANCOUNT = uint16(len(d.Answers))
	d.HDR.NSCOUNT = uint16(len(d.Authorities))
	d.HDR.ARCOUNT = uint16(len(d.Resources))

	d.HDR.MustWriteHddrToBuf(b)

	for _, q := range d.Questions {
		q.MustWriteDNSQuestion(b)
	}
	for _, d := range d.Resources{
		MustWriteDNSRecord(b, d)
	}
	for _, d := range d.Authorities{
		MustWriteDNSRecord(b, d)
	}
	for _, d := range d.Resources{
		MustWriteDNSRecord(b, d)
	}
}


type DnsHeader struct {
	Id			uint16
	QR 			bool   // 0 for query, 1 for response 
	OPCODE 		byte   // IDK, actually 4 bits 
	AA			bool   // Authorative Answer - 1 if server owns domain.
	TC 			bool   // Truncated Message
	RD 			bool   // recur or no (from sender)
	RA 			bool   // recur or no (from server)
	Z			bool    // used to DNSSEC - 3 bits
	RCODE		RCODE // the server sets this to indicate if it is successful or failed - 4 bits
	QDCOUNT     uint16 // number of entries in question section
	ANCOUNT     uint16 // number of entries in answer section
	NSCOUNT 	uint16 // number of entries in authority section
	ARCOUNT     uint16 // number of entries in additional section
}

func newhddr() *DnsHeader {
	return &DnsHeader{
		Id:      0,

		QR:      false,
		OPCODE:  0,
		AA:      false,
		TC:      false,
		RD:      false,

		RA:      false,
		Z:       false,
		RCODE:   0,

		QDCOUNT: 0,
		ANCOUNT: 0,
		NSCOUNT: 0,
		ARCOUNT: 0,
	}
}

func HddrFromBuffer(buffer *buffer.PacketBuffer) *DnsHeader {
    d := newhddr()
	d.Id = buffer.MustReadUInt16()
	allHeaderFlags := buffer.MustReadUInt16()
	// QR, OPCODE, AA, TC, RD
	firstHalf := byte(allHeaderFlags >> 8)

	secondHalf := byte(allHeaderFlags & 0xFF)
	d.RD = firstHalf & 1 > 0
	d.TC = firstHalf & (1 << 1) > 0
	d.AA = firstHalf & (1 << 2) > 0
	d.OPCODE = (firstHalf >> 3) & 0x0F
	d.QR = firstHalf & (1 << 7) > 0

	d.RCODE = RCODE(secondHalf & 0x0F)
	d.Z = (secondHalf & (1 << 6)) > 0 // Probably not gonna touch DNSSEC?
	d.RA = (secondHalf & (1 << 7))  > 0 

	d.QDCOUNT = buffer.MustReadUInt16()
	d.ANCOUNT = buffer.MustReadUInt16()
	d.NSCOUNT = buffer.MustReadUInt16()
	d.ARCOUNT = buffer.MustReadUInt16()
    return d
}

// Can't convert a byte to uint8 for some reason..
func btoi(b bool) byte {
	if b {
		return 1
	} else {
		return 0
	}
}

func (d *DnsHeader) MustWriteHddrToBuf(b *buffer.PacketBuffer) {
	b.MustWriteU16(d.Id)
	flags := (btoi(d.RD) | (btoi(d.TC) << 1) | (btoi(d.AA) << 2) | (d.OPCODE << 3) | (btoi(d.QR) << 7))
	b.MustWriteU8(flags)
	// TODO.. Probably should provide structure for RCODE bits.
	flags = ((byte(d.RCODE) << 4)) | (btoi(d.Z) << 6) | (btoi(d.RA) << 7)
	b.MustWriteU8(flags)
	b.MustWriteU16(d.QDCOUNT)
	b.MustWriteU16(d.ANCOUNT)
	b.MustWriteU16(d.NSCOUNT)
	b.MustWriteU16(d.ARCOUNT)
}


type DNSPacket struct {
    HDR *DnsHeader
    Questions []DNSQuestion
    Answers []DNSRecord
    Authorities []DNSRecord
    Resources []DNSRecord
}

func New() DNSPacket {
    return DNSPacket{
    	HDR:         &DnsHeader{},
    	Questions:   []DNSQuestion{},
    	Answers:     []DNSRecord{},
    	Authorities: []DNSRecord{},
    	Resources:   []DNSRecord{},
    }
}

func (p *DNSPacket) MustGetSomeARecord() (*ARecord, error) {
	for _, record := range p.Answers {
		switch record.(type) {
		case *ARecord:
			rec := record.(*ARecord)
			return rec, nil
		default:
			// Fallthrough
		}
	}
	return nil, fmt.Errorf("No A record in question section.")
}




func FromRaw(b []byte) DNSPacket {
	buf := buffer.FromSlice(b)
	p := New()
	p.FromPacketBuffer(buf)
	return p
}

func (p *DNSPacket) FromPacketBuffer(b *buffer.PacketBuffer) {
    p.HDR = HddrFromBuffer(b)

    for i := uint16(0); i < p.HDR.QDCOUNT; i++ {
        p.Questions = append(p.Questions, MustReadDNSQuestion(b))
    }

    for i := uint16(0); i < p.HDR.ANCOUNT; i++ {
        p.Answers = append(p.Answers, ReadDNSRecord(b))
    }

    for i := uint16(0); i < p.HDR.NSCOUNT; i++ {
        p.Authorities = append(p.Authorities, ReadDNSRecord(b))
    }

    for i := uint16(0); i < p.HDR.ARCOUNT; i++ {
        p.Resources = append(p.Resources, ReadDNSRecord(b))
    }
}


