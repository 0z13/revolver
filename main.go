package main

import (
	"fmt"
	"io"
	"os"
	"revolver/cmd/buffer"
	dnspacket "revolver/cmd/header"
)
func main() {

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

	// Copy the file contents into the array
	copy(innerBuffer[:], fileContents)

	buffer := buffer.New()
	buffer.SetInner(innerBuffer)

    packet := dnspacket.New()
    packet.FromPacketBuffer(buffer)
    
    fmt.Println(packet)
}
