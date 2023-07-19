package buffer

import (
	"strings"
)

type PacketBuffer struct {
	inner    [512]byte
	pos       int
}

func New() *PacketBuffer {
	var b [512]byte
	return &PacketBuffer{
		inner: b,
		pos: 0,
	}
}

func (b *PacketBuffer) SetInner(bs [512]byte) {
	b.inner = bs
}

func (b *PacketBuffer) Pos() int {
	return b.pos
}

func (b *PacketBuffer) Inner() [512]byte {
	return b.inner
}

func (b *PacketBuffer) Step(steps int) {
	b.pos += steps
}

func (b *PacketBuffer) Seek(pos int) {
	b.pos = pos 
}

func (b *PacketBuffer) Put(pos int, value byte) {
	// Mostly for testing.
	b.inner[pos] = value
}


func (b *PacketBuffer) MustRead() byte {
	if b.pos >= 512 {
		panic("read: overran buffer")
	}
	ret := b.inner[b.pos]
	b.pos += 1
	return ret
}

func (b *PacketBuffer) MustGet(pos int) byte {
	if b.pos >= 512 {
		panic("get: overran buffer")
	}
	return b.inner[pos]
}

func (b *PacketBuffer) MustGetRange(start int, length int) []byte {
	if b.pos >= 512 {
		panic("getRange: overran buffer")
	}
	return b.inner[start:(start + length)]
}

func (b *PacketBuffer) MustReadUInt16() uint16 {
	firstPart := b.MustRead()
	secondPart := b.MustRead()
	return (uint16(firstPart) << 8) | uint16(secondPart)
}

func (b *PacketBuffer) MustReadUInt32() uint32 {
	firstPart := uint32(b.MustRead()) << 24
	secondPart := uint32(b.MustRead()) << 16
	thirdPart := uint32(b.MustRead()) << 8
	fourthPart := uint32(b.MustRead())
	return (firstPart | secondPart | thirdPart | fourthPart) 
}

func (b *PacketBuffer) MustReadQualifiedName() string {

	jumped := false
	maxJumps := 5
	jumpsPerformed := 0

	resStr := "" 
	pos := b.Pos()
	delim := ""

	for ;; {
		if jumpsPerformed > maxJumps {
			panic("Limits of jumps exceeded")
		}
		length := b.MustGet(pos)
		// If len has to most significant big set, it represent a jump to some other jump in the packet...
		if (length & 0xC0) == 0xC0 {
			if !jumped {
				b.Seek(pos + 2)
			}

			b2 := uint16(b.MustGet(pos + 1))
			offset := (((uint16(length)) ^ 0xC0) << 8) | b2
			pos = int(offset)

			jumped = true
			jumpsPerformed += 1

			continue
		} else {
			pos += 1

			if length == 0 {
				break;
			}

			resStr += delim

			strBuffer := b.MustGetRange(pos, int(length))
			resStr += string(strBuffer)
			delim = "."
			pos += int(length)
		}
	}
	if !jumped {
		b.Seek(pos)
	}

	return resStr
}

func (b *PacketBuffer) mustWrite(v byte) {
	if b.pos >= 512 {
		panic("end of buffer")
	}
	b.inner[b.pos] = v
	b.pos += 1
}

func (b *PacketBuffer) MustWriteU8(v byte) {
	b.mustWrite(v)
}

func (b *PacketBuffer) MustWriteU16(v uint16) {
	b.mustWrite(byte(v >> 8))
	b.mustWrite(byte(v & 0xFF))
}

func (b *PacketBuffer) MustWriteU32(v uint32) {
	b.mustWrite(byte((v >> 24) & 0xFF))
	b.mustWrite(byte((v >> 16) & 0xFF))
	b.mustWrite(byte((v >> 8) & 0xFF))
	b.mustWrite(byte((v >> 0) & 0xFF))
	b.mustWrite(byte(v & 0xFF))
}

func (b *PacketBuffer) MustWriteQName(qualifiedName string) {
	splits := strings.Split(qualifiedName, ".")

	for _, label := range splits {
		length := len(label)
		if length > 63 {
			panic("label length > 63 characters.")
		}
		b.MustWriteU8(byte(length))
		for _, c := range label {
			b.MustWriteU8(byte(c))
		}
	}
	b.MustWriteU8(0)
}


