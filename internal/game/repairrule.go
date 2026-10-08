package game

// 可選規則：修護艦可被其他修護艦修理（規格125）。預設關閉，與原版相同。
// 偏好是執行期設定，不進存檔；新局與載入後保留。

// SetRepairShipsRepairable 設定規格125的可選規則。
func (g *Game) SetRepairShipsRepairable(on bool) { g.repairShipsRepairable = on }

// RepairShipsRepairable 回目前的偏好，供入口與正常玩家路徑驗收。
func (g *Game) RepairShipsRepairable() bool { return g.repairShipsRepairable }
