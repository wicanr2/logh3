# 100：勝後占領探索與目前版本的正常勝利回歸

2026-10-06，沿 [#79](https://github.com/wicanr2/logh3/issues/79)，父提交
`d84dd871f41b13cb777ae7fc3bc38a0e80f354c5`。
命中正常玩家試玩及重製驗證路由，沿既有 [110 READY](../spec/110-battle-finish.md)。
本輪沒有修改正式 Go、AI、存檔格式或 dosgolem，沒有新增原版執行收據。

勝後占領仍未驗收。這批探索沒有產生勝方占領報告。現行 S3 正常勝利、
全滅報告及公開存讀的正式 60 Hz 回歸通過，不能用它代替占領驗收。
完整來源、失敗輸出及分級見[唯一索引](../internals/data/battle-winner-player-search.json)。

## 探索結果

所有探針由公開滑鼠命令操作出貨存檔，首次戰術實行前固定種子 `12345678` 一次。
沒有注入艦隊、行星、命令、艘數或鏡頭。改武器、目標優先序及首戰巡航均屬測試玩家選擇。

| 修訂 | 輸入 | 實際結果 | 分類 |
|---|---|---|---|
| reservea | S3，31不下命令，其餘巡航接近後戰鬥 | 第九回合勝利，零占領；後續月份無戰場而停止 | 正常 Game 助手探索 |
| s1a | S1，既有艦載機優先 | 197次艦隊命令、13擊，未勝利或占領 | 正常 Game 助手探索 |
| s3beam | S3，光束優先 | 第九回合勝利，零占領 | 正常 Game 助手探索 |
| s2beam | S2，光束優先 | 第八回合勝利，零占領 | 正常 Game 助手探索 |
| s2flag | 原擬S2旗艦優先 | 生成器文字錨點失敗；後續誤跑的是未修改目標選擇的模板 | 工具失敗，不當旗艦優先證據 |
| s2flagb | S2，修正生成器後旗艦優先 | 第八回合勝利，零占領 | 乾淨重跑後的正常助手探索 |
| s3wait | S3首戰只巡航，保留敵軍 | 九回合後回命令，下一個月份沒有戰場；零勝利及占領 | 正常助手探索 |
| s3pursue | 首戰巡航後由艦隊視窗占領28 | 星域26的占領項停用，回艦隊視窗；沒有追擊提交 | 公開選單限制，非產品缺陷 |
| latea | S3光束與旗艦優先，明確驗第九回合勝利及零占領 | 50次艦隊命令、19擊、戰後保存通過 | 另由正式 bootstrap 及真實 GUI 驗證 |

110已證實：勝方占領沿當前戰術回合開始，先三次防衛回復，以局部回合達9停止。
第八或第九回合進入仍回復一次，但不進占領開火。這批晚回合零占領符合該契約。
S2首戰星域13唯一行星本來就屬同盟；不能拿這個局面要求行星換手。
S3星域26仍有帝國行星，晚回合勝利不能提供所需的正常占領 GUI 收據。

戰後下一個月份回命令畫面是正常結果。探索模板將「立即再有戰場」當作前提，
其停止不算遊戲卡住。051的戰略作戰選單在戰鬥中或非己方星域只開移動，
s3pursue不能繞過這道公開限制。

## 目前版本的完整重播

`latea`以出貨S3同盟正常載入及實行，進星域26。正式 bootstrap 從主選單、
防拷輸入、公開讀檔與相同玩家命令獨立重播兩次，沒有直接戰場入口。

- 每次158份計畫、完整報告、六檔保存及讀回檔逐位元組相同。
- 第九回合先全滅報告，再敵方無法戰鬥報告。兩則報告記錄完整六表、typed分隊及亂數。
- 勝利報告的六表與探針逐位元組相同，末端亂數為 `40AE2404`。
- 公開系統選單保存與完整診斷保存六檔共44560 bytes相同；公開載入後六檔重存也全同。
- 正式60 Hz真實X11共1010項操作、8385次更新，全部TPS60。額外四次鏡頭回授保留。
- 67張640×480完整畫面，共20582400像素，Game色號及真實X11均零差，沒有裁切或遮罩。
- 實際檢視兩張報告及讀回後星圖，文字完整，頭像與報告先後正確。
- 標準18套件13 PASS、5無測試，Game 111.906秒、gin3 56.080秒。

以上是目前 remake 的正常路徑回歸。原版同狀態等級沿原有收據，不由本輪升格。
真實視窗此批未驗音訊裝置、人耳、Windows或macOS。

## 重生入口

所有命令在專案既有 Docker 包裝器執行，使用 `logh3-ebiten:2`、Go1.24.13、
Python3.11、UID/GID1000:1000，網路關閉。映像精確身份見來源索引。
原始出貨檔唯讀，輸出在本機 `workplace/`，不加入發行包。

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_winner_player_probe.py --revision latea --weapon beam --flagships --expect-late-victory
bash tools/gui.sh xrun bash -c 'LOGH3_FLEET_CHAIN_OUTPUT=/src/workplace/winner79-r112-latea go test -overlay=workplace/winner79-r112-latea-overlay.json ./internal/game -run "^TestBattleFinishNormalProbe$" -count=1 -v > workplace/round112-latea.log 2>&1'
bash tools/gui.sh xrun python3 tools/scripts/battle_winner_player_replay.py a
bash tools/gui.sh xrun python3 tools/scripts/battle_winner_player_replay.py b
bash tools/gui.sh xrun python3 tools/scripts/battle_winner_player_replay.py a --gui
bash tools/gui.sh xrun bash -c 'go test ./internal/... ./cmd/... -count=1 > workplace/round112-standard.log 2>&1'
bash tools/gui.sh xrun python3 tools/scripts/battle_winner_player_receipt.py --check
```

生成器拒絕覆寫既有輸出。實際重跑應改用新修訂及輸出目錄，原收據保持。
最後一行只核對本輪凍結來源；新批次應另行登錄，不能覆寫索引冒稱重現。

## 剩餘閘門

#79維持OPEN，110 READY、062 DRAFT。仍缺正常勝後占領、自然升級、非旗艦自然死亡／空組，
及原版完整八步同狀態收據。下一個占領案例必須在原版開火時間內自然勝利，
星域也必須有可換手的行星；晚回合、全己方行星或戰略占領不能代替。
