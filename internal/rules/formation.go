package rules

import (
	"encoding/json"
	"os"
)

// 戰場的隊形（`docs/re/126`、規格 `docs/spec/025`）。
//
// 分隊沒有自己的命令時（艦艇 `+0x1E ＝ 0x7F`）或正在歸隊（3），每回合的
// 目的地就是**艦隊位置 ＋ 槽位偏移依艦隊朝向旋轉**（原版 `2BE1:0059`）。
// 艦艇記錄 `+0x05` 就是槽索引。

// FormationSlots 是槽位偏移表（原版 `ds:59BA`）。
type FormationSlots struct {
	Count  int `json:"count"`
	Centre int `json:"centre"`
	Slots  []struct {
		Index, X, Y int
	} `json:"slots"`
}

// LoadFormationSlots 讀槽位表（`docs/internals/data/battle-formation-slots.json`）。
func LoadFormationSlots(path string) (*FormationSlots, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s FormationSlots
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Offset 回第 i 個槽位的偏移；越界回 (0,0)（＝ 中心那一格）。
func (s *FormationSlots) Offset(i int) (int, int) {
	if s == nil || i < 0 || i >= len(s.Slots) {
		return 0, 0
	}
	return s.Slots[i].X, s.Slots[i].Y
}

// Station 回第 i 個槽位在艦隊位置 (fx,fy)、朝向 a 時的世界座標。
//
// 夾進 0–2047（原版 `2BE1:0059` 尾巴那四段比較，範圍與走一步那支
// （0–2031／0–2023）不同——捨去與夾法都是逐呼叫端的）。
func (s *FormationSlots) Station(fx, fy, a, i int) (int, int) {
	ox, oy := s.Offset(i)
	dx, dy := Rotate256(ox, oy, a)
	return clampInt(fx+dx, 0, formationMax), clampInt(fy+dy, 0, formationMax)
}

// formationMax 是 `2BE1:0059` 夾的上限（0x7FF）。
const formationMax = 0x7FF

// Rotate256 把偏移 (x,y) 依方位 a（一整圈 256）旋轉。
//
// 原版乘完之後**加 512 再算術右移 10 位**，與走一步那支同一種捨去
// （`docs/re/123`）。收據 `s3fltmove` 的九個分隊用這個算法 9／9 完全相同；
// 換成截斷或 floor 會有五、六個差一個像素（`docs/re/126`）。
func Rotate256(x, y, a int) (int, int) {
	c, s := cos256(a), sin256(a)
	return (x*c - y*s + 512) >> 10, (x*s + y*c + 512) >> 10
}
