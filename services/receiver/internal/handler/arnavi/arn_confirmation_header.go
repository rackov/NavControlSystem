package arnavi

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ConfirmationHeader struct {
	TimeAnswer uint32
}

func (c *ConfirmationHeader) Decode(data []byte) error {
	var err error
	buf := bytes.NewBuffer(data)
	tmp := make([]byte, 4)
	if _, err = buf.Read(tmp); err != nil {
		return fmt.Errorf("не удалось получить время")
	}
	c.TimeAnswer = binary.LittleEndian.Uint32(tmp)
	return err
}

func (c *ConfirmationHeader) Encode() ([]byte, error) {
	var (
		result []byte
		err    error
	)
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, c.TimeAnswer); err != nil {
		return result, err
	}
	result = buf.Bytes()
	return result, err
}

func (c *ConfirmationHeader) Length() uint16 {
	var result uint16

	if recBytes, err := c.Encode(); err != nil {
		result = uint16(0)
	} else {
		result = uint16(len(recBytes))
	}

	return result
}
