package avi

func Test(data []byte) bool {
	if len(data) < 16 {
		return false
	}

	// AVI Magic Numbers
	// https://www.garykessler.net/library/file_sigs.html
	return data[0] == 'R' &&
		data[1] == 'I' &&
		data[2] == 'F' &&
		data[3] == 'F' &&
		data[8] == 'A' &&
		data[9] == 'V' &&
		data[10] == 'I' &&
		data[11] == ' ' &&
		data[12] == 'L' &&
		data[13] == 'I' &&
		data[14] == 'S' &&
		data[15] == 'T'
}

// func ConvertToPng()
