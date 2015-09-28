package alfred

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io/ioutil"
	"math/rand"
)

const MaxUint16 = ^uint16(0)

type Payload interface {
	Marshall() ([]byte, error)
	TLVType() TLVType
}

type Request struct {
	Type          uint8
	TransactionId uint16
}

func (r Request) Marshall() (data []byte, err error) {
	data = make([]byte, 3)
	data[0] = r.Type
	binary.BigEndian.PutUint16(data[1:3], uint16(r.TransactionId))
	return
}

func (r Request) TLVType() TLVType {
	return REQUEST
}

func NewRequest(requestType uint8) (request Request) {
	return Request{
		Type:          requestType,
		TransactionId: uint16(rand.Intn(int(MaxUint16))),
	}
}

type Status struct {
	TransactionId uint16
	PacketCount   uint16
}

func (s Status) Marshall() (data []byte, err error) {
	data = make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:1], s.PacketCount)
	binary.BigEndian.PutUint16(data[2:3], s.TransactionId)

	return
}

func (s Status) TLVType() TLVType {
	return STATUS_TXEND
}

const AlfredDataHeaderLength uint16 = 10

type AlfredData struct {
	SourceMac [6]byte
	Type      uint8
	Version   uint8
	Length    uint16
	Data      []byte
}

func (a AlfredData) DecompressData() (data []byte, err error) {
	ir := bytes.NewReader(a.Data)
	r, err := gzip.NewReader(ir)
	defer r.Close()
	if err != nil {
		return
	}
	data, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}
	return
}

func UnmarshallAlfredData(in []byte) (data AlfredData, err error) {
	var mac [6]byte
	copy(mac[:], in[0:5])
	length := binary.BigEndian.Uint16(in[8:10])

	data = AlfredData{
		SourceMac: mac,
		Type:      in[6],
		Version:   in[7],
		Length:    uint16(length),
		Data:      make([]byte, length),
	}
	copy(data.Data, in[10:])
	return
}

func (a AlfredData) Marshall() (data []byte, err error) {
	data = make([]byte, 10)
	copy(data[0:5], a.SourceMac[:])
	data[6] = a.Type
	data[7] = a.Version
	binary.BigEndian.PutUint16(data[8:9], a.Length)
	data = append(data, a.Data...)
	return
}

type PushData struct {
	TransactionId uint16
	Sequence      uint16
	Data          []AlfredData
}

func (p PushData) Marshall() (data []byte, err error) {
	data = make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], p.TransactionId)
	binary.BigEndian.PutUint16(data[2:4], p.Sequence)
	for _, alfredData := range p.Data {
		temp, dataErr := alfredData.Marshall()
		if dataErr != nil {
			err = dataErr
			return
		}
		data = append(data, temp...)
	}
	return
}

func UnmarshallPushData(data []byte) (pData PushData, err error) {
	transactionId := binary.BigEndian.Uint16(data[0:2])
	sequence := binary.BigEndian.Uint16(data[2:4])
	pData = PushData{
		TransactionId: transactionId,
		Sequence:      sequence,
		Data:          make([]AlfredData, 0, 20),
	}
	var pointer uint16 = 4
	for pointer < uint16(len(data)) {
		var mac [6]byte
		copy(mac[:], data[pointer:pointer+6])
		length := binary.BigEndian.Uint16(data[pointer+8 : pointer+10])
		dataBytes := make([]byte, 0, length)
		dataBytes = append(dataBytes, data[pointer+AlfredDataHeaderLength:pointer+AlfredDataHeaderLength+length]...)
		alfredData := AlfredData{
			SourceMac: mac,
			Type:      data[pointer+6],
			Version:   data[pointer+7],
			Length:    length,
			Data:      dataBytes,
		}
		pData.Data = append(pData.Data, alfredData)
		pointer = pointer + length + AlfredDataHeaderLength
	}
	return
}

func (p PushData) TLVType() TLVType {
	return PUSH_DATA
}
