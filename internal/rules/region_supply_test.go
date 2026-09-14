package rules

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestRefillRegionSupplyMatchesOracle(t *testing.T) {
	pre, err := os.ReadFile(filepath.Join("..", "..", "workplace", "re273_rpre.bin"))
	if err != nil {
		t.Skip("缺 re273 回充前收據")
	}
	post, err := os.ReadFile(filepath.Join("..", "..", "workplace", "re273_rpost.bin"))
	if err != nil {
		t.Skip("缺 re273 回充後收據")
	}
	if len(pre) != RegionCount*RegionRecord || len(post) != len(pre) {
		t.Fatal("re273 收據大小錯誤")
	}
	got := append([]byte(nil), pre...)
	for i := 0; i < RegionCount; i++ {
		if !RefillRegionSupply(got[i*RegionRecord : (i+1)*RegionRecord]) {
			t.Fatalf("星域 %d 回充失敗", i)
		}
	}
	for i := range got {
		if got[i] != post[i] {
			t.Fatalf("第一個差異 offset 0x%X：got %02X want %02X", i, got[i], post[i])
		}
	}
}

func TestRefillRegionSupplySkipsNeutralAndCutRoute(t *testing.T) {
	makeRecord := func(owner byte, flags byte) []byte {
		r := make([]byte, RegionRecord)
		r[regionOwner], r[regionFlags] = owner, flags
		binary.LittleEndian.PutUint32(r[regionEconomy:], 900)
		binary.LittleEndian.PutUint16(r[regionDefence:], 100)
		return r
	}
	for _, r := range [][]byte{makeRecord(RegionOwnerNone, 0), makeRecord(0, 1)} {
		before := append([]byte(nil), r...)
		RefillRegionSupply(r)
		if string(r) != string(before) {
			t.Fatal("無主或斷路星域不應回充")
		}
	}
}

func TestDistributeRegionSupplyRemainderGoesToLastLargestPlanet(t *testing.T) {
	p := &Planets{PerRegion: make([]int, RegionCount),
		Data: make([]byte, PlanetHeader+3*PlanetRecord)}
	p.PerRegion[0] = 3
	for i := 0; i < 3; i++ {
		p.Record(i)[planetOwner] = 0
		p.SetPlanetEconomy(i, 1)
	}
	r := make([]byte, RegionRecord)
	writeRegionSupply(r, RegionSupplyPool{Missiles: [2]int{100, 0}, Fighters: [2]int{200, 0}})
	if !DistributeRegionSupply(r, p, 0, 0) {
		t.Fatal("分配失敗")
	}
	wantM, wantF := []int{33, 33, 34}, []int{66, 66, 68}
	for i := 0; i < 3; i++ {
		s, _ := ReadPlanetStock(p.Record(i))
		if s.Missiles != wantM[i] || s.Fighters != wantF[i] {
			t.Errorf("行星 %d 存量=%d/%d，預期 %d/%d", i,
				s.Missiles, s.Fighters, wantM[i], wantF[i])
		}
	}
}
