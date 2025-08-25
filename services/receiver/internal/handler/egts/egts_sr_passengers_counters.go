package egts

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

//SrPassengersCountersData структура подзаписи типа EGTS_SR_PASSENGERS_COUNTERS,
//которая применяется абонентским терминалом для передачи на аппаратно-программный
//комплекс данных о показаниях счетчиков пассажиропотока
type SrPassengersCountersData struct {
	RawDataFlag               string              `json:"RawDataFlag"`
	DoorsPresented            string              `json:"DoorsPresented"`
	DoorsReleased             string              `json:"DoorsReleased"`
	ModuleAddress             uint16              `json:"ModuleAddress"`
	PassengersCountersData    []PassengersCounter `json:"PassengersCountersData"`
	PassengersCountersRawData []byte              `json:"PassengersCountersRawData"`
}

//Decode разбирает байты в структуру подзаписи
func (e *SrPassengersCountersData) Decode(content []byte) error {
	var (
		err     error
		byteBuf uint8
	)
	maBuf := make([]byte, 2)
	buf := bytes.NewReader(content)

	if byteBuf, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("Не удалось получить байт флагов EGTS_SR_PASSENGERS_COUNTERS: %v", err)
	}
	e.RawDataFlag = fmt.Sprintf("%08b", byteBuf)[7:]

	if byteBuf, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("Не удалось получить наличие счетчиков на дверях: %v", err)
	}
	e.DoorsPresented = fmt.Sprintf("%08b", byteBuf)

	if byteBuf, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("Не удалось получить двери, которые открывались и закрывались: %v", err)
	}
	e.DoorsReleased = fmt.Sprintf("%08b", byteBuf)

	if _, err = buf.Read(maBuf); err != nil {
		return fmt.Errorf("Не удалось получить адрес модуля: %v", err)
	}
	e.ModuleAddress = binary.LittleEndian.Uint16(maBuf)

	if e.RawDataFlag == "0" {
		var in, out uint8
		for i := 1; i < 9; i++ {
			if e.DoorsPresented[8-i:9-i] == "0" {
				continue
			}

			if in, err = buf.ReadByte(); err != nil {
				return fmt.Errorf("Не удалось получить количество вошедших пассажиров через дверь #%d: %v", i, err)
			}

			if out, err = buf.ReadByte(); err != nil {
				return fmt.Errorf("Не удалось получить количество вышедших пассажиров через дверь #%d: %v", i, err)
			}

			e.PassengersCountersData = append(e.PassengersCountersData, PassengersCounter{
				DoorNo: uint8(i),
				In:     in,
				Out:    out,
			})
		}
	} else {
		pcdRawBuf := make([]byte, buf.Len())
		if _, err = buf.Read(pcdRawBuf); err != nil {
			return fmt.Errorf("Не удалось получить данные счетчиков пассажиропотока в необработанном виде: %v", err)
		}
		e.PassengersCountersRawData = pcdRawBuf
	}

	return err
}

//Encode преобразовывает подзапись в набор байт
func (e *SrPassengersCountersData) Encode() ([]byte, error) {
	var (
		err    error
		flags  uint64
		dpr    uint64
		drl    uint64
		result []byte
	)
	maddrBuf := make([]byte, 2)
	buf := new(bytes.Buffer)

	if flags, err = strconv.ParseUint(e.RawDataFlag, 2, 8); err != nil {
		return result, fmt.Errorf("Не удалось сгенерировать байт флагов EGTS_SR_PASSENGERS_COUNTERS: %v", err)
	}
	if err = buf.WriteByte(uint8(flags)); err != nil {
		return result, fmt.Errorf("Не удалось записать байт флагов EGTS_SR_PASSENGERS_COUNTERS: %v", err)
	}

	if dpr, err = strconv.ParseUint(e.DoorsPresented, 2, 8); err != nil {
		return result, fmt.Errorf("Не удалось закодировать поле Doors Presented для EGTS_SR_PASSENGERS_COUNTERS: %v", err)
	}
	if err = buf.WriteByte(uint8(dpr)); err != nil {
		return result, fmt.Errorf("Не удалось записать поле Doors Presented для EGTS_SR_PASSENGERS_COUNTERS: %v", err)
	}

	if drl, err = strconv.ParseUint(e.DoorsReleased, 2, 8); err != nil {
		return result, fmt.Errorf("Не удалось закодировать поле Doors Released для EGTS_SR_PASSENGERS_COUNTERS: %v", err)
	}
	if err = buf.WriteByte(uint8(drl)); err != nil {
		return result, fmt.Errorf("Не удалось записать поле Doors Released для EGTS_SR_PASSENGERS_COUNTERS: %v", err)
	}

	binary.LittleEndian.PutUint16(maddrBuf, e.ModuleAddress)
	if _, err = buf.Write(maddrBuf); err != nil {
		return result, fmt.Errorf("Не удалось записать поле Module Address для EGTS_SR_PASSENGERS_COUNTERS: %v", err)
	}

	if e.RawDataFlag == "0" {
		for _, counter := range e.PassengersCountersData {
			encodedCounter, err := counter.encode()
			if err != nil {
				return result, fmt.Errorf("Не удалось закодировать поле Passengers Counters Data для EGTS_SR_PASSENGERS_COUNTERS: %v", err)
			}
			if _, err = buf.Write(encodedCounter); err != nil {
				return result, fmt.Errorf("Не удалось записать поле Passengers Counters Data для EGTS_SR_PASSENGERS_COUNTERS: %v", err)
			}
		}
	} else {
		if _, err = buf.Write(e.PassengersCountersRawData); err != nil {
			return result, fmt.Errorf("Не удалось записать поле Passengers Counters Data (raw) для EGTS_SR_PASSENGERS_COUNTERS: %v", err)
		}
	}

	return buf.Bytes(), err
}

//Length получает длинну закодированной подзаписи
func (e *SrPassengersCountersData) Length() uint16 {
	encoded, err := e.Encode()

	if err != nil {
		return 0
	}

	return uint16(len(encoded))
}

type PassengersCounter struct {
	DoorNo uint8 `json:"DoorNo"`
	In     uint8 `json:"In"`
	Out    uint8 `json:"Out"`
}

//Encode преобразовывает структуру в набор байт
func (c *PassengersCounter) encode() ([]byte, error) {
	if c.DoorNo == 0 {
		return []byte{}, fmt.Errorf("Попытка закодировать неинициализированную структуру PassengersCounter")
	}

	return []byte{c.In, c.Out}, nil
}
