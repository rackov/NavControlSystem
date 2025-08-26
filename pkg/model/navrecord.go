package model

import (
	"encoding/json"
)

/*
EGTS_SR_EGTS_PLUS_DATA      = 15 //EGTS_SR_EGTS_PLUS_DATA
  EGTS_SR_POS_DATA            = 16 // ACN-> основных данных определения местоположения
  EGTS_SR_EXT_POS_DATA        = 17 // ACN-> дополнительных данных определения местоположения
  EGTS_SR_AD_SENSORS_DATA     = 18 // ACN-> аппаратно-программный комплекс информации о состоянии дополнительных дискретных и аналоговых входов
  EGTS_SR_COUNTERS_DATA       = 19 // ->ACN данных о значении счетных входов
  EGTS_SR_STATE_DATA          = 20 // ACN-> информации о состоянии АСН
  EGTS_SR_ACCEL_DATA          = 21 // ->ACN ???
  EGTS_SR_LOOPIN_DATA         = 22 //АСН->данных о состоянии шлейфовых входов
  EGTS_SR_ABS_DIG_SENS_DATA   = 23 //АСН->данных о состоянии одного дискретного входа
  EGTS_SR_ABS_AN_SENS_DATA    = 24 //АСН->данных о состоянии одного аналогового входа
  EGTS_SR_ABS_CNTR_DATA       = 25 //АСН->данных о состоянии одного счетного входа
  EGTS_SR_ABS_LOOPIN_DATA     = 26 //АСН->данных о состоянии одного шлейфового входа
  EGTS_SR_LIQUID_LEVEL_SENSOR = 27 //АСН->данных о показаниях ДУТ
  EGTS_SR_PASSENGERS_COUNTERS = 28 //АСН->данных о показаниях счетчиков пассажиропотока
*/

type NavRecord struct {
	Client              uint32       `json:"tid"`
	PacketID            uint32       `json:"pk_id"`
	NavigationTimestamp uint32       `json:"nav_time"`
	ReceivedTimestamp   uint32       `json:"rec_time"`
	Latitude            uint32       `json:"lat"`
	Longitude           uint32       `json:"lng"`
	Speed               uint16       `json:"speed"`
	FlagPos             byte         `json:"flag"`
	Pdop                uint16       `json:"pdop"`
	Hdop                uint16       `json:"hdop"`
	Vdop                uint16       `json:"vdop"`
	Nsat                uint8        `json:"nsat"`
	Ns                  uint16       `json:"ns"`
	DigInput            byte         `json:"din"`
	Odometer            uint32       `json:"odm"`
	Course              uint8        `json:"course"`
	Imei                string       `json:"imei"`
	Imsi                string       `json:"imsi"`
	AnSenAbs            []Sensor     `json:"in_abs_in"`  // одного аналогового входа
	DigSenAbs           []DiSensor   `json:"in_abs_dig"` // одного дискретного входа
	AnSensors           []DopAnIn    `json:"in_an"`      // дополнительных аналоговых входов
	DigSenonrs          []DopDigIn   `json:"in_dig"`     // дополнительных дискретного входа
	DigSenOuts          []int        `json:"out_dig"`    // дополнительных дискретного выхода
	LiquidSensors       LiquidSensor `json:"sn_liq"`     // данных о показаниях ДУТ
}

func (eep *NavRecord) ToBytes() ([]byte, error) {
	return json.Marshal(eep)
}

type DopAnIn struct {
	Asfe byte      `json:"asfe"`
	Ansi [8]uint32 `json:"ansi"`
}

type DopDigIn struct {
	Dioe byte    `json:"dioe"`
	Adio [8]byte `json:"adio"`
}

type DiSensor struct {
	StateNumber byte `json:"st_num"`
	Number      byte `json:"num"`
}

type LiquidSensor struct {
	FlagLiqNum uint8     `json:"fl_ln"`
	Value      [8]uint32 `json:"val_l"`
}

type Sensor struct {
	SensorNumber uint8  `json:"sn_n"`
	Value        uint32 `json:"val"`
}

type NavRecords struct {
	PacketType int16       `json:"packet_type"`
	PacketID   uint32      `json:"packet_id"`
	RecNav     []NavRecord `json:"record"`
}
