package rules

import (
	"fmt"
	"strings"
	"testing"
)

func TestUpdateSpottingObserversAndRadii(t *testing.T) {
	units := []BattleUnit{
		{Fleet: 10, Squadron: 0, Faction: 0, X: 0, Y: 0},
		{Fleet: 10, Squadron: 1, Faction: 0, X: 800, Y: 800},
		{Fleet: 10, Squadron: -1, Faction: 0, Kind: 7, X: 1000, Y: 0},
		{Fleet: 20, Squadron: 1, Faction: 1, X: 200, Y: 0},
		{Fleet: 20, Squadron: 2, Faction: 1, X: 1250, Y: 0},
		{Fleet: 20, Squadron: 3, Faction: 1, X: 2060, Y: 0},
		{Fleet: 20, Squadron: -1, Faction: 1, Kind: 7, X: 10, Y: 0},
		{Fleet: 20, Squadron: 4, Faction: 1, X: 10, Y: 0, Out: true},
	}
	planets := []BattlePlanet{{Owner: 0, X: 2000, Y: 0, Defence: 1500}}
	UpdateSpotting(units, planets)

	for _, i := range []int{3, 4, 5} {
		if units[i].Detail != 1 || !units[i].Seen {
			t.Errorf("分隊 %d：Detail=%d Seen=%v，要被旗艦／偵察隊／行星看到一次",
				i, units[i].Detail, units[i].Seen)
		}
	}
	for _, i := range []int{6, 7} {
		if units[i].Detail != 0 || units[i].Seen {
			t.Errorf("分隊 %d 不該被看到：Detail=%d Seen=%v", i, units[i].Detail, units[i].Seen)
		}
	}
	// 同一輪被多隻眼睛覆蓋也只升一次。
	UpdateSpotting(units, planets)
	if units[3].Detail != 1 {
		t.Fatalf("同一輪重跑後升成 %d，要仍是 1", units[3].Detail)
	}
}

func TestSpottingRiseAndDecay(t *testing.T) {
	units := []BattleUnit{
		{Fleet: 1, Squadron: 0, Faction: 0, X: 0, Y: 0},
		{Fleet: 2, Squadron: 1, Faction: 1, X: 100, Y: 0},
	}
	for want := 1; want <= 3; want++ {
		UpdateSpotting(units, nil)
		if units[1].Detail != want {
			t.Fatalf("第 %d 輪看到後層級 %d，要 %d", want, units[1].Detail, want)
		}
		DecaySpotting(units)
	}
	if units[1].DetailHold != 2 {
		t.Fatalf("升到 3 的停留輪數 %d，要 2", units[1].DetailHold)
	}
	DecaySpotting(units)
	if units[1].Detail != 3 || units[1].DetailHold != 1 {
		t.Fatalf("第一輪沒看到後是 Detail=%d Hold=%d，要 3／1",
			units[1].Detail, units[1].DetailHold)
	}
	DecaySpotting(units)
	if units[1].Detail != 2 || units[1].DetailHold != 0 {
		t.Fatalf("第二輪沒看到後是 Detail=%d Hold=%d，要 2／0",
			units[1].Detail, units[1].DetailHold)
	}
	DecaySpotting(units)
	DecaySpotting(units)
	if units[1].Detail != 0 {
		t.Fatalf("再兩輪沒看到後層級 %d，要 0", units[1].Detail)
	}
}

func TestSpottingSpecialReveals(t *testing.T) {
	units := []BattleUnit{
		{Fleet: 1, Squadron: 0, Faction: 0},
		{Fleet: 2, Squadron: 0, Faction: 1},
		{Fleet: 2, Squadron: 1, Faction: 1},
		{Fleet: 2, Squadron: -1, Faction: 1, Kind: 7},
	}
	RevealFleet(units, 1, 2)
	if units[0].Detail != 3 || units[0].DetailHold != 2 || units[0].Seen {
		t.Fatal("打有主行星的揭露應設 3／2，但不設 Seen")
	}
	if RevealEnemiesIfBlind(units, 0, 6) {
		t.Fatal("第六回合不該觸發 AI 防僵局揭露")
	}
	if !RevealEnemiesIfBlind(units, 0, 7) {
		t.Fatal("第七回合完全看不到敵人時應揭露")
	}
	for _, i := range []int{1, 2} {
		if units[i].Detail != 3 || units[i].DetailHold != 6 || !units[i].Seen {
			t.Fatalf("敵方分隊 %d 沒被設成 3／6 且 Seen", i)
		}
	}
	if units[3].Detail != 0 {
		t.Fatal("偵察隊自己不該被防僵局規則揭露")
	}
}

// `s3fltmove.bin`（SHA-256
// f365c46fa6087590b2da4db001109e457e6d604218a73f946829297aac3e9758）
// 的 55 筆在場分隊。每列是 fleet squad faction kind currentX currentY out
// preDetail wantDetail wantSeen；preDetail 由收據的 +0x16 bit0 與最終 +0x17
// 反推，UpdateSpotting 後逐筆核對收據。
const s3fltmoveSpotting = `
12 0 0 0 416 280 0 0 0 0
12 1 0 0 419 258 0 0 0 0
12 2 0 0 394 277 0 0 0 0
12 3 0 0 422 235 0 0 0 0
12 4 0 0 371 274 0 0 0 0
12 5 0 2 396 255 0 0 0 0
12 6 0 2 424 213 0 0 0 0
12 7 0 2 349 272 0 0 0 0
12 255 0 7 416 280 1 0 0 0
16 0 1 0 1303 864 0 0 1 1
16 1 1 0 1273 854 0 0 1 1
16 2 1 0 1333 874 0 0 1 1
16 3 1 2 1263 884 0 0 1 1
16 4 1 2 1323 904 0 0 1 1
16 5 1 3 1293 894 0 0 1 1
16 255 1 7 1303 864 1 0 0 0
17 0 1 0 1320 812 0 0 1 1
17 1 1 2 1334 841 0 0 1 1
17 2 1 4 1291 826 0 0 1 1
17 3 1 4 1349 798 0 0 1 1
17 255 1 7 1320 812 1 0 0 0
28 0 0 0 1330 941 0 1 2 1
28 1 1 0 1282 983 0 0 1 1
28 2 1 0 1306 962 0 0 1 1
28 3 1 0 1354 920 0 0 1 1
28 4 1 0 1378 899 0 0 1 1
28 5 1 2 1348 1010 0 0 1 1
28 6 1 2 1396 969 0 0 1 1
28 7 1 3 1327 986 0 0 1 1
28 8 1 3 1375 944 0 0 1 1
28 255 1 7 1330 941 1 0 0 0
29 0 1 0 1299 875 0 0 1 1
29 1 1 2 1259 895 0 0 1 1
29 2 1 2 1319 915 0 0 1 1
29 3 1 4 1269 865 0 0 1 1
29 4 1 4 1329 885 0 0 1 1
29 5 1 4 1289 905 0 0 1 1
29 6 1 6 1279 936 0 0 1 1
29 255 1 7 1299 875 1 0 0 0
30 0 1 0 1168 1160 0 0 0 0
30 1 1 2 1157 1204 0 0 0 0
30 2 1 2 1212 1171 0 0 0 0
30 3 1 4 1141 1176 0 0 0 0
30 4 1 4 1195 1144 0 0 0 0
30 5 1 4 1184 1187 0 0 0 0
30 6 1 6 1201 1215 0 0 0 0
30 255 1 7 1168 1160 1 0 0 0
31 0 1 1 1320 812 0 2 3 1
31 1 1 1 1293 830 0 2 3 1
31 2 1 1 1347 794 0 2 3 1
31 3 1 2 1311 856 0 1 2 1
31 4 1 2 1338 839 0 1 2 1
31 5 1 2 1364 821 0 1 2 1
31 6 1 6 1356 865 0 1 2 1
31 255 1 7 1320 812 1 0 0 0
`

func TestSpottingMatchesS3FleetMoveReceipt(t *testing.T) {
	r := strings.NewReader(s3fltmoveSpotting)
	var units []BattleUnit
	var wantDetail []int
	var wantSeen []bool
	for {
		var fleet, squad, faction, kind, x, y, out, pre, want, seen int
		_, err := fmt.Fscan(r, &fleet, &squad, &faction, &kind, &x, &y,
			&out, &pre, &want, &seen)
		if err != nil {
			break
		}
		units = append(units, BattleUnit{
			Fleet: fleet, Squadron: squad, Faction: faction, Kind: kind,
			X: x, Y: y, Out: out != 0, Detail: pre,
		})
		wantDetail = append(wantDetail, want)
		wantSeen = append(wantSeen, seen != 0)
	}
	if len(units) != 55 {
		t.Fatalf("fixture 讀到 %d 筆，要 55", len(units))
	}
	planets := []BattlePlanet{
		{Owner: 0, X: 1320, Y: 812, Defence: 2852},
		{Owner: 1, X: 776, Y: 1612, Defence: 3800},
		{Owner: 1, X: 1256, Y: 1652, Defence: 4700},
	}
	UpdateSpotting(units, planets)
	for i := range units {
		if units[i].Detail != wantDetail[i] || units[i].Seen != wantSeen[i] {
			t.Errorf("第 %d 筆（艦隊 %d 組 %d）：Detail/Seen=%d/%v，要 %d/%v",
				i, units[i].Fleet, units[i].Squadron, units[i].Detail, units[i].Seen,
				wantDetail[i], wantSeen[i])
		}
		if wantDetail[i] == 3 && units[i].DetailHold != 2 {
			t.Errorf("第 %d 筆升到 3 的 hold=%d，要 2", i, units[i].DetailHold)
		}
	}
}
