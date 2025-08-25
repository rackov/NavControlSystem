package egts

import (
	"bytes"
	"encoding/binary"
)

func Read_EgtsPt(data *bytes.Buffer) (pt EgtsPt, err error) {
	pt = EgtsPt{}
	err = binary.Read(data, binary.LittleEndian, &pt)
	return
}
