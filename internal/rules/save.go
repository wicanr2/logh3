package rules

import "fmt"

// 存檔的版面（`docs/internals/30`、`docs/re/41`、`docs/re/42`）。
//
// 存檔沒有自己的格式——它就是執行期那幾張表原樣寫出去。所以這裡的
// 位移與 `region.go` 的是同一套，載進來之後不必再轉換。
const (
	// SaveSlots 是存檔格數，SaveEntry 是 `SAVEDATA.IDX` 一筆的長度。
	SaveSlots    = 11
	SaveEntry    = 48
	saveEntryNum = 0x0D // 存檔序號
	saveEntryTurn = 0x0E
	saveEntryName = 0x14
	saveEntryNameLen = 18

	// WorkRecord 是 `WORK.DT?` 的長度，後面三個是對上的欄位。
	WorkRecord   = 230
	workTurn     = 0x04 // 16 位元
	workMilitary = 0x3E // 32 位元
	workTaxTotal = 0x56 // 32 位元

	// FleetRecord 是艦隊表一筆的長度（`docs/re/40`）。
	FleetRecord = 122
	FleetCount  = 32

	fleetFaction = 0x01
	fleetUnused  = 0x04 // FF ＝ 這一格沒在用
	fleetRegion  = 0x06
	fleetSlots   = 0x10 // 十個 4-byte 的槽，前兩 byte 是艦艇索引
	fleetSlotN   = 10

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

	// Turn 是回合數，Military 是軍事費，TaxTotal 是總稅收。
	Turn, Military, TaxTotal int
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
		Turn:     int(f.Work[workTurn]) | int(f.Work[workTurn+1])<<8,
		Military: u32(f.Work, workMilitary),
		TaxTotal: u32(f.Work, workTaxTotal),
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
