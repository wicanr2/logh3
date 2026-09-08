package rules

import "testing"

// 收據的記憶體傾印裡 `ds:94DC` 的種子（`docs/re/87`）。
const (
	seedTurn2  = 0x290995CF // s3turn2，第 1,150,000,000 步
	seedFight  = 0x3FA1A51D // s3fight，第 1,400,000,000 步
	fightSteps = 10         // 兩者之間 rand() 被叫了幾次
)

// 從一份收據的種子往前走，要剛好走到另一份收據的種子——乘數、加數、
// 位寬任何一項錯了都走不到。
func TestRandMatchesOracleSeeds(t *testing.T) {
	r := NewRand(seedTurn2)
	r.Advance(fightSteps)
	if r.Seed() != seedFight {
		t.Fatalf("走 %d 步之後種子是 %08X，收據裡是 %08X",
			fightSteps, r.Seed(), seedFight)
	}
	// 少走一步或多走一步都不該對上。
	r = NewRand(seedTurn2)
	r.Advance(fightSteps - 1)
	if r.Seed() == seedFight {
		t.Error("少走一步也對上了，這個檢查沒有鑑別力")
	}
}

// 傳回值是種子的高 16 位元遮掉最高那一位。
func TestRandValueRange(t *testing.T) {
	r := NewRand(1)
	for i := 0; i < 1000; i++ {
		v := r.Next()
		if v < 0 || v > 32767 {
			t.Fatalf("第 %d 個值是 %d，超出 0–32767", i, v)
		}
	}
}

// 同一個種子要走出同一串——決定性是對拍的前提。
func TestRandIsDeterministic(t *testing.T) {
	a, b := NewRand(seedFight), NewRand(seedFight)
	for i := 0; i < 100; i++ {
		if x, y := a.Next(), b.Next(); x != y {
			t.Fatalf("第 %d 個值不同：%d vs %d", i, x, y)
		}
	}
}
