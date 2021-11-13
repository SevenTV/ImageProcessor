package jpeg

func Test(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// JPEG Magic Numbers
	// https://www.garykessler.net/library/file_sigs.html
	return data[0] == 0xFF &&
		data[1] == 0xD8 &&
		data[len(data)-2] == 0xFF &&
		data[len(data)-1] == 0xD9
}
