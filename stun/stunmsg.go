package stun

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"reflect"
)

const (
	StunMsgType_BindingRequest         uint16 = 0x0001
	StunMsgType_BindingIndication      uint16 = 0x0011
	StunMsgType_BindingSuccessResponse uint16 = 0x0101
	StunMsgType_BindingErrorResponse   uint16 = 0x0111
)

const StunMsgHeaderLength = 20
const StunMsgMagicCookie = 0x2112A442

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
		MagicCookie:   StunMsgMagicCookie,
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
			var x XorMappedAddress
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

func FiledMarshal(x interface{}) ([]byte, error) {
	var bin []byte

	// 获取s的指针的反射值对象
	val := reflect.ValueOf(x)

	// 获取指针指向的值（即结构体）
	val = val.Elem()

	// 确保我们处理的是结构体
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("val.Kind() != reflect.Struct:%v!=%v", val.Kind(), reflect.Struct)
	}
	// 遍历结构体的所有字段
	for i := 0; i < val.NumField(); i++ {
		// 获取字段
		valueField := val.Field(i)
		switch valueField.Kind() {
		case reflect.Uint16:
			tmp := make([]byte, 2)
			binary.BigEndian.PutUint16(tmp, uint16(valueField.Uint()))
			bin = append(bin, tmp...)
		case reflect.Uint32:
			tmp := make([]byte, 4)
			binary.BigEndian.PutUint32(tmp, uint32(valueField.Uint()))
			bin = append(bin, tmp...)
		default:
			return nil, fmt.Errorf("unkown filed type:%v", valueField.Kind())
		}
	}

	return bin, nil
}

func FiledUnMarshal(bin []byte, x interface{}) (err error) {
	index := 0

	// 获取s的指针的反射值对象
	val := reflect.ValueOf(x)

	// 获取指针指向的值（即结构体）
	val = val.Elem()

	// 确保我们处理的是结构体
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("val.Kind() != reflect.Struct:%v!=%v", val.Kind(), reflect.Struct)
	}
	// 遍历结构体的所有字段
	for i := 0; i < val.NumField(); i++ {
		if index >= len(bin) {
			return fmt.Errorf("index >= len(bin):%v>=%v", index, len(bin))
		}
		// 获取字段的值
		valueField := val.Field(i)

		// 你可以在这里对字段进行操作
		// 例如，你可以检查字段的类型，并根据类型执行不同的操作
		switch valueField.Kind() {
		case reflect.Uint16:
			if valueField.CanSet() {
				tmp := binary.BigEndian.Uint16(bin[index:])
				index += 2
				valueField.SetUint(uint64(tmp))
			}
		case reflect.Uint32:
			if valueField.CanSet() {
				tmp := binary.BigEndian.Uint32(bin[index:])
				index += 4
				valueField.SetUint(uint64(tmp))
			}
		default:
			return fmt.Errorf("unkown filed type:%v", valueField.Kind())
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
const AttrType_ChangeRequest uint16 = 0x0003

func GetAttrTypeString(t uint16) string {
	switch t {
	case AttrType_XorMappedAddress:
		return "AttrType_XorMappedAddress"
	case AttrType_ChangeRequest:
		return "AttrType_ChangeRequest"
	default:
		return "unknown stun message type"
	}
}

type XorMappedAddress struct {
	Type     uint16
	Length   uint16
	Family   uint16
	XPort    uint16
	XAddress uint32
}

func (x *XorMappedAddress) Init() {
	x.Type = AttrType_XorMappedAddress
	x.Length = 8
}

func (x *XorMappedAddress) GetType() uint16 {
	return x.Type
}

func (x *XorMappedAddress) GetLength() uint16 {
	return x.Length
}

func (x *XorMappedAddress) Marshal() ([]byte, error) {
	return FiledMarshal(x)
}

func (x *XorMappedAddress) UnMarshal(bin []byte) (err error) {
	return FiledUnMarshal(bin, x)
}

func (x *XorMappedAddress) GetIp() net.IP {
	originalIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(originalIP, x.XAddress^StunMsgMagicCookie)
	return originalIP
}

func (x *XorMappedAddress) GetPort() uint16 {
	return x.XPort ^ uint16(StunMsgMagicCookie>>16)
}

func (x *XorMappedAddress) String() string {
	var str string
	str = fmt.Sprintf("attrType(%v)%s", x.Type, GetAttrTypeString(x.Type))
	str += fmt.Sprintf(",attrLength(%v)", x.Length)
	str += fmt.Sprintf(",attrFamily(%v)", x.Family)
	str += fmt.Sprintf(",attrXPort(%v)", x.XPort)
	str += fmt.Sprintf(",attrXAddress(%v)", x.XAddress)
	return str
}

type ChangeRequest struct {
	Type   uint16
	Length uint16
	Flag   uint32
}

func (c *ChangeRequest) Init(changeIp, changePort bool) {
	c.Type = AttrType_ChangeRequest
	c.Length = 4
	if changeIp {
		c.Flag |= c.Flag & 0x00000004
	}
	if changePort {
		c.Flag |= c.Flag & 0x00000002
	}
}

func (c *ChangeRequest) GetType() uint16 {
	return c.Type
}

func (c *ChangeRequest) GetLength() uint16 {
	return c.Length
}

func (c *ChangeRequest) Marshal() ([]byte, error) {
	return FiledMarshal(c)
}

func (c *ChangeRequest) UnMarshal(bin []byte) (err error) {
	return FiledUnMarshal(bin, c)
}

func (c *ChangeRequest) IsChangeIp() bool {
	return c.Flag&0x00000004 != 0
}

func (c *ChangeRequest) IsChangePort() bool {
	return c.Flag&0x00000002 != 0
}

func (c *ChangeRequest) String() string {
	var str string
	str = fmt.Sprintf("attrType(%v)%s", c.Type, GetAttrTypeString(c.Type))
	str += fmt.Sprintf(",attrLength(%v)", c.Length)
	str += fmt.Sprintf(",IsChangeIp(%v)", c.IsChangeIp())
	str += fmt.Sprintf(",IsChangePort(%v)", c.IsChangePort())
	return str
}
