package game

import (
	"os"
	"testing"
)

func TestTmpDump6(t *testing.T) {
	g := reachCommandPhase(t)
	g.Click(352, 125)
	g.Click(400, 125)
	g.Click(400, 85)
	os.WriteFile("/src/workplace/tmp/adm.bin", g.Frame().Pix, 0o644)
}
