package avif

func Test(data []byte) bool {
	if len(data) < 3 {
		return false
	}

	// AVIF Magic Numbers
	//https://www.garykessler.net/library/file_sigs.html
	return data[0] == 0x00 &&
		data[1] == 0x00 &&
		data[2] == 0x00
}
