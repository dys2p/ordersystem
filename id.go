package main

import (
	"math/rand"
	"strings"
	"time"
)

const idBytes = "ABCDEFGHKLMNPQRSTUVWXYZ"
const idLen = 10

func init() {
	rand.Seed(time.Now().UnixNano())
}

func IsID(s string) bool {
	if len(s) != idLen {
		return false
	}
	for _, b := range []byte(s) {
		if strings.IndexByte(idBytes, b) == -1 {
			return false
		}
	}
	return true
}

func NewID() string {
	var id = make([]byte, idLen)
	for i := range id {
		id[i] = idBytes[rand.Intn(len(idBytes))]
	}
	return string(id)
}
