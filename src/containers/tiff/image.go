package tiff

func Test(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// TIFF Magic Numbers
	// https://www.garykessler.net/library/file_sigs.html
	return data[0] == 'I' &&
		((data[1] == ' ' && data[2] == 'I') || (data[1] == 'I' && data[2] == '*' && data[3] == 0x00))
}
