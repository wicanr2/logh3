package rules

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 61案直接原版收據：36案快取全byte、十案情報全78 bytes；可見性留給完整Game接線。
func TestSpottingDisplayAndByteRulesMatchOriginal(t *testing.T) {
	base := filepath.Join("..", "..")
	p := filepath.Join(base, "workplace/explore/battle79-spot-r41-final-a.json")
	raw, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		t.Skip("缺原版顯示收據；重生入口docs/playtest/046-battle-spotting-phase.md")
	}
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != "f03cc58211076e8bc0aa5a2a2e09414ae84716d68080fa044af62169ada01471" {
		t.Fatal("原版顯示收據身份不同")
	}
	dataRaw, err := os.ReadFile(filepath.Join(base, "docs/internals/data/battle-spotting-display.json"))
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		Receipt struct {
			Rows []struct {
				Name          string
				Before, After map[string]string
				Returns       []uint32
			}
		}
	}
	if err := json.Unmarshal(dataRaw, &data); err != nil || len(data.Receipt.Rows) != 61 {
		t.Fatal("原版案例格式或數量不同")
	}
	decode := func(text string, size int) []byte {
		t.Helper()
		b, err := hex.DecodeString(text)
		if err != nil || len(b) != size {
			t.Fatal("完整紀錄格式不同")
		}
		return b
	}
	unit := func(b []byte) BattleUnit {
		w := func(n int) int { return int(binary.LittleEndian.Uint16(b[n:])) }
		return BattleUnit{Faction: int(b[0]), Fleet: int(b[1]), Squadron: int(b[3]), Kind: int(b[4]),
			X: w(52), Y: w(54), Out: b[30]&128 != 0, Seen: b[22]&1 != 0,
			Detail: int(b[23]), DetailHold: int(b[24])}
	}
	count := 0
	for _, row := range data.Receipt.Rows {
		if !strings.HasPrefix(row.Name, "cache-") && !strings.HasPrefix(row.Name, "intelligence-") {
			continue
		}
		count++
		t.Run(row.Name, func(t *testing.T) {
			if strings.HasPrefix(row.Name, "cache-") {
				got := decode(row.Before["unit"], 78)
				u := unit(got)
				u.FleetNumber = int(decode(row.Before["fleet"], 122)[2])
				force := strings.HasPrefix(row.Name, "cache-f1-")
				got[47] = byte(BattleCachedSprite(u, force, int(got[47])))
				if !bytes.Equal(got, decode(row.After["unit"], 78)) {
					t.Errorf("完整SHIP78的圖示重算與原版不同")
				}
				return
			}
			records := [][]byte{decode(row.Before["observer"], 78), decode(row.Before["candidate"], 78)}
			units := []BattleUnit{unit(records[0]), unit(records[1])}
			// 原版只呼叫observer的01FE，不額外執行候選Kind7對另一席的觀測。
			observer := units[0]
			found := ObserveSpotting(units, observer.Faction, observer.X, observer.Y, FlagshipVision)
			if len(row.Returns) != 1 {
				t.Fatal("原版診斷呼叫數不同")
			}
			wantFound := int(row.Returns[0] & 0xffff)
			if len(found) != wantFound {
				t.Errorf("原版單觀測者提升數%d，要%d", len(found), wantFound)
			}
			for i, label := range []string{"observer", "candidate"} {
				got, u := records[i], units[i]
				got[22] &^= 1
				if u.Seen {
					got[22] |= 1
				}
				got[23], got[24] = byte(u.Detail), byte(u.DetailHold)
				if !bytes.Equal(got, decode(row.After[label], 78)) {
					t.Errorf("%s完整SHIP78情報與原版不同", label)
				}
			}
		})
	}
	if count != 46 {
		t.Fatalf("規則回歸%d案，要46", count)
	}
}

// 原版輸入刷新處理已回收偵察隊；Detail1重新出現標記，全78 bytes保持可回查。
func TestRecoveredScoutMarkerRefreshMatchesOriginal(t *testing.T) {
	base := filepath.Join("..", "..")
	raw, err := os.ReadFile(filepath.Join(base, "workplace/explore/battle79-spot-r41-recovered-a.json"))
	if os.IsNotExist(err) {
		t.Skip("缺原版回收刷新收據；入口docs/playtest/046-battle-spotting-phase.md")
	}
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != "d9cb659cc3bffced7dd514d0ba1d3cba507f3d75f4ed4c549657763fa06152f9" {
		t.Fatal("原版回收收據身份不同")
	}
	var data struct {
		Rows []struct {
			Name          string
			Before, After map[string]string
		}
	}
	if err := json.Unmarshal(raw, &data); err != nil || len(data.Rows) != 2 {
		t.Fatal("原版回收案例不同")
	}
	for _, row := range data.Rows {
		t.Run(row.Name, func(t *testing.T) {
			before, err := hex.DecodeString(row.Before["scout"])
			if err != nil || len(before) != ShipRecord {
				t.Fatal("原版完整SHIP格式不同")
			}
			w := func(n int) int { return int(binary.LittleEndian.Uint16(before[n:])) }
			u := BattleUnit{Faction: int(before[0]), Kind: 7, Squadron: -1, Out: true,
				X: w(52), Y: w(54), Detail: int(before[23]), MarkerFlags: int(before[25]),
				MarkerX: w(26), MarkerY: w(28), Marked: before[25]&128 != 0}
			UpdateBattleMarker(&u, false, true)
			s := Slot{Ship: bytes.Clone(before)}
			s.ApplyBattleMarker(0, u)
			want, err := hex.DecodeString(row.After["scout"])
			if err != nil || !bytes.Equal(s.Ship, want) {
				t.Fatal("回收後force0刷新全78 bytes與原版不同")
			}
		})
	}
}

// 受控原版01FE案例包含現在位置與目的地分離。直接前後全紀錄，不拿自然局面補反例。
func TestUpdateSpottingControlledCurrentPosition(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "docs", "internals", "data", "battle-spotting-coordinates.json"))
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		Rows []struct {
			Name          string
			Before, After map[string]string
		}
	}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 6 {
		t.Fatal("原版案例數不同")
	}
	p := filepath.Join("..", "..", "workplace", "logh3-round39-coordinate-a.json")
	raw, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		t.Skip("缺原版受控收據；入口docs/playtest/046-battle-spotting-phase.md")
	}
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != "24fec495ea3ab7884e61564d81b00093c1c28e706583dc4c62c07fdb3b3d1ae4" {
		t.Fatal("原版受控收據身份不同")
	}
	for _, r := range data.Rows {
		t.Run(r.Name, func(t *testing.T) {
			units := []BattleUnit{}
			records := [][]byte{}
			for i, label := range []string{"observer", "candidate"} {
				b, err := hex.DecodeString(r.Before[label])
				if err != nil || len(b) != 78 {
					t.Fatal("原版前態格式不同")
				}
				w := func(n int) int { return int(binary.LittleEndian.Uint16(b[n:])) }
				units = append(units, BattleUnit{Faction: int(b[0]), Squadron: i, Kind: int(b[4]), Out: b[30]&128 != 0,
					X: w(52), Y: w(54), TargetX: w(40), TargetY: w(42), Seen: b[22]&1 != 0, Detail: int(b[23]), DetailHold: int(b[24])})
				records = append(records, b)
			}
			UpdateSpotting(units, nil)
			for i, label := range []string{"observer", "candidate"} {
				got := records[i]
				u := units[i]
				got[22] &^= 1
				if u.Seen {
					got[22] |= 1
				}
				got[23], got[24] = byte(u.Detail), byte(u.DetailHold)
				want, err := hex.DecodeString(r.After[label])
				if err != nil || !bytes.Equal(got, want) {
					t.Errorf("%s完整78-byte不同", label)
				}
			}
		})
	}
}

// 原版直接前後紀錄，僅驗索敵情報欄位。完整第三階段的UI／續接由087另設閘門。
func TestUpdateSpottingDirectOracle(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "docs", "internals", "data", "battle-spotting-phase.json"))
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		Index int
		Hex   string
	}
	var data struct {
		Cases []struct {
			Scenario string
			Frames   []struct {
				Name                   string
				Hashes                 map[string]string
				Units, Planets, Fleets []record
			}
		}
	}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Cases) != 2 {
		t.Fatal("原版局面數不同")
	}
	decode := func(r record, size int) []byte {
		t.Helper()
		b, err := hex.DecodeString(r.Hex)
		if err != nil || len(b) != size {
			t.Fatalf("池%d紀錄長度或hex錯誤", r.Index)
		}
		return b
	}
	w := func(b []byte, n int) int { return int(binary.LittleEndian.Uint16(b[n:])) }
	for _, c := range data.Cases {
		t.Run(c.Scenario, func(t *testing.T) {
			if len(c.Frames) < 2 {
				t.Fatal("缺直接前後紀錄")
			}
			first, last := c.Frames[0], c.Frames[len(c.Frames)-1]
			// 缺原版就明示skip；不拿自製資料或反推前態替代。
			for _, f := range []int{0, len(c.Frames) - 1} {
				name := c.Frames[f].Name + "-mem.bin"
				p := filepath.Join("..", "..", "workplace", "oracle", "battle79-spot-r39-"+c.Scenario+"-a", name)
				m, err := os.ReadFile(p)
				if os.IsNotExist(err) {
					t.Skip("缺原版RAM；重生入口tools/battle_spotting_oracle.sh")
				}
				if err != nil || len(m) != 1048576 || fmt.Sprintf("%x", sha256.Sum256(m)) != c.Frames[f].Hashes[name] {
					t.Fatalf("原版%s身份不同", name)
				}
			}
			wantCount := 135
			if c.Scenario == "s3" {
				wantCount = 55
			}
			if len(first.Units) != wantCount || len(last.Units) != wantCount {
				t.Fatal("原版分隊數不同")
			}
			units := []BattleUnit{}
			pools := []int{}
			fleets := map[int][]byte{}
			for _, r := range first.Fleets {
				fleets[r.Index] = decode(r, 122)
			}
			for _, r := range first.Units {
				b := decode(r, 78)
				// 原版完整清單也保存未放出的偵察隊；正式Game不把它們當觀測者。
				if b[4] == 7 && fleets[int(b[1])][58] == 255 {
					continue
				}
				units = append(units, BattleUnit{Fleet: int(b[1]), Squadron: int(b[3]), Faction: int(b[0]), Kind: int(b[4]),
					Out: b[30]&128 != 0, X: w(b, 52), Y: w(b, 54), TargetX: w(b, 40), TargetY: w(b, 42),
					Seen: b[22]&1 != 0, Detail: int(b[23]), DetailHold: int(b[24])})
				pools = append(pools, r.Index)
			}
			var planets []BattlePlanet
			for _, r := range first.Planets {
				b := decode(r, 82)
				planets = append(planets, BattlePlanet{Owner: int(b[17]), X: w(b, 70), Y: w(b, 72), Defence: w(b, 20)})
			}
			UpdateSpotting(units, planets)
			wants := map[int][]byte{}
			for i, r := range last.Units {
				if r.Index != first.Units[i].Index {
					t.Fatal("原版前後清單不同")
				}
				wants[r.Index] = decode(r, 78)
			}
			for i, pool := range pools {
				want, got := wants[pool], units[i]
				if got.Seen != (want[22]&1 != 0) || got.Detail != int(want[23]) || got.DetailHold != int(want[24]) {
					t.Errorf("池%d情報 got=%v/%d/%d want=%v/%d/%d", pool, got.Seen, got.Detail, got.DetailHold, want[22]&1 != 0, want[23], want[24])
				}
			}
		})
	}
}
