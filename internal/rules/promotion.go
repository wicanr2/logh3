package rules

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

// PromotionSpec 是108核過的功績門檻、兩側軍階及通知文字。原版資料由外部載入。
type PromotionSpec struct {
	Need           []uint16                       `json:"need"`
	RankThresholds []int                          `json:"rank_thresholds"`
	Ranks          [][]struct{ Hex, Text string } `json:"ranks"`
	Formats        map[string]string              `json:"formats"`
}

func LoadPromotionSpec(path string) (*PromotionSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p PromotionSpec
	if err = json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	if err = p.Validate(); err != nil {
		return nil, fmt.Errorf("%s：%w", path, err)
	}
	return &p, nil
}

func (p *PromotionSpec) Validate() error {
	if p == nil || len(p.Need) != 40 || len(p.Ranks) != 2 || len(p.RankThresholds) != len(rankSteps) {
		return fmt.Errorf("升級資料缺40個門檻或兩陣營12級軍階")
	}
	for i, n := range p.Need {
		if n == 0 {
			return fmt.Errorf("升級門檻%d為零", i)
		}
	}
	for i, n := range p.RankThresholds {
		if n != rankSteps[i] {
			return fmt.Errorf("軍階下限%d不同", i)
		}
	}
	for _, f := range p.Ranks {
		if len(f) != 12 {
			return fmt.Errorf("軍階表不完整")
		}
		for _, r := range f {
			b, e := hex.DecodeString(r.Hex)
			if e != nil || len(b) == 0 || len(b)%2 != 0 || r.Text == "" {
				return fmt.Errorf("軍階字串不合法")
			}
		}
	}
	for _, key := range promotionFormats {
		b, e := hex.DecodeString(p.Formats[key])
		if e != nil || len(b) == 0 {
			return fmt.Errorf("升級通知格式%s缺失", key)
		}
	}
	return nil
}

var promotionFormats = []string{"DS:3DE7", "DS:3DEF", "DS:3DFE", "DS:3E0F", "DS:3E1E", "DS:3E2D", "DS:3E3C", "DS:3E4B", "DS:3E5A", "DS:3E69", "DS:3E78", "DS:3E87", "DS:3E96"}

func (p *PromotionSpec) Format(key string) []byte {
	if p == nil {
		return nil
	}
	b, _ := hex.DecodeString(p.Formats[key])
	return b
}

func (p *PromotionSpec) RankCodes(faction, rank int) []uint16 {
	if p == nil || faction < 0 || faction >= len(p.Ranks) || rank < 0 || rank >= len(p.Ranks[faction]) {
		return nil
	}
	b, _ := hex.DecodeString(p.Ranks[faction][rank].Hex)
	out := make([]uint16, len(b)/2)
	for i := range out {
		out[i] = uint16(b[2*i])<<8 | uint16(b[2*i+1])
	}
	return out
}

// PromotionContext 明示全局重算的輸入；Paths保持本次戰略處理開始的快取（研究421）。
type PromotionContext struct {
	Spec     *PromotionSpec
	Kinds    *ShipKindTable
	Mobility []int
	Graph    *StarGraph
	Params   UnrestParams
	Paths    *UnrestPaths
	Military [2]int // 當下軍事費；Game的結構欄位可能比WORK底稿新
}

// PromoteBattleAdmiral 一次成長、指定艦隊重算及全局重算。錯誤不改五表或亂數。
// 保持表的共用底層；存檔欄位只同步本步寫入的四組經費。
func (s *Slot) PromoteBattleAdmiral(fleet, admiral int, c PromotionContext, rnd *Rand) (uint16, error) {
	if err := s.validatePromotion(fleet, admiral, c, rnd); err != nil {
		return 0, err
	}
	r := s.admiral(admiral)
	if r[6] == 99 || binary.LittleEndian.Uint32(r[2:]) < uint32(c.Spec.Need[min(int(r[6]), 39)]) {
		return 0, nil
	}
	a := *s
	a.Admiral, a.Fleet, a.Ship, a.Stellar, a.Work = bytes.Clone(s.Admiral), bytes.Clone(s.Fleet), bytes.Clone(s.Ship), bytes.Clone(s.Stellar), bytes.Clone(s.Work)
	next := NewRand(rnd.Seed())
	flags := promoteAdmiral(a.admiral(admiral), c.Spec.Need, next)
	a.recalculatePromotionFleet(fleet, c)
	a.recalculatePromotionGlobal(c)
	for _, p := range [][2][]byte{{s.Admiral, a.Admiral}, {s.Fleet, a.Fleet}, {s.Ship, a.Ship}, {s.Stellar, a.Stellar}, {s.Work, a.Work}} {
		copy(p[0], p[1])
	}
	w, _ := ReadWork(s.Work)
	s.ShipBudget, s.TaxTotal, s.Upkeep, s.MoveCost = w.ShipBudget, w.TaxTotal, w.Upkeep, w.MoveCost
	rnd.SetSeed(next.Seed())
	return flags, nil
}

func (s *Slot) validatePromotion(fleet, admiral int, c PromotionContext, rnd *Rand) error {
	if s == nil || len(s.Admiral) != 82*AdmiralRecord || len(s.Fleet) != FleetCount*FleetRecord || len(s.Ship) != ShipCount*ShipRecord || len(s.Stellar) != 32*RegionRecord || len(s.Work) != WorkRecord || s.Planets == nil || rnd == nil || c.Graph == nil || c.Paths == nil {
		return fmt.Errorf("升級所需五表、行星、路徑或亂數不完整")
	}
	if err := c.Spec.Validate(); err != nil {
		return err
	}
	if fleet < 0 || fleet >= FleetCount || admiral < 0 || admiral >= 82 || s.Fleet[fleet*FleetRecord+4] == 255 {
		return fmt.Errorf("升級艦隊或提督索引不合法")
	}
	f := s.fleet(fleet)
	if admiral != int(f[4]) && admiral != int(f[5]) {
		return fmt.Errorf("升級提督不屬於指定艦隊")
	}
	for i := 0; i < 82; i++ {
		if s.admiral(i)[1] > 1 {
			return fmt.Errorf("提督陣營不合法")
		}
	}
	for f := 0; f < 2; f++ {
		if len(c.Paths.Dist[f]) != 32 || len(c.Paths.Mid[f]) != 32 || c.Params.Ruler[f] < 0 || c.Params.Ruler[f] >= 82 || c.Params.Blend[f] < 0 || c.Params.Blend[f] >= len(c.Params.Table) || c.Params.Rate[f] < 0 || c.Params.Rate[f] >= len(c.Params.Table) || c.Params.Rate[f] >= len(c.Params.Base) || c.Params.Base[c.Params.Rate[f]] <= 0 {
			return fmt.Errorf("騷動參數或路徑不合法")
		}
		for i := 0; i < 32; i++ {
			if len(c.Paths.Dist[f][i]) != 32 || len(c.Paths.Mid[f][i]) != 32 {
				return fmt.Errorf("路徑尺寸不合法")
			}
		}
	}
	for i := 0; i < 32; i++ {
		if s.Stellar[i*RegionRecord+0x12] > 2 {
			return fmt.Errorf("星域陣營不合法")
		}
		f := s.fleet(i)
		if f[4] == 255 {
			continue
		}
		if f[1] > 1 || int(f[4]) >= 82 || (f[5] != 255 && int(f[5]) >= 82) || f[15] > 10 {
			return fmt.Errorf("艦隊%d的陣營、提督或編制不合法", i)
		}
		if i == fleet && len(c.Mobility) <= int(f[15]) {
			return fmt.Errorf("缺編制機動力")
		}
		for g := 0; g < int(f[15]); g++ {
			pool := int(binary.LittleEndian.Uint16(f[0x10+4*g:]))
			if pool >= ShipCount {
				return fmt.Errorf("艦艇池索引越界")
			}
			sh := s.Ship[pool*ShipRecord : (pool+1)*ShipRecord]
			kind, ok := c.Kinds.Kind(int(sh[0]), int(sh[4]))
			if sh[0] > 1 || sh[4] >= 8 || !ok || len(kind.Words) != 13 || (i == fleet && binary.LittleEndian.Uint16(sh[8:]) != 0 && binary.LittleEndian.Uint16(sh[10:]) == 0) {
				return fmt.Errorf("艦種、陣營或滿編分母不合法")
			}
		}
		if i == fleet {
			pool := int(binary.LittleEndian.Uint16(f[0x38:]))
			if pool >= ShipCount {
				return fmt.Errorf("索敵池索引越界")
			}
			sh := s.Ship[pool*ShipRecord : (pool+1)*ShipRecord]
			kind, ok := c.Kinds.Kind(int(sh[0]), 7)
			if sh[0] > 1 || !ok || len(kind.Words) != 13 {
				return fmt.Errorf("索敵艦種資料不完整")
			}
		}
	}
	return nil
}

func promoteAdmiral(r []byte, need []uint16, rnd *Rand) uint16 {
	level := int(r[6])
	binary.LittleEndian.PutUint32(r[2:], binary.LittleEndian.Uint32(r[2:])-uint32(need[min(level, 39)]))
	r[6]++
	flags, count := uint16(1), 0
	if RankOf(int(r[6])) != RankOf(level) {
		flags |= 2
	}
	grow := func(off, low, high int, bit uint16) {
		if r[off] >= 200 {
			return
		}
		span := high
		if r[off] < 100 {
			span = low
		}
		n := (rnd.Next()%(span+1) + rnd.Next()%(span+1)) / 100
		if n > 0 {
			r[off] += byte(n)
			flags |= bit
			count++
		}
	}
	grow(0x0b, 300, 50, 4)
	if rnd.Next()%2 != 0 {
		grow(0x13, 200, 100, 8)
	} else {
		grow(0x12, 200, 100, 16)
	}
	switch rnd.Next() % 3 {
	case 0:
		grow(0x0f, 100, 50, 32)
	case 1:
		grow(0x10, 100, 50, 64)
	case 2:
		grow(0x11, 100, 50, 128)
	}
	if count < 3 {
		if rnd.Next()%2 != 0 {
			grow(0x0c, 200, 50, 256)
		} else {
			grow(0x0d, 200, 50, 512)
		}
	}
	if count < 3 {
		if rnd.Next()%2 != 0 {
			grow(0x0e, 100, 50, 1024)
		} else {
			grow(0x17, 200, 50, 2048)
		}
	}
	return flags
}

// 以下保持297B:05B2／0E8B的byte、低word及unsigned32運算（研究419第二版）。
func (s *Slot) recalculatePromotionFleet(index int, c PromotionContext) {
	f := s.fleet(index)
	a := s.admiral(int(f[4]))
	d := s.admiral(int(f[5]))
	f[14] = byte(FleetSlots(int(a[6])))
	ab := FleetAbilities(a, d)
	copy(f[0x3d:0x48], ab[:])
	minimum := [4]byte{255, 255, 255, 255}
	integrity, living := 0, 0
	p := int(int16(uint16((270-2*int(f[0x46]))*c.Mobility[int(f[15])]))) / 100
	for i := 0; i < int(f[15]); i++ {
		pool := int(binary.LittleEndian.Uint16(f[0x10+4*i:]))
		sh := s.Ship[pool*ShipRecord : (pool+1)*ShipRecord]
		kind, _ := c.Kinds.Kind(int(sh[0]), int(sh[4]))
		base := [3]int{kind.Byte(8) * int(f[0x43]) / 100, kind.Byte(6) * int(f[0x42]) / 80, kind.Byte(7) * int(f[0x42]) / 60}
		for j, off := range []int{0x1f, 0x21, 0x22} {
			v := byte(base[j])
			sh[off] = v - byte(uint16(uint32(v)*uint32(uint16(p)))/100)
		}
		for j, off := range []int{0x1f, 0x21, 0x22, 0x23} {
			minimum[j] = min(minimum[j], sh[off])
		}
		alive, full := binary.LittleEndian.Uint16(sh[8:]), binary.LittleEndian.Uint16(sh[10:])
		if alive != 0 {
			loss := uint16((uint32(full-alive) * uint32(200-int(a[0xb]))) / 100)
			loss = min(loss, full)
			sh[6] = byte(uint32(full-loss) * 100 / uint32(full))
			integrity += int(sh[6])
			living++
		}
	}
	copy(f[0x4d:0x51], minimum[:])
	f[0x3c] = 100
	scout := int(binary.LittleEndian.Uint16(f[0x38:]))
	sh := s.Ship[scout*ShipRecord : (scout+1)*ShipRecord]
	kind, _ := c.Kinds.Kind(int(sh[0]), 7)
	sh[0x1f], sh[0x21], sh[0x22] = byte(kind.Byte(8)), byte(kind.Byte(6)), byte(kind.Byte(7))
	if living == 0 {
		f[0] |= 8
		return
	}
	f[0] &^= 8
	f[0x3c] = byte(integrity / living)
	old := f[0]
	if f[0x3c] < 50 {
		f[0] |= 1
	} else {
		f[0] &^= 1
	}
	if (old^f[0])&1 != 0 {
		f[0] |= 4
	}
}

func (s *Slot) recalculatePromotionGlobal(c PromotionContext) {
	ResumRegionEconomy(s.Stellar, s.Planets)
	u := UnrestState{Regions: s.Stellar, Planets: s.Planets, Fleets: s.Fleet, Admirals: s.Admiral, Graph: c.Graph, Params: c.Params, Paths: c.Paths}
	tax, upkeep, moves := [2]int{}, [2]int{}, [2]int{}
	for i := 0; i < 32; i++ {
		r := s.Stellar[i*RegionRecord : (i+1)*RegionRecord]
		if r[0x12] < 2 {
			tax[int(r[0x12])] += RegionTax(r)
		}
		r[0x7d] = byte(u.Unrest(i))
		f := s.fleet(i)
		if f[4] == 255 {
			continue
		}
		var sums [2]uint32
		for g := 0; g < int(f[15]); g++ {
			pool := int(binary.LittleEndian.Uint16(f[0x10+4*g:]))
			sh := s.Ship[pool*ShipRecord : (pool+1)*ShipRecord]
			kind, _ := c.Kinds.Kind(int(sh[0]), int(sh[4]))
			for j := range sums {
				sums[j] += uint32(kind.Byte(4+j)) * uint32(binary.LittleEndian.Uint16(sh[8:]))
			}
		}
		for j, n := range sums {
			binary.LittleEndian.PutUint16(f[0xa+2*j:], uint16((n/1000*uint32(200-int(f[0x47])))/100))
		}
		seat := int(f[1])
		upkeep[seat] += int(int16(binary.LittleEndian.Uint16(f[0xa:])))
		if f[8] != 2 && f[8] != 5 {
			moves[seat] += int(int16(binary.LittleEndian.Uint16(f[0xc:])))
		}
	}
	for f := 0; f < 2; f++ {
		for _, v := range []struct{ o, n int }{{0x56, tax[f]}, {0x5e, upkeep[f]}, {0x6e, moves[f]}} {
			binary.LittleEndian.PutUint32(s.Work[v.o+4*f:], uint32(v.n))
		}
		available := int(int32(uint32(c.Military[f]) + uint32(tax[f]) - uint32(upkeep[f]) - uint32(moves[f])))
		budget := 0
		if available >= 21474836 {
			budget = available / 100 * int(s.Work[0x7e+f])
		} else if available >= 0 {
			budget = available * int(s.Work[0x7e+f]) / 100
		}
		binary.LittleEndian.PutUint32(s.Work[0x46+4*f:], uint32(budget))
	}
}
