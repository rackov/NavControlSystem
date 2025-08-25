package arnavi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	bytesTime = []byte{0x21, 0x5A, 0xAA, 0x5A}
	dataTime  = ConfirmationHeader{
		TimeAnswer: 1521113633,
	}
)

func TestConfirmationHeader_Decode(t *testing.T) {
	newTime := ConfirmationHeader{}
	err := newTime.Decode(bytesTime)
	if assert.NoError(t, err) {
		assert.Equal(t, newTime, dataTime)
	}
}
func TestConfirmationHeader_Encode(t *testing.T) {

	buteTest, err := dataTime.Encode()
	if assert.NoError(t, err) {
		assert.Equal(t, bytesTime, buteTest)
	}
}
