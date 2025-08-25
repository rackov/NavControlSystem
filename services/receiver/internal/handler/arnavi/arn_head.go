package arnavi

// константы пакетов
const (
	AnswerStart  = 4
	SignedStart  = 0x7B
	SignedEnd    = 0x7D
	SegnedHeader = 0xFF
	SigPackStart = 0x5B
	SigPackEnd   = 0x5D
	SignAuth     = 0xFF
)

// размеры
const (
	SizeAuth     = 10
	SizeScan     = 2
	SizeUnixTime = 4
)

// константы packets
const (
	// подтверждение сервера на пакет авторизации
	ConfirmationHeaderType = 0x00
	// пакет данных
	PackTagsType = 0x01
	// текстовых данных
	PackTextType = 0x03
	//  пакет с файлом
	PackFileType = 0x04
	// пакет двоичных данных
	PackBinaryType = 0x06
	//  пакет с коммандой
	PackCom2Type = 0x08
	//  пакет с коммандой
	PackComType = 0x09
	//отладочная информация для передачи на WEB
	PackWebType = 0x13
)

// константы tag_nums
const (
	TagsLat         = 3
	TagsLon         = 4
	TagsSpeedCourse = 5
	TagsLL0         = 70
	TagsLL1         = 71
	TagsLL2         = 72
	TagsLL3         = 73
	TagsLL4         = 74
	TagsLL5         = 75
	TagsLL6         = 76
	TagsLL7         = 77
	TagsLL8         = 78
	TagsLL9         = 79
)
