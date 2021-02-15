package directory_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/tasdomas/uniquemachines/server/directory"
)

func TestUniqueMachinesCount(t *testing.T) {
	c := qt.New(t)

	d := directory.New()

	// New machine being registered.
	resp := d.UpdateMachine("machine1", "m1token1", "m1token2")
	c.Assert(resp, qt.Equals, "machine1")
	c.Assert(d.Count(), qt.Equals, 1)

	// Machine updating its token.
	resp = d.UpdateMachine("machine1", "m1token2", "m1token3")
	c.Assert(resp, qt.Equals, "machine1")
	c.Assert(d.Count(), qt.Equals, 1)

	// Clone.
	resp = d.UpdateMachine("machine1", "m1token1", "m1token4")
	c.Assert(resp, qt.Not(qt.Equals), "machine1")
	c.Assert(d.Count(), qt.Equals, 2)
}
