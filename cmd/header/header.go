package dnsheader

import "revolver/cmd/buffer"

type RCODE byte
const (
	NOERROR RCODE = iota
	FORMERR
	SERVFAIL
	NXDOMAIN
	NOTIMP
	REFUSED
)

type DnsHeader struct {
	Id			uint16
	QR 			bool   // 0 for query, 1 for response 
	OPCODE 		byte   // IDK, actually 4 bits 
	AA			bool   // Authorative Answer - 1 if server owns domain.
	TC 			bool   // Truncated Message
	RD 			bool   // recur or no (from sender)
	RA 			bool   // recur or no (from server)
	Z			bool // used to DNSSEC - 3 bits
	RCODE		RCODE // the server sets this to indicate if it is successful or failed - 4 bits
	QDCOUNT     uint16 // number of entries in question section
	ANCOUNT     uint16 // number of entries in answer section
	NSCOUNT 	uint16 // number of entries in authority section
	ARCOUNT     uint16 // number of entries in additional section
}

func New() *DnsHeader {
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

func (d *DnsHeader) FromBuffer(buffer *buffer.PacketBuffer) {
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
}
