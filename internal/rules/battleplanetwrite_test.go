package rules

import "testing"

// 佔領要留得住：原版整場都直接改行星記錄（`docs/re/150` 量到 `+0x11`
// 從 0 變 1），remake 的戰場副本得寫回去。
func TestApplyBattlePlanetsRoundTrips(t *testing.T) {
	p := &Planets{PerRegion: []int{2}, Data: make([]byte, PlanetHeader+2*PlanetRecord)}
	r0 := p.rec(0)
	r0[planetOwner] = 0
	r0[planetDefence], r0[planetDefence+1] = 0x10, 0x27 // 10000
	r0[planetGunState] = 3
	r0[planetGunAim], r0[planetGunAim+1] = 0x40, 0x00
	WritePlanetStock(r0, PlanetStock{Missiles: 900, Fighters: 40,
		Ships: [7]int{7, 6, 5, 4, 3, 2, 1}})

	bp := p.BattlePlanetsFor(0, []PlanetSite{{X: 1, Y: 2}, {X: 3, Y: 4}})
	if len(bp) != 2 {
		t.Fatalf("配出 %d 顆，預期 2", len(bp))
	}
	if bp[0].Owner != 0 || bp[0].Defence != 10000 || bp[0].GunState != 3 || bp[0].GunAim != 0x40 {
		t.Fatalf("讀出來的第 0 顆 ＝ %+v", bp[0])
	}
	if bp[0].Stock.Missiles != 900 || bp[0].Stock.Ships[0] != 7 {
		t.Fatalf("讀出來的存量 ＝ %+v", bp[0].Stock)
	}

	// 打下來：capturePlanet 的四個效果。
	capturePlanet(&bp[0], 1)
	bp[0].Defence = 250
	bp[0].GunState, bp[0].GunAim = 0, 0xFFFF
	p.ApplyBattlePlanets(bp)

	if got := p.Owner(0); got != 1 {
		t.Errorf("擁有者 ＝ %d，預期 1", got)
	}
	if got := p.Defence(0); got != 250 {
		t.Errorf("防衛力 ＝ %d，預期 250", got)
	}
	back := p.BattlePlanetsFor(0, []PlanetSite{{}, {}})
	if back[0].GunState != 0 || back[0].GunAim != 0xFFFF {
		t.Errorf("要塞砲 ＝ %d／%04X，預期 0／FFFF", back[0].GunState, back[0].GunAim)
	}
	if back[0].Stock.Ships != ([7]int{}) || back[0].Stock.Missiles != 450 ||
		back[0].Stock.Fighters != 0 {
		t.Errorf("佔領後的存量 ＝ %+v，預期艘數全 0、飛彈 450、艦載機 0", back[0].Stock)
	}
	// 沒被碰到的那一顆不受影響。
	if p.Owner(1) != 0 || p.Defence(1) != 0 {
		t.Errorf("第 1 顆被動到了：擁有者 %d、防衛力 %d", p.Owner(1), p.Defence(1))
	}
}

// 星域擁有者跟著行星走：全屬同一方才算那一方的（`docs/re/17`）。
func TestRegionOwnerFollowsCapturedPlanets(t *testing.T) {
	p := &Planets{PerRegion: []int{2}, Data: make([]byte, PlanetHeader+2*PlanetRecord)}
	p.rec(0)[planetOwner], p.rec(1)[planetOwner] = 0, 1
	if got := p.RegionOwner(0); got != RegionOwnerNone {
		t.Fatalf("混雜時 ＝ %d，預期無主 %d", got, RegionOwnerNone)
	}
	bp := p.BattlePlanetsFor(0, []PlanetSite{{}, {}})
	capturePlanet(&bp[1], 0) // 把同盟那顆打下來
	p.ApplyBattlePlanets(bp)
	if got := p.RegionOwner(0); got != 0 {
		t.Errorf("兩顆都是帝國時 ＝ %d，預期 0", got)
	}
}
