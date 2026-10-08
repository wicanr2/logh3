// 此目錄的程式只在 dosgolem 的 module 內編譯（decomp/tools/golem_probe_run.sh 與 fleet_attack_run.sh），
// 獨立的 go.mod 讓主 module 的 `go build ./...` 不會掃到它。
module github.com/wicanr2/logh3-decomp-harness

go 1.24
