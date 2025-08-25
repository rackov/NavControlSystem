package arnavi

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ScanPaked struct {
	//start sign PACKAGE 5B
	StartSign byte
	// parcel number from 0x01 to 0xFB
	Id byte
}
type ScanPacket struct {
	TypeContent  byte   `json:"type_content"`
	LengthPacket uint16 `json:"length_packet"`
}

func (sc *ScanPaked) Decode(rec []byte) error {
	var err error
	buf := bytes.NewBuffer(rec)
	err = binary.Read(buf, binary.LittleEndian, sc)
	if sc.StartSign != SigPackStart {
		return fmt.Errorf("не верная сигнатура packed: %X", sc.StartSign)
	}
	if (sc.Id == 0) || (sc.Id > 0xFB) {
		return fmt.Errorf("выход Id за пределы")
	}
	return err
}
func (sc *ScanPacket) Decode(rec []byte) error {
	var err error
	buf := bytes.NewBuffer(rec)
	err = binary.Read(buf, binary.LittleEndian, sc)
	return err
}
