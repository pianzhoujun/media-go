package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNext(t *testing.T) {
	data := make([]byte, 0)
	data = append(data, 0xf1)
	data = append(data, 0xf1)

	bs := NewBitStream(data)
	assert.Equal(t, bs.Next(), 1)
	assert.Equal(t, bs.Next(), 1)
	assert.Equal(t, bs.Next(), 1)
	assert.Equal(t, bs.Next(), 1)
	assert.Equal(t, bs.Next(), 0)
	assert.Equal(t, bs.Next(), 0)
	assert.Equal(t, bs.Next(), 0)
	assert.Equal(t, bs.Next(), 1)

	assert.Equal(t, bs.Next(), 1)
	assert.Equal(t, bs.Next(), 1)
	assert.Equal(t, bs.Next(), 1)
	assert.Equal(t, bs.Next(), 1)
	assert.Equal(t, bs.Next(), 0)
	assert.Equal(t, bs.Next(), 0)
	assert.Equal(t, bs.Next(), 0)
	assert.Equal(t, bs.Next(), 1)

	assert.Equal(t, bs.Next(), -1)
}
