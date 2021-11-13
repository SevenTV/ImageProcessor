package mp4

func Test(data []byte) bool {
	if len(data) < 12 {
		return false
	}

	// MP4 Magic Numbers
	// https://www.garykessler.net/library/file_sigs.html
	return data[4] == 'f' &&
		data[5] == 't' &&
		data[6] == 'y' &&
		data[7] == 'p' &&
		((data[8] == 'M' && data[9] == 'S' && data[10] == 'N' && data[11] == 'V') || (data[8] == 'i' && data[9] == 's' && data[10] == 'o' && data[11] == 'm') || (data[8] == 'm' && data[9] == 'p' && data[10] == '4' && data[11] == '2'))
}
