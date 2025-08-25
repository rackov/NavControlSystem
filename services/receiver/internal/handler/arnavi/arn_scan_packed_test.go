package arnavi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dumpScan = []byte{
		0x5B,
		0x01,
		0x01,       // type
		0x3C, 0x00, // size   60
	}
	testScan = ScanPaked{
		StartSign: 0x5B,
		Id:        1,
		// TypeContent:  1,
		// LengthPacket: 60,
	}
)

func TestScanPaked_Decode(t *testing.T) {
	newScan := ScanPaked{}
	err := newScan.Decode(dumpScan)
	if assert.NoError(t, err) {
		assert.Equal(t, newScan, testScan)
	}
}
