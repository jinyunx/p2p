package stun

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
)

const (
	StunMsgType_BindingRequest         uint16 = 0x0001
	StunMsgType_BindingIndication      uint16 = 0x0011
	StunMsgType_BindingSuccessResponse uint16 = 0x0101
	StunMsgType_BindingErrorResponse   uint16 = 0x0111
)

const StunMsgHeaderLength = 20

type TLV struct {
	Type   uint16
	Length uint16
	Value  []byte
}

type StunMsg struct {
	StunMsgType   uint16
	MsgLength     uint16
	MagicCookie   uint32
	TransactionID [12]byte
	Attrs         []TLV
}

func InitStunMsg(stunMsgType uint16, attrs []TLV) (*StunMsg, error) {
	s := &StunMsg{
		StunMsgType:   stunMsgType,
		MsgLength:     0,
		MagicCookie:   0x2112A442,
		TransactionID: [12]byte{},
		Attrs:         attrs,
	}
	_, err := rand.Read(s.TransactionID[:])
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for _, v := range s.Attrs {
		s.MsgLength += uint16(4 + len(v.Value))
	}
	return s, nil
}

func (s *StunMsg) UnMarshal(bin []byte) error {
	if len(bin) < StunMsgHeaderLength {
		return fmt.Errorf("len(bin) < StunMsgHeaderLength:%v<%v", len(bin), StunMsgHeaderLength)
	}
	index := 0
	s.StunMsgType = binary.BigEndian.Uint16(bin[index:])
	index += 2
	s.MsgLength = binary.BigEndian.Uint16(bin[index:])
	index += 2
	s.MagicCookie = binary.BigEndian.Uint32(bin[index:])
	index += 4
	copy(s.TransactionID[:], bin[index:index+len(s.TransactionID)])
	index += len(s.TransactionID)
	log.Println(s)

	for index < len(bin) {
		if len(bin)-index < 4 {
			return fmt.Errorf("len(bin)-index < 4:%v<4", len(bin)-index)
		}
		tlv := &TLV{}
		tlv.Type = binary.BigEndian.Uint16(bin[index:])
		index += 2
		tlv.Length = binary.BigEndian.Uint16(bin[index:])
		index += 2
		tlv.Value = make([]byte, tlv.Length)
		copy(tlv.Value, bin[index:index+int(tlv.Length-1)])
		index += int(tlv.Length)

		s.Attrs = append(s.Attrs, *tlv)
	}
	return nil
}

func (s *StunMsg) Marshal() []byte {
	totalLength := StunMsgHeaderLength + s.MsgLength
	bin := make([]byte, totalLength)

	index := 0
	binary.BigEndian.PutUint16(bin[index:], s.StunMsgType)
	index += 2
	binary.BigEndian.PutUint16(bin[index:], s.MsgLength)
	index += 2
	binary.BigEndian.PutUint32(bin[index:], s.MagicCookie)
	index += 4
	copy(bin[index:index+len(s.TransactionID)], s.TransactionID[:])
	index += len(s.TransactionID)

	for _, v := range s.Attrs {
		binary.BigEndian.PutUint16(bin[index:], v.Type)
		index += 2
		binary.BigEndian.PutUint16(bin[index:], v.Length)
		index += 2
		copy(bin[index:index+len(v.Value)-1], v.Value)
		index += len(v.Value)
	}
	return bin
}
