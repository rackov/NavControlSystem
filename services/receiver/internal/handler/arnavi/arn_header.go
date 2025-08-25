package arnavi

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// организация сессии
type HeadOne struct {
	Signature byte   `json:"signature"`
	Version   byte   `json:"version"`
	IdImei    uint64 `json:"id_imei"`
	ExtId     uint64 `json:"ext_id"`
}

func (e *HeadOne) Decode(content []byte) error {
	var (
		err error
	)
	buf := bytes.NewReader(content)

	// получаем преффикс
	if e.Signature, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось получить сигнатуру %v", err)
	}
	if e.Signature != SegnedHeader {
		return fmt.Errorf("протокол не arnavi sig not FF from %X", e.Signature)
	}
	// получаем версию протокола
	// 0x22 - HEADER1 - 5 don't support,
	// 0x23 (GPRS) or 0x25 (WIFI) - HEADER2
	// 0x24 - HEADER3 EXT ID
	if e.Version, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось получить версию протокола %v", err)
	}

	switch e.Version {
	case 0x23:
		tmp := make([]byte, 8)
		if _, err = buf.Read(tmp); err != nil {
			return fmt.Errorf("не удалось получить id_emei")
		}
		e.IdImei = binary.LittleEndian.Uint64(tmp)
	case 0x24:
		tmp := make([]byte, 8)
		if _, err = buf.Read(tmp); err != nil {
			return fmt.Errorf("не удалось получить id_emei")
		}
		e.IdImei = binary.LittleEndian.Uint64(tmp)
		if _, err = buf.Read(tmp); err != nil {
			return fmt.Errorf("не удалось получить ext_id")
		}
		e.ExtId = binary.LittleEndian.Uint64(tmp)
	default:
		return fmt.Errorf("версия протокола %X не поддерживается", e.Version)
	}
	return err
}

func (e *HeadOne) Encode() ([]byte, error) {
	var (
		result []byte
		err    error
	)
	buf := new(bytes.Buffer)
	if err = buf.WriteByte(e.Signature); err != nil {
		return result, fmt.Errorf("не удалось записать сигнатуру")
	}
	if err = buf.WriteByte(e.Version); err != nil {
		return result, fmt.Errorf("не удалось записать версию")
	}
	switch e.Version {
	case 0x23:
		tmp := make([]byte, 8)
		binary.LittleEndian.PutUint64(tmp, e.IdImei)
		if _, err = buf.Write(tmp); err != nil {
			return result, fmt.Errorf("не удалось записать id_emei")
		}
	case 0x24:
		tmp := make([]byte, 8)
		binary.LittleEndian.PutUint64(tmp, e.IdImei)
		if _, err = buf.Write(tmp); err != nil {
			return result, fmt.Errorf("не удалось записать id_emei")
		}
		binary.LittleEndian.PutUint64(tmp, e.ExtId)
		if _, err = buf.Write(tmp); err != nil {
			return result, fmt.Errorf("не удалось записть ext_id")
		}

	default:
		return result, fmt.Errorf("версия протокола %X не поддерживается", e.Version)
	}
	result = buf.Bytes()
	return result, err

}

// Length получает длинну закодированной подзаписи
func (e *HeadOne) Length() uint16 {
	var result uint16

	if recBytes, err := e.Encode(); err != nil {
		result = uint16(0)
	} else {
		result = uint16(len(recBytes))
	}
	return result
}
