package egts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	srAbsCntrDataBytes    = []byte{0x06, 0x75, 0x1D, 0x70}
	testEgtsSrAbsCntrData = SrAbsCntrData{
		CounterNumber: 6,
		CounterValue:  7347573,
	}
)

func TestEgtsSrAbsCntrData_Encode(t *testing.T) {
	posDataBytes, err := testEgtsSrAbsCntrData.Encode()
	if assert.NoError(t, err) {
		assert.Equal(t, posDataBytes, srAbsCntrDataBytes)
	}
}

func TestEgtsSrAbsCntrData_Decode(t *testing.T) {
	adSensData := SrAbsCntrData{}

	if err := adSensData.Decode(srAbsCntrDataBytes); assert.NoError(t, err) {
		assert.Equal(t, adSensData, testEgtsSrAbsCntrData)
	}

}

// проверяем что рекордсет работает правильно с данным типом подзаписи
func TestEgtsSrAbsCntrDataRs(t *testing.T) {
	egtsSrAbsCntrDataRDBytes := append([]byte{0x19, 0x04, 0x00}, srAbsCntrDataBytes...)
	egtsSrAbsCntrDataRD := RecordDataSet{
		RecordData{
			SubrecordType:   EGTS_SR_ABS_CNTR_DATA,
			SubrecordLength: testEgtsSrAbsCntrData.Length(),
			SubrecordData:   &testEgtsSrAbsCntrData,
		},
	}
	testStruct := RecordDataSet{}

	testBytes, err := egtsSrAbsCntrDataRD.Encode()
	if assert.NoError(t, err) {
		assert.Equal(t, testBytes, egtsSrAbsCntrDataRDBytes)

		if err = testStruct.Decode(egtsSrAbsCntrDataRDBytes); assert.NoError(t, err) {
			assert.Equal(t, egtsSrAbsCntrDataRD, testStruct)
		}
	}
}
