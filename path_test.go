package baserouter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenPath(t *testing.T) {
	need := []byte("/test/:name/last")
	insertPath := []byte("/test/:/last")
	p := genPath(need)
	assert.Equal(t, p.originalPath, need)
	assert.Equal(t, p.insertPath, insertPath)
	assert.NotNil(t, p.paramPath[len("/test/")])
}
