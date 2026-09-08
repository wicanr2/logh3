package rules

import (
	"encoding/json"
	"fmt"
	"os"
)

// 艦種屬性表（`ds:4E3A`，`docs/internals/40`）：一個陣營一筆 208 bytes，
// 內含 8 個艦種各 26 bytes ＝ 13 個 word。

// ShipKind 是一個艦種在表裡的欄位。
type ShipKind struct {
	Kind      int `json:"kind"`
	FullShips int `json:"full_ships"`
	// SemiMajor／SemiMinor 是攻擊橢圓的長短半軸（`+0x16`／`+0x18`，
	// `docs/re/94`）。
	SemiMajor int `json:"semi_major"`
	SemiMinor int `json:"semi_minor"`

	// Words 是這一列原始的 13 個 word。畫面上讀的是**位元組**
	// （`docs/re/112`），所以下面幾個取值函式都從這裡拆。
	Words []int `json:"words"`
}

// Byte 回這一列第 off 個位元組（13 個 word ＝ 26 個位元組）。
func (k ShipKind) Byte(off int) int {
	i := off / 2
	if i < 0 || i >= len(k.Words) {
		return 0
	}
	if off%2 == 0 {
		return k.Words[i] & 0xFF
	}
	return k.Words[i] >> 8
}

// 艦隊詳細資料那一層讀的欄位（`docs/re/112`）。
const (
	kindTopSpeed    = 0x06 // 最高速度
	kindBattleSpeed = 0x07 // 戰鬥速度
	kindBeamName    = 0x0A // 第一組武器的名字索引（0xFF ＝ 沒搭載）
	kindBeamPer     = 0x0B // 第一組武器每艘幾門
	kindMissileName = 0x0E // 第二組武器的名字索引
	kindLauncherPer = 0x0F // 發射裝置每艘幾座
	kindMissilePer  = 0x10 // 第二組武器每艘幾發
	kindFighterName = 0x12 // 艦載機的名字索引
	kindFighterPer  = 0x13 // 艦載機每艘幾架
	kindNoWeapon    = 0xFF
)

// TopSpeed／BattleSpeed 是明細畫面上那兩格。
func (k ShipKind) TopSpeed() int { return k.Byte(kindTopSpeed) }

// BattleSpeed 見 TopSpeed。
func (k ShipKind) BattleSpeed() int { return k.Byte(kindBattleSpeed) }

// Durability 是明細畫面上的「耐久力」＝ 兩個半軸的平均
// （`0x1524F`：`mov ax,[bx+16h] / add ax,[bx+18h] / shr ax,1`）。
func (k ShipKind) Durability() int { return (k.SemiMajor + k.SemiMinor) / 2 }

// Weapon 回一組裝備的「名字索引、每艘幾件」；索引 −1 ＝ 沒搭載。
func (k ShipKind) Weapon(nameOff, perOff int) (name, per int) {
	n := k.Byte(nameOff)
	if n == kindNoWeapon {
		return -1, 0
	}
	return n, k.Byte(perOff)
}

// BeamWeapon／MissileWeapon／LauncherPer／FighterWeapon 是那四組。
func (k ShipKind) BeamWeapon() (int, int) { return k.Weapon(kindBeamName, kindBeamPer) }

// MissileWeapon 見 BeamWeapon。
func (k ShipKind) MissileWeapon() (int, int) { return k.Weapon(kindMissileName, kindMissilePer) }

// LauncherPer 回發射裝置每艘幾座；搭載與否看的是飛彈那一格。
func (k ShipKind) LauncherPer() (int, int) { return k.Weapon(kindMissileName, kindLauncherPer) }

// FighterWeapon 見 BeamWeapon。
func (k ShipKind) FighterWeapon() (int, int) { return k.Weapon(kindFighterName, kindFighterPer) }

// ShipKindTable 是整張表。
type ShipKindTable struct {
	Factions []struct {
		Faction int        `json:"faction"`
		Kinds   []ShipKind `json:"kinds"`
	} `json:"factions"`
}

// LoadShipKindTable 讀 `docs/internals/data/ship-kind-table.json`。
func LoadShipKindTable(path string) (*ShipKindTable, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t ShipKindTable
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, fmt.Errorf("%s：%w", path, err)
	}
	if len(t.Factions) == 0 {
		return nil, fmt.Errorf("%s 裡沒有陣營", path)
	}
	return &t, nil
}

// Kind 回某個陣營某個艦種的欄位。
func (t *ShipKindTable) Kind(faction, kind int) (ShipKind, bool) {
	if t == nil {
		return ShipKind{}, false
	}
	for _, f := range t.Factions {
		if f.Faction != faction {
			continue
		}
		for _, k := range f.Kinds {
			if k.Kind == kind {
				return k, true
			}
		}
	}
	return ShipKind{}, false
}

// SkirmishAttack 回單擊算式裡的攻擊係數 `攻擊方 +0x16 ÷ K`
// （`2350:0023`，`docs/re/86`、`docs/re/94`）。
//
// ⚠ **`faction`／`kind`／`ability` 是「目標」的**，不是攻擊方的：K 是
// 目標的方向性防禦（橢圓），只有 `attack`（艦艇記錄的 `+0x16`）取自
// 攻擊方。`dir` 是方向索引。表裡查不到那個艦種就回 0。
func (t *ShipKindTable) SkirmishAttack(faction, kind, dir, ability, attack int) int {
	k, ok := t.Kind(faction, kind)
	if !ok {
		return 0
	}
	den := AttackCoefficient(k.SemiMajor, k.SemiMinor, dir, ability)
	if den <= 0 {
		return 0
	}
	return attack / den
}
