package arnavi

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type PacketS struct {
	TypeContent  byte       `json:"type_content"`
	LengthPacket uint16     `json:"length_packet"`
	TimePacket   uint32     `json:"time_packet"`
	Data         BinaryData `json:"data"`
	CheckSum     byte       `json:"check_sum"`
}

func (p *PacketS) Decode(rec []byte) error {
	var (
		err error
	)
	buf := bytes.NewBuffer(rec)
	if p.TypeContent, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось считать тип пакета %v", err)
	}
	tmp := make([]byte, 2)
	if _, err = buf.Read(tmp); err != nil {
		return fmt.Errorf("не удалось считать длинну пакета %v", err)
	}
	p.LengthPacket = binary.LittleEndian.Uint16(tmp)
	timeTmp := make([]byte, 4)
	if _, err = buf.Read(timeTmp); err != nil {
		return fmt.Errorf("не удалось считать время пакета %v", err)
	}
	p.TimePacket = binary.LittleEndian.Uint32(timeTmp)

	packetBuf := buf.Next(int(p.LengthPacket))
	switch p.TypeContent {
	case PackTagsType:
		p.Data = &TagsData{}
	}
	p.Data.Decode(packetBuf)

	if p.CheckSum, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось считать crc пакета %v", err)
	}
	// chk := Crc_sum(rec[3 : 6+p.LengthPacket])
	// if p.CheckSum != chk {
	// 	return fmt.Errorf("неверная контрольная сумма в пакете"+
	// 		" %X, подсчитано %X   1-byte %X  len: %d  последний байт %X",
	// 		p.CheckSum, chk, rec[3], p.LengthPacket, rec[3+p.LengthPacket])
	// }
	return err
}

func (p *PacketS) Encode() ([]byte, error) {
	var (
		result []byte
		dbuf   []byte
		err    error
	)
	buf := new(bytes.Buffer)
	switch p.Data.(type) {
	case *TagsData:
		p.TypeContent = PackTagsType

	default:
		return result, fmt.Errorf("не известен код для данного типа пакета")
	}
	if dbuf, err = p.Data.Encode(); err != nil {
		return result, fmt.Errorf("не записаны данные packet %v", err)
	}

	p.LengthPacket = uint16(len(dbuf))

	buf.WriteByte(p.TypeContent)

	if err = binary.Write(buf, binary.LittleEndian, p.LengthPacket); err != nil {
		return result, fmt.Errorf("не записаны размер packet %v", err)
	}
	if err = binary.Write(buf, binary.LittleEndian, p.TimePacket); err != nil {
		return result, fmt.Errorf("не записаны время packet %v", err)
	}
	buf.Write(dbuf)

	p.CheckSum = Crc_sum(buf.Bytes()[3 : 6+p.LengthPacket])
	buf.WriteByte(p.CheckSum)

	result = buf.Bytes()
	return result, err
}

func (p *PacketS) Length() uint16 {
	var result uint16

	if recBytes, err := p.Encode(); err != nil {
		result = uint16(0)
	} else {
		result = uint16(len(recBytes))
	}
	return result
}
