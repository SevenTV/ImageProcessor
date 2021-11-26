package mov

func Test(data []byte) bool {
	return (data[4] == 'f' &&
		data[5] == 't' &&
		data[6] == 'y' &&
		data[7] == 'p' &&
		data[8] == 'q' &&
		data[9] == 't') || (data[4] == 'm' &&
		data[5] == 'o' &&
		data[6] == 'o' &&
		data[7] == 'v')
}
