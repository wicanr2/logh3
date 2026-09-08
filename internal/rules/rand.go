package rules

// Rand 是原版的亂數產生器（`3CD5:000C`，`docs/re/87`）。
//
// 標準的線性同餘：
//
//	種子 ＝ 種子 × 0x41C64E6D ＋ 0x3039
//	回傳 (種子 >> 16) & 0x7FFF
//
// 種子是 32 位元，放在 `ds:94DC`。**戰損帶亂數**（`docs/re/86`），
// 所以要重現原版的戰鬥結果就得連這個一起重現。
type Rand struct {
	seed uint32
}

// 原版的乘數與加數。
const (
	randMul  = 0x41C64E6D
	randAdd  = 0x3039
	randMask = 0x7FFF
)

// NewRand 用某個種子開一份（原版的 `srand` 在 `3CD5:003C`：
// 高位清 0、低位放參數）。
func NewRand(seed uint32) *Rand { return &Rand{seed: seed} }

// Seed 回目前的種子，對得上記憶體裡的 `ds:94DC`。
func (r *Rand) Seed() uint32 { return r.seed }

// SetSeed 直接設種子——從收據的記憶體傾印接上原版的狀態時用。
func (r *Rand) SetSeed(s uint32) { r.seed = s }

// Next 走一步，回原版 `rand()` 的傳回值（0–32767）。
func (r *Rand) Next() int {
	r.seed = r.seed*randMul + randAdd
	return int(r.seed>>16) & randMask
}

// Advance 走 n 步，只更新種子。
func (r *Rand) Advance(n int) {
	for i := 0; i < n; i++ {
		r.seed = r.seed*randMul + randAdd
	}
}
