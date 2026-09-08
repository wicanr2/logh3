package rules

// 曆法：一回合三天、一個月十回合、一年十二個月（`docs/re/25`）。
//
// 遊戲只存一個**回合計數器**（`ds:8F18`，開局 0）與紀元年（`ds:8FF6`，796），
// 年月日是印的時候算出來的：
//
//	a1 18 8f     mov ax,[8F18]
//	33 d2 f7 f3  xor dx,dx; div bx     ; bx = 120
//	03 06 f6 8f  add ax,[8FF6]         ; ← 年
//	…（餘數再除 10 得月、除 1 得回合序號 ×3 + 1 得日）
//
// Epoch 是回合計數器的起算年（`ds:8FF6`，796）。
const Epoch = 796

const (
	TurnsPerMonth = 10
	MonthsPerYear = 12
	DaysPerTurn   = 3
	TurnsPerYear  = TurnsPerMonth * MonthsPerYear
)

// Date 把回合計數器換成年月日。epoch 是紀元年（原版是 796）。
func Date(turn, epoch int) (year, month, day int) {
	year = turn/TurnsPerYear + epoch
	rest := turn % TurnsPerYear
	month = rest/TurnsPerMonth + 1
	day = rest%TurnsPerMonth*DaysPerTurn + 1
	return
}
