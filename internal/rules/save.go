package rules

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// 存檔的版面（`docs/internals/30`、`docs/re/41`、`docs/re/42`）。
//
// 存檔沒有自己的格式——它就是執行期那幾張表原樣寫出去。所以這裡的
// 位移與 `region.go` 的是同一套，載進來之後不必再轉換。
const (
	// SaveSlots 是存檔格數，SaveEntry 是 `SAVEDATA.IDX` 一筆的長度。
	SaveSlots        = 11
	SaveEntry        = 48
	saveEntryNum     = 0x0D // 存檔序號
	saveEntryTurn    = 0x0E
	saveEntryName    = 0x14
	saveEntryNameLen = 18

	// WorkRecord 是 `WORK.DT?` 的長度，後面三個是對上的欄位。
	WorkRecord   = 230
	workTurn     = 0x04 // 16 位元
	workMilitary = 0x3E // 32 位元 × 陣營
	workTaxTotal = 0x56 // 32 位元 × 陣營
	workBlend    = 0xC0 // 8 位元 × 陣營（＝ ds:8FD4，`docs/internals/55`）
	workShipPct  = 0x7E // 8 位元 × 陣營（＝ ds:8F92，造船百分比）
	workBuildPct = 0xC2 // 7 bytes × 陣營（＝ ds:8FD6，七個艦種的分配）

	// FleetRecord 是艦隊表一筆的長度（`docs/re/40`）。
	FleetRecord = 122
	FleetCount  = 32

	// fleetFlags 的 bit0 把艦隊分成兩類：有設的在原版 `2E54:005D` 當場
	// 處理完，沒設的留給後面五個階段挑（`docs/internals/33`、`docs/re/196`）。
	fleetFlags   = 0x00
	fleetFaction = 0x01
	fleetNumber  = 0x02 // 艦隊編號
	// fleetUnused 同時是**提督索引**：`113A:05CF` 拿它 ×28 去查
	// 執行期提督表 `ds:B77C`（`docs/re/77`）。沒有提督就是沒在用。
	fleetUnused   = 0x04 // FF ＝ 這一格沒在用
	fleetAdjutant = 0x05
	fleetRegion   = 0x06
	fleetTarget   = 0x07 // 要去的星域；FF ＝ 沒有命令
	fleetState    = 0x08 // 移動狀態；5 ＝ 不動（`docs/re/77`）
	fleetMorale   = 0x3C // 艦隊士氣值（建立艦隊時寫 100，戰鬥中會變）
	fleetAction   = 0x48 // 低四位 ＝ 艦隊行動（0 移動／1 戰鬥／2 防禦／3 占領／4 待命）
	fleetClassKey = 0x09 // 艦隊類別鍵 0–15，經 ds:5F52 換成 0–3（docs/re/190）
	fleetExpected = 0x0E // 編制槽的**應有**數；+0x0F 是現有數（docs/re/194）
	fleetGroups   = 0x0F // 這一支有幾組編制
	fleetSlots    = 0x10 // 十個 4-byte 的槽，前兩 byte 是艦艇索引
	fleetScout    = 0x38 // 索敵分遣隊的艦艇池索引（不在編制表裡）
	fleetSlotN    = 10

	// fleetParked 是 `+0x08` 代表「這一支不會動」的值（`113A:0350`）。
	fleetParked = 5

	// ShipRecord 是艦艇池一筆的長度，ShipCount 是筆數（`docs/re/42`）。
	ShipRecord = 78
	ShipCount  = 352
)

// SaveIndexEntry 是 `SAVEDATA.IDX` 裡的一格。
type SaveIndexEntry struct {
	// Index 是格號 0–10，也就是檔案副檔名最後一個字（十六進位）。
	Index int

	// Used ＝ 這一格有存檔。沒用到的格整格是零。
	Used bool

	// Num 是存檔序號，Turn 是回合數。
	Num, Turn int

	// Name 是存檔名稱，Big5 字碼（右邊補的半形空白已經去掉）。
	Name []uint16

	// Stamp 是記錄開頭那六個 byte，畫面上印成「93/12/03 00:00:00」。
	// 三格存檔完全一樣，所以**怎麼解還沒定**（只有一組樣本）：
	// 目前照「年 BCD、月在高四位、日 BCD、時分秒 BCD」讀得出畫面上的字。
	Stamp [6]byte
}

// StampFields 把 Stamp 拆成畫面上那六個數字。解法未驗（見 Stamp）。
func (e SaveIndexEntry) StampFields() (year, month, day, hour, min, sec int) {
	bcd := func(b byte) int { return int(b>>4)*10 + int(b&0x0F) }
	return bcd(e.Stamp[0]), int(e.Stamp[1] >> 4), bcd(e.Stamp[2]),
		bcd(e.Stamp[3]), bcd(e.Stamp[4]), bcd(e.Stamp[5])
}

// SlotExt 回這一格對應的副檔名末字：0–9 用數字，10 用 `A`。
func SlotExt(index int) string {
	if index < 0 || index >= SaveSlots {
		return ""
	}
	return string("0123456789A"[index])
}

// ParseSaveIndex 讀 `SAVEDATA.IDX`。
func ParseSaveIndex(b []byte) ([]SaveIndexEntry, error) {
	if len(b) < SaveSlots*SaveEntry {
		return nil, fmt.Errorf("SAVEDATA.IDX 只有 %d bytes，要 %d",
			len(b), SaveSlots*SaveEntry)
	}
	out := make([]SaveIndexEntry, SaveSlots)
	for i := range out {
		r := b[i*SaveEntry : (i+1)*SaveEntry]
		e := SaveIndexEntry{Index: i}
		for _, v := range r {
			if v != 0 {
				e.Used = true
				break
			}
		}
		if e.Used {
			e.Num = int(r[saveEntryNum])
			e.Turn = int(r[saveEntryTurn]) | int(r[saveEntryTurn+1])<<8
			e.Name = saveName(r[saveEntryName : saveEntryName+saveEntryNameLen])
			copy(e.Stamp[:], r[:6])
		}
		out[i] = e
	}
	return out, nil
}

// saveName 把名稱欄位讀成 Big5 字碼。名稱右邊補半形空白補滿 18 bytes。
func saveName(b []byte) []uint16 {
	var out []uint16
	for i := 0; i+1 < len(b); i += 2 {
		if b[i] == 0 || b[i] == ' ' {
			break
		}
		out = append(out, uint16(b[i])<<8|uint16(b[i+1]))
	}
	return out
}

// SlotFiles 是一格存檔的六個檔，外加 `PLADAT.DAT`。
//
// 行星表存檔沒有檔頭——「第 k 個星域有幾顆行星」是靜態的，不隨局面變，
// 所以那 64 bytes 要從 `PLADAT.DAT` 借（`docs/internals/30`）。
type SlotFiles struct {
	Stellar, Planet, Admiral, Fleet, Ship, Work []byte
	PlanetHeader                                []byte // `PLADAT.DAT`
}

// Slot 是載進來的一格存檔。
type Slot struct {
	// Stellar 是星域表（32 × 126），Planets 是行星表。
	Stellar []byte
	Planets *Planets

	// Admiral／Fleet／Ship 是提督表、艦隊表與艦艇池，先原樣留著。
	Admiral, Fleet, Ship []byte

	// Work 是 `WORK.DT?` 的 230 bytes。只有四個欄位解出來了
	// （回合數、軍事費、總稅收、`Blend`），其餘原樣留著——存回去的時候
	// 要照抄，不然會把還沒解的東西寫壞。
	Work []byte

	// Turn 是回合數。Military（軍事費）、TaxTotal（總稅收）與 Blend
	// **都是每個陣營一格**，索引就是陣營——原版的 `ds:8F52`、`ds:8F6A`
	// 與 `ds:8FD4` 也是（`docs/internals/27`、`docs/internals/55`）。
	Turn                      int
	Military, TaxTotal, Blend [2]int

	// ShipPct 是切給造船的百分比（`ds:8F92`），BuildPct 是那筆錢在七個
	// 艦種之間的分配（`ds:8FD6`）。兩個都是每陣營一份
	// （`docs/internals/27`、`docs/re/256`、`docs/re/257`）。
	ShipPct  [2]int
	BuildPct [2][BuildPlanRows]int
}

// LoadSlot 把一格存檔組成遊戲狀態。
//
// **未安定度不會是對的**：那一欄原版在載入時重算，算式還沒解
// （`docs/re/24`、`docs/re/41`），這裡留的是檔案裡的值。要畫面對得上
// 的地方得自己用量到的值覆蓋（`SetUnrest`）。
func LoadSlot(f SlotFiles) (*Slot, error) {
	if len(f.Stellar) < RegionCount*RegionRecord {
		return nil, fmt.Errorf("STELLAR 只有 %d bytes，要 %d",
			len(f.Stellar), RegionCount*RegionRecord)
	}
	if len(f.Work) < WorkRecord {
		return nil, fmt.Errorf("WORK 只有 %d bytes，要 %d", len(f.Work), WorkRecord)
	}
	if len(f.PlanetHeader) < PlanetHeader {
		return nil, fmt.Errorf("缺 PLADAT.DAT 的 %d bytes 檔頭", PlanetHeader)
	}
	if len(f.Fleet) < FleetCount*FleetRecord {
		return nil, fmt.Errorf("FLEET 只有 %d bytes，要 %d",
			len(f.Fleet), FleetCount*FleetRecord)
	}
	if len(f.Ship) < ShipCount*ShipRecord {
		return nil, fmt.Errorf("SHIP 只有 %d bytes，要 %d",
			len(f.Ship), ShipCount*ShipRecord)
	}
	raw := make([]byte, PlanetHeader, PlanetHeader+len(f.Planet))
	copy(raw, f.PlanetHeader[:PlanetHeader])
	raw = append(raw, f.Planet...)
	planets, err := LoadPlanets(raw)
	if err != nil {
		return nil, err
	}
	s := &Slot{
		Stellar:  append([]byte(nil), f.Stellar...),
		Planets:  planets,
		Admiral:  f.Admiral,
		Fleet:    f.Fleet,
		Ship:     f.Ship,
		Work:     append([]byte(nil), f.Work[:WorkRecord]...),
		Turn:     int(f.Work[workTurn]) | int(f.Work[workTurn+1])<<8,
		Military: [2]int{u32(f.Work, workMilitary), u32(f.Work, workMilitary+4)},
		TaxTotal: [2]int{u32(f.Work, workTaxTotal), u32(f.Work, workTaxTotal+4)},
		Blend:    [2]int{int(f.Work[workBlend]), int(f.Work[workBlend+1])},
		ShipPct:  [2]int{int(f.Work[workShipPct]), int(f.Work[workShipPct+1])},
	}
	for faction := 0; faction < 2; faction++ {
		for row := 0; row < BuildPlanRows; row++ {
			s.BuildPct[faction][row] = int(f.Work[workBuildPct+faction*BuildPlanRows+row])
		}
	}
	return s, nil
}

func u32(b []byte, o int) int {
	return int(b[o]) | int(b[o+1])<<8 | int(b[o+2])<<16 | int(b[o+3])<<24
}

// Region 回第 i 個星域的記錄。
func (s *Slot) Region(i int) []byte {
	if i < 0 || (i+1)*RegionRecord > len(s.Stellar) {
		return nil
	}
	return s.Stellar[i*RegionRecord : (i+1)*RegionRecord]
}

// SetUnrest 覆蓋一個星域的未安定度。
func (s *Slot) SetUnrest(region, unrest int) {
	if r := s.Region(region); r != nil {
		r[RegionUnrest] = byte(unrest)
	}
}

// RecomputeUnrest 照原版載入時的算式把整份未安定度重算一遍
// （規格 `docs/spec/006`）。存檔裡那一欄是存檔當下的值，原版載入時
// 不用它——所以讀完存檔要跑這一支，畫面上的治安才會對。
//
// `params` 不在存檔裡（是難度與統治者那幾個全域），要另外給。
func (s *Slot) RecomputeUnrest(g *StarGraph, params UnrestParams) {
	st := &UnrestState{
		Regions: s.Stellar, Planets: s.Planets, Fleets: s.Fleet,
		Admirals: s.Admiral, Graph: g, Params: params,
	}
	out := make([]int, RegionCount)
	for i := range out {
		out[i] = st.Unrest(i)
	}
	for i, v := range out {
		s.SetUnrest(i, v)
	}
}

// FleetUsed ＝ 第 i 格艦隊有在用。沒用的整筆填 FF，只留編號（`docs/re/40`）。
func (s *Slot) FleetUsed(i int) bool {
	r := s.fleet(i)
	return r != nil && r[fleetUnused] != 0xFF
}

// FleetFaction 回艦隊的陣營，FleetRegion 回它在哪個星域。
func (s *Slot) FleetFaction(i int) int { return int(s.fleet(i)[fleetFaction]) }

// FleetRegion 回艦隊所在的星域編號。
func (s *Slot) FleetRegion(i int) int { return int(s.fleet(i)[fleetRegion]) }

// Fleets 把存檔裡的艦隊表讀成 `[]Fleet`（`docs/internals/33`）。
//
// **沒在用的那幾格不收**：原版整筆填 `FF`，只留編號。
// `w` 給了就順便算每一支的戰力（原版的算式，`docs/re/79`）；nil 就留 0。
func (s *Slot) Fleets(w StrengthWeights) []Fleet {
	var out []Fleet
	for i := 0; i < FleetCount; i++ {
		if !s.FleetUsed(i) {
			continue
		}
		r := s.fleet(i)
		f := Fleet{
			Index:    i,
			Number:   int(r[fleetNumber]),
			Faction:  int(r[fleetFaction]),
			Admiral:  int(r[fleetUnused]),
			Adjutant: int(r[fleetAdjutant]),
			Region:   int(r[fleetRegion]),
			Target:   NoTarget,
			Parked:   r[fleetState] == fleetParked,
			Action:   int(r[fleetAction] & 0x0F),
			Scout:    int(r[fleetScout]) | int(r[fleetScout+1])<<8,
			Morale:   int(r[fleetMorale]),
			Flags:    int(r[fleetFlags]),
			Move:     int(r[fleetState]),
			Groups:   int(r[fleetGroups]),
			Expected: int(r[fleetExpected]),
			ClassKey: int(r[fleetClassKey]),
		}
		if r[fleetAdjutant] == 0xFF {
			f.Adjutant = -1
		}
		if t := r[fleetTarget]; t != 0xFF {
			f.Target = int(t)
		}
		if w != nil {
			f.Strength = s.FleetStrength(w, i)
		}
		out = append(out, f)
	}
	return out
}

// FleetShips 回這支艦隊十個槽指到的艦艇索引（`docs/re/42`）。
func (s *Slot) FleetShips(i int) []int {
	r := s.fleet(i)
	if r == nil {
		return nil
	}
	var out []int
	for k := 0; k < fleetSlotN; k++ {
		o := fleetSlots + k*4
		if r[o] == 0xFF {
			continue
		}
		out = append(out, int(r[o])|int(r[o+1])<<8)
	}
	return out
}

// FleetGroupShips 回這支艦隊**現在編制裡**那幾組指到的艦艇索引。
//
// 與 `FleetShips` 的差別是**上界**：這一支只走前 `+0x0F` 組，
// `FleetShips` 走完十個槽、只跳過 `0xFF`。編制縮編過的艦隊，後面幾個槽
// 還留著舊的艦艇索引（`FLEET.DT8` 的第 26 支：`+0x0F` ＝ 1，
// 卻有 7 個槽不是 `0xFF`）。
//
// **原版的戰力與候選評分都以 `+0x0F` 為界**（`2E54:1876`、`2E54:1574`，
// `docs/internals/54`、`docs/re/168`），所以那兩條算式要用這一支。
func (s *Slot) FleetGroupShips(i int) []int {
	r := s.fleet(i)
	if r == nil {
		return nil
	}
	n := int(r[fleetGroups])
	if n > fleetSlotN {
		n = fleetSlotN
	}
	var out []int
	for k := 0; k < n; k++ {
		o := fleetSlots + k*4
		if r[o] == 0xFF {
			continue
		}
		out = append(out, int(r[o])|int(r[o+1])<<8)
	}
	return out
}

// Ships 回第 i 筆艦艇記錄。
func (s *Slot) ShipAt(i int) []byte {
	if i < 0 || (i+1)*ShipRecord > len(s.Ship) {
		return nil
	}
	return s.Ship[i*ShipRecord : (i+1)*ShipRecord]
}

func (s *Slot) fleet(i int) []byte {
	if i < 0 || (i+1)*FleetRecord > len(s.Fleet) {
		return nil
	}
	return s.Fleet[i*FleetRecord : (i+1)*FleetRecord]
}

// SlotFileNames 回一格存檔的六個檔名。
func SlotFileNames(index int) []string {
	e := SlotExt(index)
	return []string{"STELLAR.DT" + e, "PLANET.DT" + e, "ADMIRAL.DT" + e,
		"FLEET.DT" + e, "SHIP.DT" + e, "WORK.DT" + e}
}

// Save 把這一格存檔寫進 `dir`（規格 `docs/spec/009`）。
//
// 存檔就是執行期那幾張表原樣寫出去（`docs/re/41`），所以這裡沒有格式轉換
// ——把手上的表直接寫成檔。`WORK` 只改解出來的那幾個欄位，其餘照抄，
// 因為剩下的還沒解，動它等於把不懂的東西寫壞。
//
// **不會寫進原版目錄**：`dir` 由呼叫端指定，原版檔案一律唯讀。
func (s *Slot) Save(dir string, index int) error {
	if len(s.Work) < WorkRecord {
		return fmt.Errorf("WORK 只有 %d bytes，寫不出去", len(s.Work))
	}
	work := append([]byte(nil), s.Work...)
	work[workTurn], work[workTurn+1] = byte(s.Turn), byte(s.Turn>>8)
	for f := 0; f < 2; f++ {
		put32(work, workMilitary+4*f, s.Military[f])
		put32(work, workTaxTotal+4*f, s.TaxTotal[f])
		work[workBlend+f] = byte(s.Blend[f])
		work[workShipPct+f] = byte(s.ShipPct[f])
		for row := 0; row < BuildPlanRows; row++ {
			work[workBuildPct+f*BuildPlanRows+row] = byte(s.BuildPct[f][row])
		}
	}
	planet := s.Planets.Data
	if len(planet) < PlanetHeader {
		return fmt.Errorf("行星表只有 %d bytes", len(planet))
	}
	for i, b := range [][]byte{s.Stellar, planet[PlanetHeader:], s.Admiral,
		s.Fleet, s.Ship, work} {
		name := SlotFileNames(index)[i]
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func put32(b []byte, o, v int) {
	b[o], b[o+1] = byte(v), byte(v>>8)
	b[o+2], b[o+3] = byte(v>>16), byte(v>>24)
}

// WriteSaveIndex 把一格的目錄資料寫回 `SAVEDATA.IDX` 的位元組。
//
// 只動回合數與名稱兩欄——其餘（開頭那 12 bytes、序號、`FF FF`）原樣留著，
// 語意還沒解完（`docs/internals/30`）。
func WriteSaveIndex(idx []byte, index int, turn int, name []uint16) error {
	if index < 0 || index >= SaveSlots || len(idx) < SaveSlots*SaveEntry {
		return fmt.Errorf("存檔目錄寫不到第 %d 格", index)
	}
	r := idx[index*SaveEntry : (index+1)*SaveEntry]
	r[saveEntryTurn], r[saveEntryTurn+1] = byte(turn), byte(turn>>8)
	for i := 0; i < saveEntryNameLen; i += 2 {
		r[saveEntryName+i], r[saveEntryName+i+1] = ' ', ' '
	}
	for i, c := range name {
		if saveEntryName+i*2+1 >= saveEntryName+saveEntryNameLen {
			break
		}
		r[saveEntryName+i*2] = byte(c >> 8)
		r[saveEntryName+i*2+1] = byte(c)
	}
	return nil
}

// LoadWorkTemplate 讀新遊戲開局的 `WORK` 區塊。
//
// 存檔的 `WORK.DT?` 就是 `ds:8F14` 起那 230 bytes（`docs/re/45`），
// 而 `WORK.DTX` 不存在——新遊戲那一份只能從記憶體量。這支讀的就是量到的
// 那一份（`docs/internals/data/work-796-01.json`）。
func LoadWorkTemplate(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d struct {
		Hex string `json:"hex"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	out, err := hex.DecodeString(d.Hex)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(out) != WorkRecord {
		return nil, fmt.Errorf("%s: WORK 有 %d bytes，要 %d", path, len(out), WorkRecord)
	}
	return out, nil
}
