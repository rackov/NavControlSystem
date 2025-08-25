package arnavi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/rackov/NavControlSystem/pkg/logger"
	"github.com/sirupsen/logrus"
)

type Server struct {
	listen net.Listener
	Addr   string
	// logger        chan []byte
	log  *logrus.Logger
	file *os.File
}
type ConServer struct {
	conn          net.Conn
	authorization bool
	log           *logrus.Logger
	// cursor        int
	id     int // номер принятого пакета, если пакет отправлен то =0
	ImeiId uint64
}

func (ser *Server) Stop() error {
	var err error
	if ser.listen != nil {
		err = ser.listen.Close()
	}
	// Добавляем закрытие файла
	if ser.file != nil {
		ser.file.Close() // Закрываем файл
	}
	return err
}
func (ser *Server) Run() {

	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}

	ser.log = logger.GetLogger()

	var err error

	ser.listen, err = net.Listen("tcp", ser.Addr)
	if err != nil {
		ser.log.Fatalf("Не удалось открыть соединение: %v", err)
	}
	defer ser.listen.Close()

	ser.log.Infof("Запущен сервер %s...", ser.Addr)
	for {
		sc := ConServer{}

		sc.conn, err = ser.listen.Accept()
		sc.authorization = false
		sc.id = 0

		sc.log = ser.log
		// sc.log.SetLevel(logrus.DebugLevel)

		if err != nil {
			ser.log.WithField("err", err).Errorf("Ошибка соединения")
		} else {
			go sc.handleConn()
		}

	}
}

func (sc *ConServer) handleConn() {
	local_info := sc.conn.RemoteAddr()
	readBuf := make([]byte, 1024)
	localBuffer := new(bytes.Buffer)
	var (
		err error
		n   int
	)
	// sc.cursor = 0
	for {
		n, err = sc.conn.Read(readBuf)
		if err != nil {
			break
		}
		if n > 0 {
			sc.log.Debugf("dump:\n%s", hex.Dump(readBuf[:n]))
			localBuffer.Write(readBuf[:n])

			err = sc.processExistingData(localBuffer)

			if err != nil {
				sc.log.Infof("ошибка %v", err)
				break
			}
		}
	}
	sc.log.Infof("Disconnected from %s\n error: %v", local_info, err)

	sc.conn.Close()

}

func (sc *ConServer) processExistingData(data *bytes.Buffer) error {
	var (
		err error
		// sign byte
	)

	if !sc.authorization {
		return sc.Authorization(data)
	}

	for {
		buf := data.Bytes()
		nsize := data.Len()
		if nsize == 0 {
			return nil
		}
		if sc.id == 0 {
			if nsize < SizeScan {
				return nil
			}
			scan := ScanPaked{}
			if err = scan.Decode(buf); err != nil {
				return fmt.Errorf("не удалось просканировать scan %v", err)
			}
			sc.id = int(scan.Id)
			data.Next(SizeScan)
			continue
		}

		if buf[0] == SigPackEnd {
			if err = sc.finishpacked(); err != nil {
				return fmt.Errorf(" %v", err)
			}
			data.Next(1)
			sc.log.Debugf("отправлен  пакет подтверждения № %d", sc.id)
			sc.id = 0
			continue
		}
		if nsize < 3 {
			return err
		}
		scp := ScanPacket{}
		if err = scp.Decode(buf); err != nil {
			return fmt.Errorf("не удалось декодировать packet %v", err)
		}
		if nsize < (int(scp.LengthPacket) + 8) {
			return err
		}
		// запись пакета
		err = sc.savePacket(data)
		if err != nil {
			return err
		}
		data.Next(int(scp.LengthPacket) + 8)
	}

}
func (sc *ConServer) finishpacked() error {
	var err error
	buf, err := AnswerPacked(sc.id)

	if err != nil {
		return fmt.Errorf("ошибка формирования подтверждения  %v", err)
	}
	if _, err := sc.conn.Write(buf); err != nil {
		return fmt.Errorf("ошибка формирования подтверждения  %v", err)
	}

	return err

}
func (sc *ConServer) savePacket(data *bytes.Buffer) error {
	var err error
	packets := PacketS{}
	buf := data.Bytes()
	if err = packets.Decode(buf); err != nil {
		return err
	}
	unixTime := time.Unix(int64(packets.TimePacket), 0)
	formattedTime := unixTime.Format("2006-01-02 15:04:05")
	switch pack := packets.Data.(type) {
	case *TagsData:
		sc.log.Infof(" %d получен пакет время: %s \n { time:%d, lat: %d, lon: %d }\n"+
			"course %d, speed %f, Satellites %X :, GPS %d, Glonass %d  LL: %v \n DATA: %v",
			sc.ImeiId,
			formattedTime, packets.TimePacket,
			pack.Latitude, pack.Longitude,
			pack.Course, pack.Speed, pack.Satellites, pack.Satellites&0xf, (pack.Satellites>>4)&0xf,
			pack.LL,
			pack.Data)

	default:
		sc.log.Infof("получен пакет неизвестный тип пакета № %X", packets.TypeContent)
	}

	return err
}

func (sc *ConServer) Authorization(data *bytes.Buffer) error {
	var (
		err error
	)
	nsize := data.Len()

	if nsize < SizeAuth {
		//ждем
		return err
	}
	headOne := HeadOne{}
	buf := data.Bytes()
	err = headOne.Decode(buf)
	if err != nil {
		return fmt.Errorf("ошибка разбора старт. пакета %v", err)
	}
	if headOne.Version == 0x24 {
		if nsize < (SizeAuth + 8) {
			return err
		}
		data.Next(SizeAuth + 8)
		sc.authorization = true
		sc.log.Debugf("Принят расширенный пакет авторизации Id|Imei: %d ", headOne.IdImei)
		// sc.log.Debugf("dump %s \n", hex.Dump(buf[:(SizeAuth+8)]))

		return err
	}
	sc.ImeiId = headOne.IdImei
	sc.log.Debugf("Принят пакет авторизации Id|Imei: %d ", headOne.IdImei)
	// sc.log.Debugf("dump %s", hex.Dump(buf[:SizeAuth]))
	_, err = sc.conn.Write(AnswerHeader())
	if err != nil {
		sc.log.Errorf("ошибка отправки %v", err)
	}
	data.Next(SizeAuth)
	sc.authorization = true
	return err

}
