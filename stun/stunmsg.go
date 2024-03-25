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

func GetStunMsgTypeString(t uint16) string {
	switch t {
	case StunMsgType_BindingRequest:
		return "StunMsgType_BindingRequest"
	case StunMsgType_BindingIndication:
		return "StunMsgType_BindingIndication"
	case StunMsgType_BindingSuccessResponse:
		return "StunMsgType_BindingSuccessResponse"
	case StunMsgType_BindingErrorResponse:
		return "StunMsgType_BindingErrorResponse"
	default:
		return "unknown stun message type"
	}
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
	Attrs         []Attr
}

func InitStunMsg(stunMsgType uint16, attrs []Attr) (*StunMsg, error) {
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
		s.MsgLength += 4 + v.GetLength()
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

	attrs, err := UnMarshalAttrs(bin[index:])
	if err != nil {
		return err
	}
	s.Attrs = attrs
	return nil
}

func (s *StunMsg) Marshal() ([]byte, error) {
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

	err := MarshalAttrs(s.Attrs, bin[index:])
	if err != nil {
		return nil, err
	}
	return bin, nil
}

func (s *StunMsg) String() string {
	var str string
	str = fmt.Sprintf("StunMsgType(%v)%s", s.StunMsgType, GetStunMsgTypeString(s.StunMsgType))
	str += fmt.Sprintf(",MsgLength(%v)", s.MsgLength)
	str += fmt.Sprintf(",MagicCookie(%v)", s.MagicCookie)
	str += fmt.Sprintf(",TransactionID(%v)", s.TransactionID)

	for _, a := range s.Attrs {
		str += "\n"
		str += a.String()
	}
	return str
}

/*
0                   1                   2                   3
	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|         Type                  |            Length             |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         Value (variable)                      ...
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

type Attr interface {
	GetType() uint16
	GetLength() uint16
	Marshal() ([]byte, error)
	UnMarshal([]byte) error
	String() string
}

func UnMarshalAttrs(bin []byte) ([]Attr, error) {
	var attrs []Attr

	index := 0
	for {
		if len(bin[index:]) == 0 {
			break
		}
		if len(bin[index:]) < 4 {
			return nil, fmt.Errorf("len(%v) < 4", len(bin[index:]))
		}

		t := binary.BigEndian.Uint16(bin[index:])
		l := binary.BigEndian.Uint16(bin[index+2:])
		length := int(l + 4)

		switch t {
		case AttrType_XorMappedAddress:
			var x XorMappedAddressValue
			x.Init()
			err := x.UnMarshal(bin[index : index+length])
			if err != nil {
				return nil, err
			}
			attrs = append(attrs, &x)
			index += length
		default:
			return nil, fmt.Errorf("unknow type(%v)", t)
		}
	}
	return attrs, nil
}

func MarshalAttrs(attrs []Attr, bin []byte) error {
	for _, a := range attrs {
		switch a.GetType() {
		case AttrType_XorMappedAddress:
			b, err := a.Marshal()
			if err != nil {
				return err
			}
			if len(b) > len(bin) {
				fmt.Errorf("len(b)%v > len(bin)%v", len(b), len(bin))
			}
			copy(bin, b)
		default:
			return fmt.Errorf("unknow type(%v)", a.GetType())
		}
	}
	return nil
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

func GetAttrTypeString(t uint16) string {
	switch t {
	case AttrType_XorMappedAddress:
		return "AttrType_XorMappedAddress"
	default:
		return "unknown stun message type"
	}
}

type XorMappedAddressValue struct {
	Type     uint16
	Length   uint16
	Family   uint16
	XPort    uint16
	XAddress uint32
}

func (x *XorMappedAddressValue) Init() {
	x.Type = AttrType_XorMappedAddress
	x.Length = 8
}

func (x *XorMappedAddressValue) GetType() uint16 {
	return x.Type
}

func (x *XorMappedAddressValue) GetLength() uint16 {
	return x.Length
}

func (x *XorMappedAddressValue) Marshal() ([]byte, error) {
	bin := make([]byte, 4+x.Length)
	index := 0
	binary.BigEndian.PutUint16(bin[index:], x.Type)
	index += 2

	binary.BigEndian.PutUint16(bin[index:], x.Length)
	index += 2

	binary.BigEndian.PutUint16(bin[index:], x.Family)
	index += 2

	binary.BigEndian.PutUint16(bin[index:], x.XPort)
	index += 2

	binary.BigEndian.PutUint32(bin[index:], x.XAddress)
	index += 4

	return bin, nil
}

func (x *XorMappedAddressValue) UnMarshal(bin []byte) (err error) {
	if len(bin) != 12 {
		return fmt.Errorf("len(bin)%v != 12", len(bin))
	}

	index := 0
	t := binary.BigEndian.Uint16(bin[index:])
	if t != x.Type {
		return fmt.Errorf("type(%v) != AttrType_XorMappedAddress", t)
	}
	index += 2

	l := binary.BigEndian.Uint16(bin[index:])
	if l != 8 {
		return fmt.Errorf("length(%v) != 8", l)
	}
	index += 2

	x.Type = t
	x.Length = l

	x.Family = binary.BigEndian.Uint16(bin[index:])
	index += 2
	x.XPort = binary.BigEndian.Uint16(bin[index:])
	index += 2
	x.XAddress = binary.BigEndian.Uint32(bin[index:])
	return nil
}

func (x *XorMappedAddressValue) String() string {
	var str string
	str = fmt.Sprintf("attrType(%v)%s", x.Type, GetAttrTypeString(x.Type))
	str += fmt.Sprintf(",attrLength(%v)", x.Length)
	str += fmt.Sprintf(",attrFamily(%v)", x.Family)
	str += fmt.Sprintf(",attrXPort(%v)", x.XPort)
	str += fmt.Sprintf(",attrXAddress(%v)", x.XAddress)
	return str
}
