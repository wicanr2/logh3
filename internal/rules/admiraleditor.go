package rules

import "fmt"

// 索引只描述已查明的資料位移，未知原版欄名見internals/08。
var AdmiralEditableOffsets = [13]int{11, 12, 13, 14, 15, 16, 17, 18, 19, 23, 20, 21, 22}

type AdmiralSpecial struct {
	Transfer      int `json:"transfer"`
	Movement      int `json:"movement"`
	Attack        int `json:"attack"`
	FighterAttack int `json:"fighter_attack"`
}

func DefaultAdmiralSpecial(index int) AdmiralSpecial {
	s := AdmiralSpecial{Transfer: 25}
	if index == 5 {
		s.Transfer = 100
	}
	if index == 0 || index == 52 {
		s.Movement, s.Attack = 50, 30
	}
	if index == 10 || index == 74 {
		s.FighterAttack = 30
	}
	return s
}

type AdmiralOverride struct {
	Abilities map[int]int     `json:"abilities,omitempty"`
	Special   *AdmiralSpecial `json:"special,omitempty"`
}

type AdmiralPreferences struct {
	Version  int                     `json:"version"`
	Admirals map[int]AdmiralOverride `json:"admirals,omitempty"`
}

func NewAdmiralPreferences() AdmiralPreferences {
	return AdmiralPreferences{Version: 1, Admirals: map[int]AdmiralOverride{}}
}

func (p AdmiralPreferences) Validate() error {
	if p.Version != 1 {
		return fmt.Errorf("不支援的提督設定版本")
	}
	for i, v := range p.Admirals {
		if i < 0 || i >= 82 {
			return fmt.Errorf("提督索引超出範圍")
		}
		for field, n := range v.Abilities {
			if field < 0 || field >= len(AdmiralEditableOffsets) || n < 1 || n > 200 {
				return fmt.Errorf("提督能力必須為1至200")
			}
		}
		if s := v.Special; s != nil && (s.Transfer < 0 || s.Transfer > 100 || s.Movement < 0 || s.Movement > 200 || s.Attack < 0 || s.Attack > 200 || s.FighterAttack < 0 || s.FighterAttack > 200) {
			return fmt.Errorf("特殊能力百分比超出範圍")
		}
	}
	return nil
}

// 武器能力同時是防禦係數的乘數。全局設定必須對未來任何艦種／角度保持正分母。
func (p AdmiralPreferences) ValidateCombat(kinds *ShipKindTable) error {
	needed := false
	for _, v := range p.Admirals {
		for field := range v.Abilities {
			if field >= 4 && field <= 6 {
				needed = true
			}
		}
	}
	if !needed {
		return nil
	}
	if kinds == nil {
		return fmt.Errorf("武器能力驗證缺艦種表")
	}
	minimum := 1
	for faction := 0; faction < 2; faction++ {
		for kind := 0; kind < 8; kind++ {
			k, ok := kinds.Kind(faction, kind)
			if !ok || k.SemiMajor <= 0 || k.SemiMinor <= 0 || k.SemiMinor > k.SemiMajor {
				return fmt.Errorf("武器能力驗證的艦種資料無效")
			}
			for dir := 0; dir < 256; dir++ {
				radius := EllipseRadius(k.SemiMajor, k.SemiMinor, dir).Int()
				if radius <= 0 {
					return fmt.Errorf("艦種防禦半徑無效")
				}
				minimum = max(minimum, (100+radius-1)/radius)
			}
		}
	}
	for _, v := range p.Admirals {
		for field, n := range v.Abilities {
			if field >= 4 && field <= 6 && n < minimum {
				return fmt.Errorf("武器能力至少%d", minimum)
			}
		}
	}
	return nil
}

func (p AdmiralPreferences) Clone() AdmiralPreferences {
	q := NewAdmiralPreferences()
	for i, v := range p.Admirals {
		x := AdmiralOverride{Abilities: map[int]int{}}
		for f, n := range v.Abilities {
			x.Abilities[f] = n
		}
		if v.Special != nil {
			s := *v.Special
			x.Special = &s
		}
		q.Admirals[i] = x
	}
	return q
}

func (p AdmiralPreferences) Special(index int) AdmiralSpecial {
	if v, ok := p.Admirals[index]; ok && v.Special != nil {
		return *v.Special
	}
	return DefaultAdmiralSpecial(index)
}

func (p AdmiralPreferences) Apply(data []byte) {
	for i, v := range p.Admirals {
		if (i+1)*AdmiralRecord > len(data) {
			continue
		}
		for f, n := range v.Abilities {
			data[i*AdmiralRecord+AdmiralEditableOffsets[f]] = byte(n)
		}
	}
}

// 百分比只改本次係數或新額度，不寫回基本能力。
func BonusAmount(amount, percent int) int {
	if amount <= 0 || percent <= 0 {
		return amount
	}
	n := int64(amount) * int64(100+percent) / 100
	return int(min(n, int64(0x7fffffff)))
}
