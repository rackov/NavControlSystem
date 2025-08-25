package arnavi

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

type AllTags struct {
	TagNum byte
	Val    []byte
}

type TagsData struct {
	// битовая маска наличия основных полей:
	// lat=0001, lon=0011, Speed&sat&all = 00111
	ListActive uint64    `json:"list_active"`
	Latitude   int       `json:"lat"`
	Longitude  int       `json:"lon"`
	Speed      float32   `json:"speed"`
	Satellites int       `json:"satellites"`
	Altitude   int       `json:"altitude"`
	Course     int       `json:"course"`
	LL         [8]int    `json:"ll"`
	Data       []AllTags `json:"else_data"`
}

func (r *TagsData) Decode(rec []byte) error {
	var (
		err   error
		count int
	)
	temp := AllTags{}
	count = len(rec) / 5
	buf := bytes.NewBuffer(rec)
	r.ListActive = 0
	for i := 0; i < count; i++ {
		if temp.TagNum, err = buf.ReadByte(); err != nil {
			return fmt.Errorf("не удалось прочитать tag_num %v", err)
		}

		switch temp.TagNum {
		case TagsLat:
			var tLat float32
			binary.Read(buf, binary.LittleEndian, &tLat)
			r.Latitude = int(math.Round(float64(tLat)*1000000)) * 10
			r.ListActive = r.ListActive | 1
		case TagsLon:
			var tLon float32
			binary.Read(buf, binary.LittleEndian, &tLon)
			r.Longitude = int(math.Round(float64(tLon)*1000000)) * 10
			r.ListActive = r.ListActive | 2
		case TagsSpeedCourse:
			var tbyte byte
			if tbyte, err = buf.ReadByte(); err != nil {
				return fmt.Errorf("не удалось прочитать course %v", err)
			}
			r.Course = int(tbyte) * 2
			if tbyte, err = buf.ReadByte(); err != nil {
				return fmt.Errorf("не удалось прочитать Altitude %v", err)
			}
			r.Altitude = int(tbyte) * 10

			if tbyte, err = buf.ReadByte(); err != nil {
				return fmt.Errorf("не удалось прочитать Satellites %v", err)
			}
			r.Satellites = int(tbyte)
			if tbyte, err = buf.ReadByte(); err != nil {
				return fmt.Errorf("не удалось прочитать Speed %v", err)
			}
			r.Speed = 1.852 * float32(tbyte)
			r.ListActive = r.ListActive | 4
		// case TagsLL0:
		case TagsLL1:
			var tbyte uint16
			binary.Read(buf, binary.LittleEndian, &tbyte)
			r.LL[0] = int(tbyte)
		case TagsLL2:
			var tbyte uint16
			binary.Read(buf, binary.LittleEndian, &tbyte)
			r.LL[1] = int(tbyte)
		case TagsLL3:
			var tbyte uint16
			binary.Read(buf, binary.LittleEndian, &tbyte)
			r.LL[2] = int(tbyte)
		case TagsLL4:
			var tbyte uint16
			binary.Read(buf, binary.LittleEndian, &tbyte)
			r.LL[3] = int(tbyte)
		case TagsLL5:
			var tbyte uint16
			binary.Read(buf, binary.LittleEndian, &tbyte)
			r.LL[4] = int(tbyte)
		case TagsLL6:
			var tbyte uint16
			binary.Read(buf, binary.LittleEndian, &tbyte)
			r.LL[5] = int(tbyte)
		case TagsLL7:
			var tbyte uint16
			binary.Read(buf, binary.LittleEndian, &tbyte)
			r.LL[6] = int(tbyte)
		case TagsLL8:
			var tbyte uint16
			binary.Read(buf, binary.LittleEndian, &tbyte)
			r.LL[7] = int(tbyte)
		default:
			bytesTmpBuf := make([]byte, 4)
			if _, err = buf.Read(bytesTmpBuf); err != nil {
				return fmt.Errorf("не удалось получить значение tag: %v", err)
			}
			temp.Val = bytesTmpBuf
			r.Data = append(r.Data, temp)

		}
	}
	return err
}

func (r *TagsData) Encode() ([]byte, error) {
	var (
		result []byte
		err    error
	)
	buf := new(bytes.Buffer)

	// count=count+sumOfBits(int(r.ListActive))
	if r.ListActive&1 == 1 {
		if err = buf.WriteByte(TagsLat); err != nil {
			return result, fmt.Errorf("не удалось записать TagsLat")
		}
		tLat := float32(r.Latitude/10) / 1000000
		binary.Write(buf, binary.LittleEndian, &tLat)
	}
	if r.ListActive&2 == 2 {
		if err = buf.WriteByte(TagsLon); err != nil {
			return result, fmt.Errorf("не удалось записать TagsLon")
		}
		tLon := float32(r.Longitude/10) / 1000000
		binary.Write(buf, binary.LittleEndian, &tLon)
	}

	if r.ListActive&4 == 4 {
		if err = buf.WriteByte(TagsSpeedCourse); err != nil {
			return result, fmt.Errorf("не удалось записать TagsSpeedCourse")
		}
		buf.WriteByte(byte(r.Course / 2))
		buf.WriteByte(byte(r.Altitude / 10))
		buf.WriteByte(byte(r.Satellites))
		buf.WriteByte(byte(r.Speed / 1.852))

	}

	for _, d := range r.Data {

		buf.WriteByte(d.TagNum)
		buf.Write(d.Val)
	}

	result = buf.Bytes()
	return result, err
}

func (r *TagsData) Length() uint16 {
	var result uint16

	if recBytes, err := r.Encode(); err != nil {
		result = uint16(0)
	} else {
		result = uint16(len(recBytes))
	}
	return result
}

// func sumOfBits(n int) int {
// 	sum := 0

// 	for n > 0 {
// 		sum += n & 1 // Добавляем последний бит (0 или 1)
// 		n >>= 1      // Сдвигаем число вправо на 1 бит
// 	}

// 	return sum
// }
