package main

import "revolver/cmd/buffer"

func main() {
    buffer := buffer.New()
	buffer.Put(0, 3)
	buffer.Put(1, 9)
	buffer.MustRead_uint16()
}
