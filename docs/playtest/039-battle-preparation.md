# 039 — 戰場準備階段蒐證與共用補充回寫驗證

日期：2026-10-02；工作項[#79](https://github.com/wicanr2/logh3/issues/79)。
結果：共用行星補充／戰後資源回寫通過限定範圍；完整自動第一階段及規格062仍 DRAFT。
固定身份、原始位址與雜湊見[研究367](../re/367-battle-preparation.md)，
單一欄位收據與完整原版旗標見[battle-preparation.json](../internals/data/battle-preparation.json)。

## 原版驗證及比較等級

`battle79-preparation-b`／`c` 透過 `tools/oracle.sh` 重生：相同檢查點、完整輸入、
游標、步數、同一乾淨dosgolem提交；第一階段前事先固定種子。
1MiB前後狀態、640×480原生色號與調色盤均可逐位元組重播。
此為**原版重播證據**，沒有重製側同狀態截圖／差分，不登記為畫面對拍通過。

原版常式診斷由 `battle_preparation_probe.go` 產生兩份完整收據，SHA-256皆為
`e922c6b273b6971fc0e746a3f66183d40508758902a70acec6b766e08930f069`。
完整注入位址／bytes留本機，抽出的十六案欄位存JSON；每案種子未消費。
原版與remake的**相同輸入欄位比較**只涵蓋共用行星補充的艘數、修理上限與庫存，
以及由原版補給輸出驅動的戰後保存；不證明完整回合順序、報告或整段亂數序列。

## 分級 IDA 匯出

`tools/ida.sh raw idat -A -L/work/re79-reviewed-<段>-internal.log`
加 `-S` 指定 `battle_rules_ida.py <輸出> <段> --index=/work/re79-reviewed-index.json <查詢>`。
每個輸入皆為新的 `re79-reviewed-<段>.com`；禁止在正式資料庫連續試跑。
切段雜湊及12筆審查索引於研究367；未審查指令保留unknown警示。

| 段 | 完整查詢 | JSON SHA-256 |
|---|---|---|
| 1930 | `019D:117 02B4:F7 03AB:DA 0485:8B 0510:DC 10B2:8C 0944:15E 113E:160` | `4435e9a07ac186a1859ccd6c4f649cc77648c6058867842095878a4267aa8084` |
| 1AAB | `006D:123 025B:F8 0353:EA 043D:80 04BD:C6 0879:82 0D9F:CE` | `dd3fe01debc2104d6b682814941f444ff5517406e18a9a287c8179f4c645a482` |
| 181B | `003F:44 0083:77` | `9623c97da92fe44722089409cc3d6468d25a81acb33657a085375eaeea968e5d` |

查詢數值為十六進位；函式範圍只是導覽，另保留原始retf／控制流，未用範圍名字當語意證據。

## 正式規則及垂直鏈

```sh
tools/gui.sh test -count=1 -json ./internal/rules ./internal/game -run '^(TestPlanetReplenishMatchesPreparationOracle|TestPlanetCommandPlayerPaths|TestPlanetReplenishResourcesPersist|TestPlanetSupplyAndReplenish|TestBattleResultPersists|TestSaveLiveFleetStateLoop)$'
tools/gui.sh test -count=1 ./internal/... ./cmd/...
```

限定回歸16個通過事件、零失敗、零跳過；完整標準命令18個產品套件成功或無測試。
原版五個補充案例直接核對完整相關欄位；另驗艘數不變的彈藥結果仍可回寫。
父提交在隔離暫存樹加入相同收據與測試，三個修理上限案例及彈藥回寫案例失敗，
修正後全部通過；舊樹與現在工作樹分離，不曾暫時退回產品程式。

正式 `Click` 的行星補充、撤退、艦艇池回寫、SaveSlot及正式載入器抽驗通過：
50艘／修理上限75／行星庫存20 → 70艘／95／0，兩格彈藥11／22保留。
這是合成診斷局面；原版欄位證據與正常GUI入口各自記錄。

## 正常 GUI 抽樣

沿用 [`battle_rules_gui_check.py`](../../tools/scripts/battle_rules_gui_check.py)，
`--head 129c1d9c09295ddc822dcf3bc675dd52bfb5b721`
`--dirty '第25輪共用補充／戰後回寫修正尚未提交'`
`--image-id sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8`
`--output /src/workplace/oracle/battle79-routines`，在 `tools/gui.sh xrun` 執行。
腳本所有命令、PNG／色號、狀態與雜湊保存於 `gui-check.json`。

三份GUI由主選單進S3正常戰場、實行一次、獨立新遊戲編隊存讀。
FLEET3904bytes、SHIP27456bytes與行星完整存讀相同；雙方艦隊數1／3。
不宣稱正常GUI已實際操作補充，不宣稱中途戰場存檔或macOS／Windows實機。

| 本機收據（`workplace/oracle/battle79-routines/`） | SHA-256 |
|---|---|
| `remake-before-fix-final.log` | `994ee6c4162301b1b19ed46cb6113d9971ad07b3c01d63a85c8ac24b58bc06bd` |
| `remake-targeted-final.jsonl` | `dd864eff47a60c48f51185f5342875a8513b5346f8c2d66970f262737b7fe1cc` |
| `product-tests.log` | `dba8f1480a288207525061b2034f32deb3df85f83afbea34316b4651b6edd5f4` |
| `gui-check.json` | `4b8592c044e757e2ee807ccd3b1eaea2580e87759982bd52f41f87ef1c2c4eba` |

| 驗收當下來源 | SHA-256 |
|---|---|
| `internal/rules/planetcmd.go` | `a3d300491e1bc3a3f5e21df16fdbcce75779b3fb207781b40c34944268404be6` |
| `internal/rules/resolve.go` | `d224ee581956e34f4cc34c83f15762571306eb888dd0c2007ac41ee7d1a1492a` |
| `internal/game/planetcmd_test.go` | `5801c76d71b2f186fda6dc78594fe40258278725f5e7b1b12a44566e2eb9658d` |
| `internal/rules/battle_preparation_oracle_test.go` | `89c7476812bff8ca68e27cc5f8546ec47ee78a7e92bdabd6115087017409e54e` |

最後清理與遠端狀態記於WORKLOG；#79保持open，完整Goal仍未完成。
