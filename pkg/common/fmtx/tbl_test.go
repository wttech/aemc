package fmtx_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"testing"
)

func TestTblValue(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	a.Equal("<empty>", fmtx.TblValue(nil))
	a.Equal("<empty>", fmtx.TblValue(""))
	a.Equal("true", fmtx.TblValue(true))
	a.Equal("4", fmtx.TblValue(4))
	a.Equal("25", fmtx.TblValue(25.0))
	a.Equal("aaa = 123, hello = world", fmtx.TblValue(map[string]any{"hello": "world", "aaa": 123}))
	a.Equal("a, b, 123, 456, false", fmtx.TblValue([]any{"a", "b", 123, "456", false}))
}
