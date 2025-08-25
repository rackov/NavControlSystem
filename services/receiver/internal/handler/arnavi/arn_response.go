package arnavi

import (
	"bytes"
	"fmt"
	"time"
)

type ResCom struct {
	StartSign byte       `json:"start_sign"`
	Size      byte       `json:"size"`
	CodeCom   byte       `json:"code_com"`
	CheckSum  byte       `json:"check_sum"`
	Data      BinaryData `json:"data"`
	EndSign   byte       `json:"end_sign"`
}

func (r *ResCom) Decode(rec []byte) error {
	var (
		err error
	)
	buf := bytes.NewBuffer(rec)

	if r.StartSign, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось прочитать начальную сигнатуру пакета")
	}
	if r.StartSign != SignedStart {
		return fmt.Errorf("не верная сигнатура команды/пакета")
	}
	if r.Size, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось прочитать размер пакета")
	}
	if r.CodeCom, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось прочитать код команды")
	}
	if r.Size > 0 {
		if r.CheckSum, err = buf.ReadByte(); err != nil {
			return fmt.Errorf("не удалось прочитать контрольную сумму пакета")
		}
		if r.CheckSum != Crc_sum(rec[AnswerStart:AnswerStart+r.Size]) {
			return fmt.Errorf("неверная контрольная сумма")
		}
	} else {
		goto EndCom
	}

	switch r.CodeCom {
	case ConfirmationHeaderType:
		r.Data = &ConfirmationHeader{}
	default:
		return fmt.Errorf("не известный тип подзаписи: %d. Длина: %d. Содержимое: %X", r.CodeCom, r.Size, rec)

	}

	if err = r.Data.Decode(rec[AnswerStart : AnswerStart+r.Size]); err != nil {
		return err
	}
	buf.Next(int(r.Size))

EndCom:
	if r.EndSign, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось прочитать конечную сигнатуру пакета")
	}
	if r.EndSign != SignedEnd {
		return fmt.Errorf("не верная конечную сигнатуру пакет")
	}

	return err
}

func (r *ResCom) Encode() ([]byte, error) {
	var (
		result []byte
		err    error
		datab  []byte
	)
	buf := new(bytes.Buffer)

	buf.WriteByte(SignedStart)
	if r.Data != nil {
		r.Size = byte(r.Data.Length())
	} else {
		r.Size = 0
	}
	buf.WriteByte(r.Size)

	switch r.Data.(type) {
	case *ConfirmationHeader:
		r.CodeCom = ConfirmationHeaderType
	case nil:

	default:
		return result, fmt.Errorf("не известен код для данного типа пакета")
	}

	buf.WriteByte(r.CodeCom)

	if r.Size > 0 {

		datab, err = r.Data.Encode()
		if err != nil {
			return result, err
		}
		buf.WriteByte(Crc_sum(datab))
		buf.Write(datab)

	}

	if err = buf.WriteByte(r.EndSign); err != nil {
		return result, err
	}
	result = buf.Bytes()
	return result, err
}

func AnswerHeader() []byte {
	currentTime := time.Now()
	answ := ResCom{
		StartSign: SignedStart,
		CodeCom:   0,
		EndSign:   SignedEnd,
		Data:      &ConfirmationHeader{TimeAnswer: uint32(currentTime.Unix())},
	}
	b, _ := answ.Encode()
	return b
}

func AnswerPacked(id_packed int) ([]byte, error) {
	answ := ResCom{
		StartSign: SignedStart,
		CodeCom:   byte(id_packed),
		EndSign:   SignedEnd,
	}
	return answ.Encode()
}
