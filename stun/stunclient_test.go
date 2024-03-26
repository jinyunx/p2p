package stun

import (
	"log"
	"testing"
)

func TestBindingRequest(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	BindingRequest("stun.qq.com:3478")
}
