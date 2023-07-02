package header

const (
	NOERROR byte = iota
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
	Z			byte   // used to DNSSEC - 3 bits
	RCODE		byte   // the server sets this to indicate if it is successful or failed - 4 bits
	QDCOUNT     uint16 // number of entries in question section
	ANCOUNT     uint16 // number of entries in answer section
	NSCOUNT 	uint16 // number of entries in authority section
	ARCOUNT     uint16 // number of entries in additional section
}


