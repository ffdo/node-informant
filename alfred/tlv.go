package alfred

import (
	"encoding/binary"
	"fmt"
)

type TLVType uint8

const (
	PUSH_DATA TLVType = iota
	ANNOUNCE_MASTER
	REQUEST
	STATUS_TXEND
	STATUS_ERROR
	MODESWITCH
)

type AlfredTLV struct {
	Type   TLVType
	Length uint16
	Data   []byte
}

func NewAlfredTLV(payload Payload) (tlv AlfredTLV, err error) {
	data, err := payload.Marshall()
	tlv = AlfredTLV{
		Type:   payload.TLVType(),
		Length: uint16(len(data)),
		Data:   data,
	}
	return
}

func NewRequestTLV(requestType uint8) (tlv AlfredTLV) {
	request := NewRequest(requestType)
	tlv, _ = NewAlfredTLV(request)
	return
}

func (tlv AlfredTLV) Marshall() (data []byte, err error) {
	data = make([]byte, 4)
	data[0] = byte(tlv.Type)
	data[1] = byte(0x00)
	binary.BigEndian.PutUint16(data[2:], tlv.Length)
	data = append(data, tlv.Data...)

	return
}

func UnmarshallTLVHeader(data []byte) (tlv AlfredTLV, err error) {
	length := binary.BigEndian.Uint16(data[2:4])
	tlv = AlfredTLV{
		Type:   TLVType(data[0]),
		Length: uint16(length),
	}
	return
}

func Unmarshall(data []byte) (tlv AlfredTLV, err error) {
	tlv, err = UnmarshallTLVHeader(data)
	if err != nil {
		return
	}
	if uint64(len(data)-4) != uint64(tlv.Length) {
		err = fmt.Errorf("Payload length %d is longer than specified length %d", len(data)-4, tlv.Length)
		return
	}
	tlv.Data = data[4:]
	return
}
