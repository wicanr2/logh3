//go:build ignore

// 第116輪唯讀觀測片段，由replay產生器附加到暫存Game檔。
package game

// natural116CleanupEvents只保存輸出，不寫局面、命令、鏡頭或亂數。
var natural116CleanupEvents []ejson.RawMessage
var natural116CleanupSeen = map[string]bool{}
var natural116CleanupObservation = func(g *Game, slot *rules.Slot, index int, stage string, mapping []int) {
	key := fmt.Sprintf("%d:%d:%d:%d", g.turn, g.battleRegion, g.battleTurn, index)
	if stage == "before" {
		dead, flagship := false, false
		for _, u := range g.battleUnits {
			if u.Fleet == index {
				dead = dead || u.Squadron > 0 && u.Ships == 0
				flagship = flagship || u.Squadron == 0 && u.Ships > 0
			}
		}
		if !dead || !flagship {
			return
		}
		if natural116CleanupSeen[key] {
			panic("正常空組事件重複 " + key)
		}
		natural116CleanupSeen[key] = true
	} else if !natural116CleanupSeen[key] {
		return
	}
	tables := map[string]string{}
	for name, data := range map[string][]byte{"fleet": slot.Fleet, "ship": slot.Ship, "admiral": slot.Admiral,
		"work": g.workRecord(), "stellar": g.res.Stellar, "planet": g.res.Planets.Data} {
		tables[name] = fmt.Sprintf("%x", data)
	}
	b, err := ejson.Marshal(map[string]any{"key": key, "stage": stage, "index": index, "turn": g.turn,
		"region": g.battleRegion, "battle_turn": g.battleTurn, "seed": g.battleRand.Seed(),
		"mapping": mapping, "tables": tables, "units": g.battleUnits, "fleets": g.fleets})
	if err != nil {
		panic(err)
	}
	natural116CleanupEvents = append(natural116CleanupEvents, b)
}

func (g *Game) BattleCleanupProbeEvents() []ejson.RawMessage {
	return append([]ejson.RawMessage(nil), natural116CleanupEvents...)
}
