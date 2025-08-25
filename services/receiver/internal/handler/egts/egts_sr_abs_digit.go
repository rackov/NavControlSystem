package egts

import (
	"errors"
)

// SrAbsDigSensData структура подзаписи типа EGTS_SR_ABS_DIG_SENS_DATA, которая применяется абонентским
// АСН->данных о состоянии одного дискретного входа
type SrAbsDigSensData struct {
	StateNumber byte `json:"StateNumber"`
	Number      byte `json:"Number"`
}

// Decode разбирает байты в структуру подзаписи
func (e *SrAbsDigSensData) Decode(content []byte) error {
	if len(content) < int(e.Length()) {
		return errors.New("Некорректный размер данных")
	}
	e.StateNumber = content[0]
	e.Number = content[1]
	return nil
}

// Encode преобразовывает подзапись в набор байт
func (e *SrAbsDigSensData) Encode() ([]byte, error) {
	return []byte{
		byte(e.StateNumber),
		byte(e.StateNumber),
	}, nil
}

// Length получает длинну закодированной подзаписи
func (e *SrAbsDigSensData) Length() uint16 {
	return 2
}
