package arnavi

type BinaryData interface {
	Decode([]byte) error
	Encode() ([]byte, error)
	Length() uint16
}
