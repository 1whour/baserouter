package baserouter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable(t *testing.T) {
	have := make(map[byte]int)
	missing := make(map[byte]int)

	for _, c := range offsetTable {
		have[c]++
		missing[c]++
	}

	for i := 0; i < 256; i++ {
		v := have[byte(i)]
		assert.Equal(t, v, 1, fmt.Sprintf("(%c), index:%d", i, i))

		delete(missing, byte(i))
		if v != 1 {
			return
		}
	}

	assert.Equal(t, len(missing), 0)
}
