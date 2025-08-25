package egts

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// SrAdSensorsData структура подзаписи типа EGTS_SR_AD_SENSORS_DATA, которая применяется абонентским
// терминалом для передачи на аппаратно-программный комплекс информации о состоянии дополнительных
// дискретных и аналоговых входов
type SrAdSensorsData struct {
	DigitalInputsOctetExists     byte      `json:"DIOE"`
	DigitalOutputs               byte      `json:"DOUT"`
	AnalogSensorFieldExists      byte      `json:"ASFE"`
	AdditionalDigitalInputsOctet [8]byte   `json:"ADIO"`
	AnalogSensors                [8]uint32 `json:"ANS"`
}

// Decode разбирает байты в структуру подзаписи
func (e *SrAdSensorsData) Decode(content []byte) error {
	var (
		err           error
		flags         byte
		analogSensVal []byte
	)
	buf := bytes.NewReader(content)

	//байт флагов
	if flags, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("Не удалось получить байт цифровых выходов ad_sesor_data: %v", err)
	}
	e.DigitalInputsOctetExists = flags

	if e.DigitalOutputs, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("Не удалось получить битовые флаги дискретных выходов: %v", err)
	}

	if flags, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("Не удалось получить байт аналоговых выходов ad_sesor_data: %v", err)
	}
	e.AnalogSensorFieldExists = flags

	for i := 0; i < 8; i++ {
		if (e.DigitalInputsOctetExists >> i & 0x1) == 1 {
			if e.AdditionalDigitalInputsOctet[i], err = buf.ReadByte(); err != nil {
				return fmt.Errorf("Не удалось получить байт показания ADIO1: %v", err)
			}
		}
	}
	tmpBuf := make([]byte, 3)

	for i := 0; i < 8; i++ {
		if (e.AnalogSensorFieldExists >> i & 0x1) == 1 {
			if _, err = buf.Read(tmpBuf); err != nil {
				return fmt.Errorf("Не удалось получить показания ANS1: %v", err)
			}
			analogSensVal = append(tmpBuf, 0x00)
			e.AnalogSensors[i] = binary.LittleEndian.Uint32(analogSensVal)
		}
	}

	return err
}

// Encode преобразовывает подзапись в набор байт
func (e *SrAdSensorsData) Encode() ([]byte, error) {
	var (
		err    error
		result []byte
	)

	buf := new(bytes.Buffer)

	if err = buf.WriteByte(e.DigitalInputsOctetExists); err != nil {
		return result, fmt.Errorf("Не удалось записать байт флагов ext_pos_data: %v", err)
	}

	if err = buf.WriteByte(e.DigitalOutputs); err != nil {
		return result, fmt.Errorf("Не удалось записать битовые флаги дискретных выходов: %v", err)
	}

	if err = buf.WriteByte(e.AnalogSensorFieldExists); err != nil {
		return result, fmt.Errorf("Не удалось записать байт байт аналоговых выходов ad_sesor_data: %v", err)
	}

	for i := 0; i <= 7; i++ {
		if (e.DigitalInputsOctetExists >> i & 0x1) == 1 {
			if err = buf.WriteByte(e.AdditionalDigitalInputsOctet[i]); err != nil {
				return result, fmt.Errorf("Не удалось записать байт показания ADIO1: %v", err)
			}
		}
	}

	for i := 0; i <= 7; i++ {
		if (e.AnalogSensorFieldExists >> i & 0x1) == 1 {
			sensVal := make([]byte, 4)
			binary.LittleEndian.PutUint32(sensVal, e.AnalogSensors[i])
			if _, err = buf.Write(sensVal[:3]); err != nil {
				return result, fmt.Errorf("Не удалось запистаь показания ANS1: %v", err)
			}
		}
	}

	result = buf.Bytes()
	return result, err
}

// Length получает длинну закодированной подзаписи
func (e *SrAdSensorsData) Length() uint16 {
	var result uint16

	if recBytes, err := e.Encode(); err != nil {
		result = uint16(0)
	} else {
		result = uint16(len(recBytes))
	}

	return result
}
