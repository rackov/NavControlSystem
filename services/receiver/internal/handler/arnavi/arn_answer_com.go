package arnavi

import (
	"bytes"
	"fmt"
)

type AnswerCom struct {
	StartSign byte `json:"start_sign"`
	IdPacked  byte `json:"id_packed"`
	CodeError byte `json:"code_error"`
	EndSign   byte `json:"end_sign"`
}

func (a *AnswerCom) Decode(answer []byte) error {
	var err error
	buf := bytes.NewReader(answer)

	if a.StartSign, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не возможно прочитать StartSign %v", err)
	}
	if a.StartSign != SigPackStart {
		return fmt.Errorf("неверная сигнатура StartSign %X", a.StartSign)
	}
	if a.IdPacked, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не возможно прочитать IdPacked %v", err)
	}
	if a.CodeError, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не возможно прочитать CodeError %v", err)
	}
	if a.EndSign, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не возможно прочитать StartSign %v", err)
	}
	if a.EndSign != SigPackEnd {
		return fmt.Errorf("неверная сигнатура StartSign %X", a.EndSign)
	}

	return err
}

func (a *AnswerCom) Encode() ([]byte, error) {
	var (
		err    error
		result []byte
	)
	buf := new(bytes.Buffer)
	buf.WriteByte(a.StartSign)
	buf.WriteByte(a.IdPacked)
	buf.WriteByte(a.CodeError)
	buf.WriteByte(a.EndSign)

	result = buf.Bytes()
	return result, err
}

func EncodeAnswerCom(IdPacked int, CodeEror int) ([]byte, error) {
	answ := AnswerCom{
		StartSign: SigPackStart,
		IdPacked:  byte(IdPacked),
		CodeError: byte(CodeEror),
		EndSign:   SigPackEnd,
	}
	return answ.Encode()
}
