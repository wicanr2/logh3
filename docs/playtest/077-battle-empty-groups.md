# 077：戰後清空分隊驗證

2026-10-05，第88輪，遠端#79。已依[109 READY](../spec/109-battle-empty-groups.md)修正正式8-1。
原始方法與分級見[研究422](../re/422-battle-empty-groups.md)，
唯一原版來源索引為[清理資料](../internals/data/battle-cleanup.json)。
原版索引SHA-256 `adc816af7e25a0eba94aed714721f3cb13a330592043581535521e9e969192af`。
143份[產品來源](../internals/data/battle-cleanup-game.json) SHA-256
`828115d7b77055b2bb478732fc575b2926fe833a8248a6ce0f2629563facd566`。
父提交1d15aaafc6f0d24b6742d907575540d4cb7e4305；建置image為golang:1.24-bookworm及logh3-ebiten:2。
原版執行器固定f7523f9、乾淨，沒有改時序、輸入或顯示語意。

## 同狀態局部資料

| 驗收 | 結果 | 範圍 |
|---|---|---|
| 原版雙跑 | 35案6534142 bytes相同，77組作廢 | 明示非旗艦SHIP+08=0，固定seed12345678，停8-2 |
| 純規則及正式Game | 35案完整六表、有效下回合補給池序、Groups與存活命令均相同 | 清理提交四byte引用；SHIP+03與+05及未用尾槽保持 |
| 原子性 | 五個Game、四個規則非法案不提交 | 越界／重複／錯誤擁有者／缺分隊／空旗艦等 |
| 保存與重讀 | 七案每案六檔逐byte符合原版 | 結構化資源刻意保留舊值，保存不復活死亡組、不重算速度／費用 |
| 清理到升級 | 四案雙跑1113851 bytes相同；正式升級入口、末端五表與seed相同 | 明示一次足額功績，組數7／1／6／1，每案一人升級 |

```sh
bash tools/gui.sh test ./internal/rules -run '^TestPruneEmptyBattleGroupsOriginal$' -count=1
bash tools/gui.sh test ./internal/game -run '^TestBattleCleanup' -count=1 -v
bash tools/gui.sh test ./internal/... ./cmd/... -count=1
```

35案原版來源SHA-256 `80f108f6db41137c245dd52040616234779cf93349fbd7c0fae87d1aa2ef9949`；
四案鏈來源SHA-256 `0e45c28c9cfe015a6db29aa8d6a37a9e51a2dc559a981cf0f3b9a317e822b3db`。
完整原始注入、每次呼叫的原bytes、前後表與工具版本保留在凍結來源。
Game保留活的命令、XY、目的地、額度、庫存與偵察隊；只以映射更新一般組Squadron。
保存與LoadSlot只驗資料重讀，LoadSlot仍返回命令階段，不宣稱戰場續戰。

## 正常玩家抽樣

既有battle_promotion_player.go由bootstrap／主選單／138／S3帝國到首戰，
公開戰術選單實行一次，再由存進度選單保存。沒有注入功績、艘數、位置或seed。
兩次14份完整圖／六檔保存／索引／輸入紀錄相同，WORK+06為1、戰略回合380。
此抽樣驗清理接線不阻塞正常回合，沒有自然死亡或升級事件。

```sh
bash tools/gui.sh run tools/scripts/battle_promotion_player.go -out workplace/cleanup79-r88-normal-a
bash tools/gui.sh run tools/scripts/battle_promotion_player.go -out workplace/cleanup79-r88-normal-b
bash tools/gui.sh xrun bash -c 'export LOGH3_LEFT_HOLD_GUI_PREFIX=/src/workplace/cleanup79-r88-normal-gui; bash tools/left_hold_gui.sh'
```

真實X11鍵鼠另抽S3同盟首次戰場資料窗，一倍／兩倍各307200邏輯像素符合既有正常Game／原版基準。
僅保留首戰輸入與清單保持的回歸範圍，沒有把資料窗圖當成死亡後戰場對拍。
標準18產品套件13 PASS／5無測試，最終記錄cleanup79-r88-product-standard-final.log。

## 失敗與限制

修正前28個死亡案有原池組數／擁有者差異。七個無死亡案的行星比較曾因header範圍寫錯而假失敗。
驗證工具首次另把池清單起點寫成DS:9894+36；由實際bytes核為+38後重跑通過。
誤用go test ./...納入tools與workplace，已停止該容器，恢復它改動的go.mod；正式範圍另乾淨重跑。
正式18套件初跑只有舊移動fixture失敗，原因為只放兩個分隊但原表多組；明示一組上界後同命令通過。
保存環境變數第一次未進gui.sh容器，改在xrun內export，七案六檔收據補齊；產品比較本身此前已通過。

109與108保持READY。正常自然死亡與升級GUI、8-5原版結束判定、完整062同態仍缺，#79保持OPEN。
其他未核准差異、CPU被動與音訊實體裝置／人耳門檻保持。
工作根root-owned及.md目錄零筆，UID/GID1000，所有本輪容器已退出移除，未重建image。
