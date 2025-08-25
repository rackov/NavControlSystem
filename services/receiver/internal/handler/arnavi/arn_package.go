package arnavi

import (
	"bytes"
	"fmt"
)

// PACKAGE - structure and description of the packages on the device
type PackageS struct {
	//start sign PACKAGE 5B
	StartSign byte
	// parcel number from 0x01 to 0xFB
	Id     byte
	Packet BinaryData
	// end sign PACKAGE	 5D
	EndSign byte
}

func (p *PackageS) Decode(pac []byte) error {
	var err error
	buf := bytes.NewReader(pac)
	if p.StartSign, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось прочитать начальную сигнатуру %d", SigPackStart)
	}
	if p.StartSign != SigPackStart {
		return fmt.Errorf("сигнатура %d не верная %d", SigPackStart, p.StartSign)
	}
	if p.Id, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось считать parcel number (id)")
	}

	// data packet

	if p.EndSign, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось прочитать конечную сигнатуру %d", SigPackEnd)
	}
	if p.EndSign != SigPackStart {
		return fmt.Errorf("сигнатура %d не верная %d", SigPackEnd, p.EndSign)
	}

	return err
}

func (p *PackageS) Encode() ([]byte, error) {
	var (
		result []byte
		err    error
	)
	buf := new(bytes.Buffer)

	if err = buf.WriteByte(p.StartSign); err != nil {
		return result, fmt.Errorf("не удалось записать сигнатуру %v", err)
	}
	if err = buf.WriteByte(p.Id); err != nil {
		return result, fmt.Errorf("не удалось записать id пакета %v", err)
	}
	// data packet
	if err = buf.WriteByte(p.EndSign); err != nil {
		return result, fmt.Errorf("не удалось записать сигнатуру %v", err)

	}
	result = buf.Bytes()
	return result, err
}

// Length получает длинну закодированной подзаписи
func (p *PackageS) Length() uint16 {
	var result uint16

	if recBytes, err := p.Encode(); err != nil {
		result = uint16(0)
	} else {
		result = uint16(len(recBytes))
	}
	return result
}
