package dnspacket

import "revolver/cmd/buffer"
import "net"

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
    // AAAA
)

type DnsQuestion struct {
    Name string
    QType QueryType
}

func ReadDnsQuestion(b *buffer.PacketBuffer) DnsQuestion {
    name := b.MustReadQualifiedName() 
    qtype := QueryType(b.MustReadUInt16())
    b.MustReadUInt16() // Class name.
    return DnsQuestion{
        Name: name,
        QType: qtype,
    }
}

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
    return r.domain
}

func (r *UnknownRecord) TTL() uint32 {
    return r.ttl
}

type ARecord struct {
    domain string
    addr net.IP 
    ttl  uint32
}

func (r *ARecord) Domain() string {
    return r.domain
}

func (r *ARecord) TTL() uint32 {
    return r.ttl
}

func ReadDNSRecord(b *buffer.PacketBuffer) DNSRecord {
    domainName := b.MustReadQualifiedName()
    qtype := QueryType(b.MustReadUInt16())
    b.MustReadUInt16() // Class I think. Whatever.
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
            	addr:   addr,
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
            panic("Only supports A records atm :')")
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

type DNSPacket struct {
    HDR *DnsHeader
    Questions []DnsQuestion
    Answers []DNSRecord
    Authorities []DNSRecord
    Resources []DNSRecord
}

func New() DNSPacket {
    return DNSPacket{
    	HDR:         &DnsHeader{},
    	Questions:   []DnsQuestion{},
    	Answers:     []DNSRecord{},
    	Authorities: []DNSRecord{},
    	Resources:   []DNSRecord{},
    }
}

func (p *DNSPacket) FromPacketBuffer(b *buffer.PacketBuffer) {
    p.HDR = HddrFromBuffer(b)

    for i := uint16(0); i <= p.HDR.QDCOUNT; i++ {
        p.Questions = append(p.Questions, ReadDnsQuestion(b))
    }

    for i := uint16(0); i <= p.HDR.ANCOUNT; i++ {
        p.Answers = append(p.Answers, ReadDNSRecord(b))
    }

    for i := uint16(0); i <= p.HDR.NSCOUNT; i++ {
        p.Authorities = append(p.Authorities, ReadDNSRecord(b))
    }

    for i := uint16(0); i <= p.HDR.ARCOUNT; i++ {
        p.Resources = append(p.Resources, ReadDNSRecord(b))
    }
}


