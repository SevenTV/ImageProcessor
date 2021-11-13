package png

func Test(data []byte) bool {
	if len(data) < 8 {
		return false
	}

	// PNG Magic Numbers
	// https://www.garykessler.net/library/file_sigs.html
	return data[0] == 0x89 &&
		data[1] == 'P' &&
		data[2] == 'N' &&
		data[3] == 'G' &&
		data[4] == 0x0D &&
		data[5] == 0x0A &&
		data[6] == 0x1A &&
		data[7] == 0x0A &&
		data[len(data)-8] == 'I' &&
		data[len(data)-7] == 'E' &&
		data[len(data)-6] == 'N' &&
		data[len(data)-5] == 'D' &&
		data[len(data)-4] == 0xAE &&
		data[len(data)-3] == 'B' &&
		data[len(data)-2] == 0x60 &&
		data[len(data)-1] == 0x82
}
