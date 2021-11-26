package avif

func Test(data []byte) bool {
	if len(data) < 12 {
		return false
	}

	// AVIF Magic Numbers
	//https://www.garykessler.net/library/file_sigs.html
	return data[0] == 0x00 &&
		data[1] == 0x00 &&
		data[2] == 0x00 &&
		data[3] == 0x28 &&
		data[4] == 0x66 &&
		data[5] == 0x74 &&
		data[6] == 0x79 &&
		data[7] == 0x70 &&
		data[8] == 0x61 &&
		data[9] == 0x76 &&
		data[10] == 0x69 &&
		data[11] == 0x73
}
