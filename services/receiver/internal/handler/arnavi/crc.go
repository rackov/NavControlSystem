package arnavi

func Crc_sum(content []byte) byte {
	var res byte
	res = 0x0
	for _, b := range content {
		res = res + b
	}
	return res
}
