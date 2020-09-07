package baserouter

var offsetTable [256]byte
var offsetToChar [256]byte

func init() {
	initOffsetTable()
	initCharTable()
}

func initOffsetTable() {

	var offset byte
	// init offsetTable
	for i := 0; i < 256; i++ {
		offsetTable[i] = byte(i)
	}

	offset = 1
	for i := byte('a'); i <= 'z'; i++ {
		// swap 'a' - 'z'
		newPos := i - 'a' + offset
		tmp := offsetTable[i]
		offsetTable[i] = newPos
		offsetTable[newPos] = tmp
	}
	offset += 26 /* + a-z */

	for i := byte('A'); i <= 'Z'; i++ {
		// swap 'A' - 'Z'
		newPos := i - 'A' + offset
		tmp := offsetTable[i]
		offsetTable[i] = newPos
		offsetTable[newPos] = tmp
	}
	offset += 26 /* + A-Z */

	for i := byte('0'); i <= '9'; i++ {
		// swap '0' - '9'
		newPos := i - '0' + offset
		tmp := offsetTable[i]
		offsetTable[i] = newPos
		offsetTable[newPos] = tmp
	}
	offset += 10 /* 0-9 */

	tmp := offsetTable['/']
	offsetTable['/'] = offset
	offsetTable[offset] = tmp

}

func initCharTable() {
	for char, offset := range offsetTable {
		offsetToChar[offset] = byte(char)
	}
}

func getCharFromOffset(offset int) (char byte) {
	return offsetToChar[offset]
}

func getCodeOffset(b byte) int {
	return int(offsetTable[b])
}
