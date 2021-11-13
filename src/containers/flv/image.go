package flv

func Test(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// FLV Magic Numbers
	// https://www.garykessler.net/library/file_sigs.html
	return data[0] == 'F' &&
		data[1] == 'L' &&
		data[2] == 'V' &&
		data[3] == 0x01
}
