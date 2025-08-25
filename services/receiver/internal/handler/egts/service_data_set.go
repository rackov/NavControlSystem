package egts

import (
	"bytes"
	"encoding/binary"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// RecordData структура секции подзапси у записи ServiceDataRecord
type RecordData struct {
	SubrecordType   byte       `json:"SRT"`
	SubrecordLength uint16     `json:"SRL"`
	SubrecordData   BinaryData `json:"SRD"`
}

// Decode разбирает байты в структуру подзаписи
func (rds *RecordDataSet) Decode(recDS []byte) error {
	var (
		err error
	)
	buf := bytes.NewBuffer(recDS)
	for buf.Len() > 0 {
		rd := RecordData{}
		if rd.SubrecordType, err = buf.ReadByte(); err != nil {
			return fmt.Errorf("Не удалось получить тип записи subrecord data: %v", err)
		}

		tmpIntBuf := make([]byte, 2)
		if _, err = buf.Read(tmpIntBuf); err != nil {
			return fmt.Errorf("Не удалось получить длину записи subrecord data: %v", err)
		}
		rd.SubrecordLength = binary.LittleEndian.Uint16(tmpIntBuf)

		subRecordBytes := buf.Next(int(rd.SubrecordLength))

		switch rd.SubrecordType {
		case EGTS_SR_POS_DATA:
			rd.SubrecordData = &SrPosData{}
		case EGTS_SR_TERM_IDENTITY:
			rd.SubrecordData = &SrTermIdentity{}
		case EGTS_SR_MODULE_DATA:
			rd.SubrecordData = &SrModuleData{}
		case EGTS_SR_RECORD_RESPONSE:
			rd.SubrecordData = &SrResponse{}
		case EGTS_SR_RESULT_CODE:
			rd.SubrecordData = &SrResultCode{}
		case EGTS_SR_EXT_POS_DATA:
			rd.SubrecordData = &SrExtPosData{}
		case EGTS_SR_AD_SENSORS_DATA:
			rd.SubrecordData = &SrAdSensorsData{}
			// case EGTS_SR_STATE_DATA:
			// 	// признак косвенный в спецификациях его нет
			// 	if rd.SubrecordLength == uint16(5) {
			// 		rd.SubrecordData = &SrStateData{}
			// 	} else {
			// 		// TODO: добавить секцию EGTS_SR_ACCEL_DATA
			// 		return fmt.Errorf("Не реализованная секция EGTS_SR_ACCEL_DATA: %d. Длина: %d. Содержимое: %X", rd.SubrecordType, rd.SubrecordLength, subRecordBytes)
			// 	}
		case EGTS_SR_STATE_DATA:
			rd.SubrecordData = &SrStateData{}
		case EGTS_SR_LIQUID_LEVEL_SENSOR:
			rd.SubrecordData = &SrLiquidLevelSensor{}
		case EGTS_SR_ABS_CNTR_DATA:
			rd.SubrecordData = &SrAbsCntrData{}
		case EGTS_SR_AUTH_INFO:
			rd.SubrecordData = &SrAuthInfo{}
		case EGTS_SR_COUNTERS_DATA:
			rd.SubrecordData = &SrCountersData{}
		// case EGTS_SR_EGTS_PLUS_DATA:
		// rd.SubrecordData = &StorageRecord{}
		case EGTS_SR_ABS_AN_SENS_DATA:
			rd.SubrecordData = &SrAbsAnSensData{}
		case EGTS_SR_ABS_DIG_SENS_DATA:
			rd.SubrecordData = &SrAbsDigSensData{}
		case EGTS_SR_DISPATCHER_IDENTITY:
			rd.SubrecordData = &SrDispatcherIdentity{}
		case EGTS_SR_PASSENGERS_COUNTERS:
			rd.SubrecordData = &SrPassengersCountersData{}
		default:
			log.Infof("Не известный тип подзаписи: %d. Длина: %d. Содержимое: %X", rd.SubrecordType, rd.SubrecordLength, subRecordBytes)
			continue
		}

		if err = rd.SubrecordData.Decode(subRecordBytes); err != nil {
			return err
		}

		*rds = append(*rds, rd)
	}

	return err
}

// Encode преобразовывает подзапись в набор байт
func (rds *RecordDataSet) Encode() ([]byte, error) {
	var (
		result []byte
		err    error
	)
	buf := new(bytes.Buffer)

	for _, rd := range *rds {
		if rd.SubrecordType == 0 {
			switch rd.SubrecordData.(type) {
			case *SrPosData:
				rd.SubrecordType = EGTS_SR_POS_DATA
			case *SrTermIdentity:
				rd.SubrecordType = EGTS_SR_TERM_IDENTITY
			case *SrResponse:
				rd.SubrecordType = EGTS_SR_RECORD_RESPONSE
			case *SrResultCode:
				rd.SubrecordType = EGTS_SR_RESULT_CODE
			case *SrExtPosData:
				rd.SubrecordType = EGTS_SR_EXT_POS_DATA
			case *SrAdSensorsData:
				rd.SubrecordType = EGTS_SR_AD_SENSORS_DATA
			case *SrStateData:
				rd.SubrecordType = EGTS_SR_STATE_DATA
			case *SrLiquidLevelSensor:
				rd.SubrecordType = EGTS_SR_LIQUID_LEVEL_SENSOR
			case *SrAbsCntrData:
				rd.SubrecordType = EGTS_SR_ABS_AN_SENS_DATA
			case *SrAuthInfo:
				rd.SubrecordType = EGTS_SR_AUTH_INFO
			case *SrCountersData:
				rd.SubrecordType = EGTS_SR_COUNTERS_DATA
			// case *StorageRecord:
			// 	rd.SubrecordType = SrEgtsPlusDataType
			case *SrAbsAnSensData:
				rd.SubrecordType = EGTS_SR_ABS_AN_SENS_DATA
			default:
				return result, fmt.Errorf("не известен код для данного типа подзаписи")
			}
		}

		if err := binary.Write(buf, binary.LittleEndian, rd.SubrecordType); err != nil {
			return result, err
		}

		if rd.SubrecordLength == 0 {
			rd.SubrecordLength = rd.SubrecordData.Length()
		}
		if err := binary.Write(buf, binary.LittleEndian, rd.SubrecordLength); err != nil {
			return result, err
		}

		srd, err := rd.SubrecordData.Encode()
		if err != nil {
			return result, err
		}
		buf.Write(srd)
	}

	result = buf.Bytes()

	return result, err
}

// Length получает длину массива записей
func (rds *RecordDataSet) Length() uint16 {
	var result uint16

	if recBytes, err := rds.Encode(); err != nil {
		result = uint16(0)
	} else {
		result = uint16(len(recBytes))
	}

	return result
}
