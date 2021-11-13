package gif

func Test(data []byte) bool {
	if len(data) < 6 {
		return false
	}

	// GIF Magic Numbers
	// https://www.garykessler.net/library/file_sigs.html
	return data[0] == 'G' &&
		data[1] == 'I' &&
		data[2] == 'F' &&
		data[3] == '8' &&
		(data[4] == '7' || data[4] == '9') &&
		data[5] == 'a' &&
		data[len(data)-2] == 0x00 &&
		data[len(data)-1] == ';'
}
