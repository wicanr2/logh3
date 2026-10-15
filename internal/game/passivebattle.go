package game

import "github.com/wicanr2/logh3/internal/rules"

// 明示的歷史被動模式不造成任何傷害（使用者裁決 2026-09-17：「強制改變 remake AI 進入戰場後都不攻擊」，
// 範圍「完全不造成傷害」）。**這是刻意偏離原版**：原版規則層（`rules.ResolveBattleTurn`、
// `FortressGunFire`、`PlanetAssault`）一行不改，差異全在這一層包起來，對拍原版戰場的測試用
// `originalBattleAI` 關掉它。
//
//   - 電腦方的戰術 AI 不下令：電腦的艦隊留在原地。
//   - 行星互打：能傷到玩家分隊的只有電腦的行星，呼叫之後把玩家分隊還原，電腦那一方的得分拿掉。
//   - 要塞砲：電腦的要塞這一步當作沒瞄準（狀態呼叫之後還原）。
//   - 打行星：電腦艦隊不算「戰鬥」命令；玩家打電腦行星時挨的反擊，同樣把玩家分隊還原。
//
// 對戰模式兩邊都是人，照原版。
// 2026-10-06使用者改為啟用兩種戰術AI；正常預設不套此模式（規格120）。

// SetPassiveBattleComputer只改執行期政策，不改六表存檔或AI模式。
func (g *Game) SetPassiveBattleComputer(passive bool) { g.passiveBattleComputer = passive }

// passiveComputer 回這一場要不要套「電腦不造成傷害」，以及電腦是哪一方。
func (g *Game) passiveComputer() (int, bool) {
	if !g.passiveBattleComputer || g.originalBattleAI || g.versus || g.faction < 0 || g.faction > 1 {
		return 0, false
	}
	return 1 - g.faction, true
}

// snapshotSide 複製某一方所有分隊，restoreSide 放回去。
func (g *Game) snapshotSide(faction int) map[int]rules.BattleUnit {
	out := map[int]rules.BattleUnit{}
	for i, u := range g.battleUnits {
		if u.Faction == faction {
			out[i] = u
		}
	}
	return out
}

func (g *Game) restoreSide(saved map[int]rules.BattleUnit) {
	for i, u := range saved {
		if i < len(g.battleUnits) {
			g.battleUnits[i] = u
		}
	}
}

// resolveBattleTurn 是 `rules.ResolveBattleTurn` 加上電腦不造成傷害。
func (g *Game) resolveBattleTurn() []rules.BattleLoss {
	cpu, passive := g.passiveComputer()
	if !passive {
		return rules.ResolveBattleTurn(g.battlePlanets, g.battleFleets, g.battleUnits, g.battleRand)
	}
	saved := g.snapshotSide(g.faction)
	losses := rules.ResolveBattleTurn(g.battlePlanets, g.battleFleets, g.battleUnits, g.battleRand)
	g.restoreSide(saved)
	kept := losses[:0]
	for _, l := range losses {
		if l.Scorer != cpu {
			kept = append(kept, l)
		}
	}
	return kept
}

// fortressGunFire 是 `rules.FortressGunFire` 加上電腦的要塞不開火。
func (g *Game) fortressGunFire() []rules.FortressGunHit {
	cpu, passive := g.passiveComputer()
	if !passive {
		hits, scenes := rules.FortressGunWithAfterHit(g.battlePlanets, g.battleFleets, g.battleUnits, g.res.ShipKinds, g.battleRand, g.transferAfterGun)
		g.battleGunScenes = scenes
		return hits
	}
	states := map[int]int{}
	for i := range g.battlePlanets {
		if p := &g.battlePlanets[i]; p.Fortress && p.Owner == cpu {
			states[i] = p.GunState
			p.GunState = rules.FortressGunIdle
		}
	}
	hits, scenes := rules.FortressGunWithAfterHit(g.battlePlanets, g.battleFleets, g.battleUnits, g.res.ShipKinds, g.battleRand, g.transferAfterGun)
	g.battleGunScenes = scenes
	for i, st := range states {
		g.battlePlanets[i].GunState = st
	}
	return hits
}

// planetAssault 是 `rules.PlanetAssault` 加上電腦艦隊不打行星、行星的反擊不傷玩家。
func (g *Game) planetAssault() []rules.PlanetAssaultHit {
	cpu, passive := g.passiveComputer()
	if !passive {
		hits, scenes := rules.PlanetAssaultWithAfterCounter(g.battlePlanets, g.battleFleets, g.battleUnits, g.res.ShipKinds, g.battleRand, g.transferAfterCounter)
		g.battleAssaultScenes = scenes
		return hits
	}
	actions := map[int]int{}
	for i := range g.battleFleets {
		if f := &g.battleFleets[i]; f.Faction == cpu {
			actions[i] = f.Action
			f.Action = 0
		}
	}
	saved := g.snapshotSide(g.faction)
	hits, scenes := rules.PlanetAssaultWithAfterCounter(g.battlePlanets, g.battleFleets, g.battleUnits, g.res.ShipKinds, g.battleRand, func(scene rules.PlanetAssaultScene) {
		if scene.Before.Fleet.Faction != g.faction {
			g.transferAfterCounter(scene)
		}
	})
	g.restoreSide(saved)
	for i, a := range actions {
		g.battleFleets[i].Action = a
	}
	// The passive policy restores player units. Its visible snapshots must do
	// the same instead of presenting damage that the player never received.
	for i := range scenes {
		for _, s := range []*rules.PlanetAssaultSnapshot{&scenes[i].Before, &scenes[i].After} {
			for j, u := range saved {
				if j < len(s.Units) {
					s.Units[j] = u
				}
			}
			for fi, f := range g.battleFleets {
				if f.Index == s.Fleet.Index {
					if a, ok := actions[fi]; ok {
						s.Fleet.Action = a
					}
					break
				}
			}
			s.Attack = rules.FleetAttack(&s.Fleet, s.Units)
			s.Integrity = rules.FleetIntegrity(&s.Fleet, s.Units)
		}
		if scenes[i].Kind == rules.PlanetCounterScene && scenes[i].Before.Fleet.Faction == g.faction {
			scenes[i].Destroyed = 0
		}
	}
	g.battleAssaultScenes = scenes
	return hits
}
