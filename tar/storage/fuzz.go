// +build gofuzz

package storage

import (
	"bytes"
	"log"
)

func Fuzz(data []byte) int {
	unpacker := NewJSONUnpacker(bytes.NewReader(data))

	for {
		entry, err := unpacker.Next()
		if err != nil {
			log.Println(err)
			return 0
		}

		log.Printf("%v", entry)
	}
	return 1
}
