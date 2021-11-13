package webp

func Test(data []byte) bool {
	if len(data) < 12 {
		return false
	}

	// WEBP Magic Numbers
	// https://www.garykessler.net/library/file_sigs.html
	return data[0] == 'R' &&
		data[1] == 'I' &&
		data[2] == 'F' &&
		data[3] == 'F' &&
		data[8] == 'W' &&
		data[9] == 'E' &&
		data[10] == 'B' &&
		data[11] == 'P'
}
