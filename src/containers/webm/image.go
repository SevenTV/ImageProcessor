package webm

func Test(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// WEBM/MKV Magic Numbers
	// https://www.garykessler.net/library/file_sigs.html
	return data[0] == 0x1A &&
		data[1] == 0x45 &&
		data[2] == 0xDF &&
		data[3] == 0xA3
}
