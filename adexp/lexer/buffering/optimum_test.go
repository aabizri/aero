package buffering

import "testing"

func TestOptimum(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	res := OptimumBufferSize()
	t.Log(res)
}
