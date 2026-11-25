package sizex_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wttech/aemc/pkg/common/sizex"
)

func TestParse(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	size, err := sizex.Parse("")
	a.NoError(err)
	a.Equal(uint64(0), size)

	size, err = sizex.Parse("100GB")
	a.NoError(err)
	a.Equal(uint64(100_000_000_000), size)

	size, err = sizex.Parse("1 GiB")
	a.NoError(err)
	a.Equal(uint64(1<<30), size)

	_, err = sizex.Parse("lots")
	a.Error(err)
}

func TestParseThreshold(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	size, err := sizex.ParseThreshold("10%", 1000)
	a.NoError(err)
	a.Equal(uint64(100), size)

	size, err = sizex.ParseThreshold("50GB", 1000)
	a.NoError(err)
	a.Equal(uint64(50_000_000_000), size)

	size, err = sizex.ParseThreshold("", 1000)
	a.NoError(err)
	a.Equal(uint64(0), size)

	_, err = sizex.ParseThreshold("150%", 1000)
	a.Error(err)

	_, err = sizex.ParseThreshold("x%", 1000)
	a.Error(err)
}
