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

/*
0                   1                   2                   3
	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|         Type                  |            Length             |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         Value (variable)                      ...
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

type TLV struct {
	Type   uint16
	Length uint16
	Value  []byte
}

type Attr interface {
	GetType() uint16
	GetLength() uint16
	Marshal() []byte
	UnMarshal([]byte) error
}

/*
0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0|     STUN Message Type     |         Message Length        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         Magic Cookie                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|                     Transaction ID (96 bits)                  |
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

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

/*
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 0 0 0|    Family     |         X-Port                |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|                   X-Address (Obfuscated IP)                   |
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

const AttrType_XorMappedAddress uint16 = 0x0020

type XorMappedAddressValue struct {
	Type     uint16
	Length   uint16
	Family   uint16
	XPort    uint16
	XAddress uint32
}

func (x *XorMappedAddressValue) GetType() uint16 {
	return x.Type
}

func (x *XorMappedAddressValue) GetLength() uint16 {
	return x.Length
}

func (x *XorMappedAddressValue) Marshal() (t *TLV) {
	t.Type = AttrType_XorMappedAddress
	t.Length = 8
	t.Value = make([]byte, t.Length)

	index := 0
	binary.BigEndian.PutUint16(t.Value[index:], x.Family)
	index += 2
	binary.BigEndian.PutUint16(t.Value[index:], x.XPort)
	index += 2
	binary.BigEndian.PutUint32(t.Value[index:], x.XAddress)
	return t
}

func (x *XorMappedAddressValue) UnMarshal(t *TLV) (err error) {
	if t.Type != AttrType_XorMappedAddress {
		return fmt.Errorf("type(%v) is not AttrType_XorMappedAddress(%v)", t.Type, AttrType_XorMappedAddress)
	}
	if t.Length != 8 {
		return fmt.Errorf("length(%v) is not 8", t.Length)
	}

	index := 0
	x.Family = binary.BigEndian.Uint16(t.Value[index:])
	index += 2
	x.XPort = binary.BigEndian.Uint16(t.Value[index:])
	index += 2
	x.XAddress = binary.BigEndian.Uint32(t.Value[index:])
	return nil
}
