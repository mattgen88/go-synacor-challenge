package main

import (
	"encoding/binary"
	"log"
	"os"

	"github.com/mattgen88/go-synacor-challenge/virtualmachine"
)

func loadChallenge() []uint16 {
	data, err := os.Open("challenge.bin")
	if err != nil {
		panic(err)
	}
	defer data.Close()

	var memory []uint16

	for {
		var i uint16
		readErr := binary.Read(data, binary.LittleEndian, &i)
		if readErr != nil {
			return memory
		}
		memory = append(memory, i)
	}
}

func main() {

	debugLog, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer debugLog.Close()

	stateLog, err := os.OpenFile("state.json", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer stateLog.Close()

	vm := virtualmachine.New(loadChallenge())
	vm.DebugFD = debugLog
	vm.StateFD = stateLog

	vm.Start()
	log.Println("Goodbye")
}
