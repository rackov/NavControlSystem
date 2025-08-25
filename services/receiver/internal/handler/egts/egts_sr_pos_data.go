package egts

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
)

const (
	Pos2int = float64(429.4967295)
)

//SrPosData структура подзаписи типа EGTS_SR_POS_DATA, которая используется абонентским
//терминалом при передаче основных данных определения местоположения
/*
	Ntm  uint32 //Navigation Time  время навигации (число секунд с 00:00:00 01.01.2010 UTC);
	Lat  uint32 //Latitude  широта по модулю, градусы/90·0xFFFFFFFF и взята целая часть
	Long uint32 // долгота по модулю, градусы/180·0xFFFFFFFF и взята целая часть
	Flg  byte   //Flags:
	// 7 bit 1 - поле ALT передается, 0 - не передается;
	// 6 bit LOHS - битовый флаг определяет полушарие долготы: 0 - восточная долгота, 1 - западная долгота;
	// 5 bit LAHS - битовый флаг определяет полушарие широты:  0 - северная широта,  1 - южная широта;
	// 4 bit MV - битовый флаг, признак движения 1-движение 0 - стоянка
	// 3 bit BB - битовый флаг, признак отправки данных из памяти ("черный ящик"): 1 из памяти, 0 актуальные
	// 2 bit CS - битовое поле, тип используемой системы: 0 - система координат WGS-84, 1 -(ПЗ-90.02) государственная геоцентрическая система координат
	// 1 bit FIX - битовое поле, тип определения координат:  0 - 2D fix, 1 - 3D fix;
	// 0 bit битовый флаг, признак "валидности" координатных данных 1 - валидны, 0 -невалидны
	Spd uint16  //Speed  скорость в км/ч с дискретностью 0,1 км/ч (используется 14 младших бит)
	Dir byte    //Direction  направление движения
	Odm [3]byte //Odometer  пройденное расстояние (пробег) в км, с дискретностью 0,1 км
	Din byte    //Digital Inputs  битовые флаги, определяют состояние основных дискретных входов 1...8 (если бит равен 1, то соответствующий вход активен, если 0, то неактивен).
	Src byte    //Source  определяет источник (событие), инициировавший посылку данной навигационной информации
*/
type SrPosData struct {
	NavigationTime uint32 `json:"NTM"`  // time.Time
	Latitude       uint32 `json:"LAT"`  //float64
	Longitude      uint32 `json:"LONG"` //float64
	// ALTE                string `json:"ALTE"`
	// LOHS                string `json:"LOHS"`
	// LAHS                string `json:"LAHS"`
	FlagPos byte
	// MV                  string `json:"MV"`
	// BB                  string `json:"BB"`
	// CS                  string `json:"CS"`
	// FIX                 string `json:"FIX"`
	// VLD                 string `json:"VLD"`
	DirectionHighestBit uint8  `json:"DIRH"`
	AltitudeSign        uint8  `json:"ALTS"`
	Speed               uint16 `json:"SPD"`
	Direction           byte   `json:"DIR"`
	Odometer            uint32 `json:"ODM"`
	DigitalInputs       byte   `json:"DIN"`
	Source              byte   `json:"SRC"`
	Altitude            uint32 `json:"ALT"`
	SourceData          int16  `json:"SRCD"`
}

// Decode разбирает байты в структуру подзаписи
func (e *SrPosData) Decode(content []byte) error {
	var (
		err   error
		flags byte
		speed uint64
	)
	buf := bytes.NewReader(content)

	// Преобразуем время навигации к формату, который требует стандарт: количество секунд с 00:00:00 01.01.2010 UTC
	// startDate := time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)
	tmpUint32Buf := make([]byte, 4)
	if _, err = buf.Read(tmpUint32Buf); err != nil {
		return fmt.Errorf("не удалось получить время навигации: %v", err)
	}
	preFieldVal := binary.LittleEndian.Uint32(tmpUint32Buf)
	e.NavigationTime = preFieldVal + 1262304000 // startDate.Add(time.Duration(preFieldVal) * time.Second)

	// В протоколе значение хранится в виде: широта по модулю, градусы/90*0xFFFFFFFF  и взята целая часть
	if _, err = buf.Read(tmpUint32Buf); err != nil {
		return fmt.Errorf("не удалось получить широту: %v", err)
	}

	preFieldVal = binary.LittleEndian.Uint32(tmpUint32Buf)
	// e.Latitude = float64(float64(preFieldVal) * 90 / 0xFFFFFFFF)
	e.Latitude = uint32(math.Round(float64((preFieldVal)) * 90.0 / Pos2int))

	// В протоколе значение хранится в виде: долгота по модулю, градусы/180*0xFFFFFFFF  и взята целая часть
	if _, err = buf.Read(tmpUint32Buf); err != nil {
		return fmt.Errorf("не удалось получить время долгату: %v", err)
	}
	preFieldVal = binary.LittleEndian.Uint32(tmpUint32Buf)
	// e.Longitude = float64(float64(preFieldVal) * 180 / 0xFFFFFFFF)
	e.Longitude = uint32(math.Round(float64(preFieldVal) * 180 / Pos2int))

	//байт флагов
	if flags, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось получить байт флагов pos_data: %v", err)
	}
	e.FlagPos = flags
	// flagBits := fmt.Sprintf("%08b", flags)
	// e.ALTE = flagBits[:1]
	// e.LOHS = flagBits[1:2]
	// e.LAHS = flagBits[2:3]
	// e.MV = flagBits[3:4]
	// e.BB = flagBits[4:5]
	// e.CS = flagBits[5:6]
	// e.FIX = flagBits[6:7]
	// e.VLD = flagBits[7:]

	// скорость
	tmpUint16Buf := make([]byte, 2)
	if _, err = buf.Read(tmpUint16Buf); err != nil {
		return fmt.Errorf("не удалось получить скорость: %v", err)
	}
	spd := binary.LittleEndian.Uint16(tmpUint16Buf)
	e.DirectionHighestBit = uint8(spd >> 15 & 0x1)
	e.AltitudeSign = uint8(spd >> 14 & 0x1)

	speedBits := fmt.Sprintf("%016b", spd)
	if speed, err = strconv.ParseUint(speedBits[2:], 2, 16); err != nil {
		return fmt.Errorf("не удалось расшифровать скорость из битов: %v", err)
	}

	// т.к. скорость с дискретностью 0,1 км
	e.Speed = uint16(speed) / 10

	if e.Direction, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось получить направление движения: %v", err)
	}
	e.Direction |= e.DirectionHighestBit << 7

	bytesTmpBuf := make([]byte, 3)
	if _, err = buf.Read(bytesTmpBuf); err != nil {
		return fmt.Errorf("не удалось получить пройденное расстояние (пробег) в км: %v", err)
	}
	bytesTmpBuf = append(bytesTmpBuf, 0x00)
	e.Odometer = binary.LittleEndian.Uint32(bytesTmpBuf)

	if e.DigitalInputs, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось получить битовые флаги, определяют состояние основных дискретных входов: %v", err)
	}

	if e.Source, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("не удалось получить источник (событие), инициировавший посылку: %v", err)
	}

	if (e.FlagPos & 128) == 128 {
		bytesTmpBuf = []byte{0, 0, 0, 0}
		if _, err = buf.Read(bytesTmpBuf); err != nil {
			return fmt.Errorf("не удалось получить высоту над уровнем моря: %v", err)
		}
		e.Altitude = binary.LittleEndian.Uint32(bytesTmpBuf)
	}

	//TODO: разобраться с разбором SourceData
	return err
}

// Encode преобразовывает подзапись в набор байт
func (e *SrPosData) Encode() ([]byte, error) {
	var (
		err    error
		flags  uint64
		result []byte
	)

	buf := new(bytes.Buffer)
	// Преобразуем время навигации к формату, который требует стандарт: количество секунд с 00:00:00 01.01.2010 UTC
	// startDate := time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)
	// if err = binary.Write(buf, binary.LittleEndian, uint32(e.NavigationTime.Sub(startDate).Seconds())); err != nil {
	if err = binary.Write(buf, binary.LittleEndian, uint32(e.NavigationTime-1262304000)); err != nil {
		return result, fmt.Errorf("не удалось записать время навигации: %v", err)
	}

	// В протоколе значение хранится в виде: широта по модулю, градусы/90*0xFFFFFFFF  и взята целая часть
	// if err = binary.Write(buf, binary.LittleEndian, uint32(e.Latitude/90*0xFFFFFFFF)); err != nil {
	if err = binary.Write(buf, binary.LittleEndian, uint32(math.Round(float64(e.Latitude)/90*Pos2int))); err != nil {
		return result, fmt.Errorf("не удалось записать широту: %v", err)
	}

	// В протоколе значение хранится в виде: долгота по модулю, градусы/180*0xFFFFFFFF  и взята целая часть
	// if err = binary.Write(buf, binary.LittleEndian, uint32(e.Longitude/180*0xFFFFFFFF)); err != nil {
	if err = binary.Write(buf, binary.LittleEndian, uint32(math.Round(float64(e.Longitude)/180*Pos2int))); err != nil {
		return result, fmt.Errorf("не удалось записать долготу: %v", err)
	}

	//байт флагов
	flags = uint64(e.FlagPos)

	if err = buf.WriteByte(uint8(flags)); err != nil {
		return result, fmt.Errorf("не удалось записать флаги: %v", err)
	}

	// скорость
	speed := e.Speed*10 | uint16(e.DirectionHighestBit)<<15 // 15 бит
	speed = speed | uint16(e.AltitudeSign)<<14              //14 бит
	spd := make([]byte, 2)
	binary.LittleEndian.PutUint16(spd, speed)
	if _, err = buf.Write(spd); err != nil {
		return result, fmt.Errorf("не удалось записать скорость: %v", err)
	}

	dir := e.Direction &^ (e.DirectionHighestBit << 7)
	if err = binary.Write(buf, binary.LittleEndian, dir); err != nil {
		return result, fmt.Errorf("не удалось записать направление движения: %v", err)
	}

	bytesTmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytesTmpBuf, e.Odometer)
	if _, err = buf.Write(bytesTmpBuf[:3]); err != nil {
		return result, fmt.Errorf("не удалось запсиать пройденное расстояние (пробег) в км: %v", err)
	}

	if err = binary.Write(buf, binary.LittleEndian, e.DigitalInputs); err != nil {
		return result, fmt.Errorf("не удалось записать битовые флаги, определяют состояние основных дискретных входов: %v", err)
	}

	if err = binary.Write(buf, binary.LittleEndian, e.Source); err != nil {
		return result, fmt.Errorf("не удалось записать источник (событие), инициировавший посылку: %v", err)
	}

	if (e.FlagPos & 128) == 128 {
		bytesTmpBuf = []byte{0, 0, 0, 0}
		binary.LittleEndian.PutUint32(bytesTmpBuf, e.Altitude)
		if _, err = buf.Write(bytesTmpBuf[:3]); err != nil {
			return result, fmt.Errorf("не удалось записать высоту над уровнем моря: %v", err)
		}
	}

	//TODO: разобраться с записью SourceData
	result = buf.Bytes()
	return result, nil
}

// Length получает длинну закодированной подзаписи
func (e *SrPosData) Length() uint16 {
	var result uint16

	if recBytes, err := e.Encode(); err != nil {
		result = uint16(0)
	} else {
		result = uint16(len(recBytes))
	}

	return result
}
