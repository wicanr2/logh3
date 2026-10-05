# 第107輪：正常行星換手與星域報告

日期：2026-10-05。父提交 `9b5c4b451022286c30b9731c17dba4cf71fbe257`。
沿 [#79](https://github.com/wicanr2/logh3/issues/79)、019／078／106 READY。
本輪補正常 remake 玩家路徑，不新增原版完整八步同狀態收據。106／110 保持 READY，062 保持 DRAFT。
計畫在[本輪入口](../goal/2026-10-05-battle-occupation-player-round-107.md)，身份與完整命令在[來源索引](../internals/data/battle-occupation-player.json)。

## 已通過的玩家鏈

正式 bootstrap 載入原版資料及片頭，公開主選單、防拷 138、取進度 S3、選同盟，正常送出戰略回合進入星域 26。
玩家透過右鍵捲動、旗艦選單、巡航與戰鬥選點，前往敵方行星。首次戰術實行前只固定種子 `0x12345678` 一次。
沒有注入位置、艘數、士氣、功績、歸屬或鏡頭；單人戰術電腦沿 061 保持被動。

| 比較點 | 實際結果 |
|---|---|
| 第六回合，第⑥步 | 艦隊 16／17／28 對局部行星 0 開火，損傷防衛力 558／279／943；艦隊 28 換手成功 |
| 行星身份 | 全域 63，雷奧米爾，位置 1320／812；擁有者 0 → 1 |
| 換手後存量 | 船艦存量全零、飛彈 40000 → 20000、艦載機清零 |
| 8-2 報告等待 | 星域 26 的三方行星數已是 0／3／0，星域擁有者仍 2；情報參謀頭像 75 |
| 報告返回 | 星域擁有者寫成 1，正常返回第七回合的戰場輸入 |
| 公開保存 | 戰術選單「存進度」保存第 0 格，再返回命令畫面；六表共 44560 bytes 等於診斷保存 |
| 公開讀回 | 系統指令「取進度」、第 0 格、同盟，正常回到命令畫面；五表全同，STELLAR 符合既有完整載入重算 |

第六回合換手的影響已到 typed 行星、星域報告、正式 UI 與保存。
報告等待時 PLANET 資源快取仍是舊擁有者；正式戰場以 battlePlanets 為準，進入保存選單前才同步。
這符合 106 的保存接線，不能據此聲稱等待時原版與 remake 的完整六表相同。

STELLAR 讀回的四個實際差異位移是十進位 3319／3401／3527／3653。
獨立重播以既有未安定度及艦隊重數契約計算完整 4032 bytes，再核整檔；沒有只忽略這四個 byte。
LoadSlot 回命令階段，不宣稱戰場續戰。

## 獨立重播與正常 GUI

[探針](../../tools/scripts/battle_occupation_player_probe.py)產生固定公開輸入，正常 Game 使用出貨劇本與正式 renderer。
[重播器](../../tools/scripts/battle_occupation_player_replay.py)改由正式 bootstrap 開始，逐 Tick 呼叫正式滑鼠更新與繪圖。
a／b 先通過；c／d 另補完整 STELLAR 載入重算的核對。四批輸出各 32 個檔案全部相同。
獨立重播的六表 44560 bytes 也與探針保存相同。

正式 `cmd/gin3` 預設 60 Hz、預設片頭、未用 `-script` 或直接畫面入口。
唯讀觀測層記錄狀態、每次 Update 與快照；唯一局面變更是已明示的測試種子。
GUI 用真實 X11 鍵鼠完成 472 項操作，正式 c 批額外 23 次鏡頭回授完整記錄。
報告、換手後戰場、公開保存、公開讀回四張原生 640×480 全圖，共 1228800 pixels，Game 色號與實際 X11 都零差。
公開六表符合獨立重播，讀回後診斷重存也逐檔符合完整載入結果。GUI 來源的遊戲及 app 觀測檔與重播來源相同。
此處的全圖一致是 remake 重播與實際 GUI 一致，沒有原版全畫面比較。

第一批 GUI a 的分類文字沿用舊交戰描述，實際 `natural_finish`／`natural_annihilation` 都是 false。
修正文案後以同一 image、同一命令乾淨重跑 GUI c。正式結論採 c；a 保留作輔助，不把兩批像素相加。

標準命令 `bash tools/gui.sh test ./internal/... ./cmd/... -count=1`：18 套件，13 通過、5 無測試。
Game 115.435 秒，gin3 59.560 秒。正式 Go、資料格式與 AI 沒有修改。

## 失敗探針與限制

所有探索固定同一種子，未重擲；不同劇本、陣營與公開武器選擇均完整記錄。
S1 同盟沒有走到勝後占領；S2 同盟的自然勝利發生在第八或第九回合，沒有足夠回合完成回復後攻占。
原版 110 的停止線保持，不調整正式回合數、回復步數或電腦攻擊行為。
S1 帝國有自然勝利，但沒有勝後占領報告。這些失敗不能推成遊戲不可玩。

早期 S2 探針停在正常 StateRegionFlipReport，修正公開戰後報告處理；後續找不到下一戰時停在 StateCommandPhase。
兩者是測試玩家的假設不足，原批次與失敗日誌保留。
第一個外層產生器與內層產生器撞名的失敗也保留，後續分開檔名。

本輪通過的是第⑥步攻打行星與 8-2 星域換手。
110 的勝後占領 GUI、非旗艦自然死亡／空組、自然升級及原版完整整回合同態仍缺。
登陸艦演出的原版 callback、畫面與音訊沒有新增驗收，#48 不因換手成功關閉。
沒有更改四項使用者裁決，MIDI 與 #89 裁決、人耳、完整曲目循環、macOS 及正式發行交付保持未完成。

## 重生

所有命令在儲存庫根目錄執行，包裝器內是 Docker、network none、UID/GID 1000、原版輸入唯讀。
revision 拒絕覆寫已有輸出，重新驗證時使用新的英數 revision。

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_occupation_player_probe.py --scenario S3 --revision s3planetb --weapon missile --objective planet
bash tools/gui.sh xrun env LOGH3_FLEET_CHAIN_OUTPUT=/src/workplace/occupy79-r107-s3planetb go test -overlay=workplace/occupy79-r107-s3planetb-overlay.json ./internal/game -run '^TestBattleFinishNormalProbe$' -count=1 -v
bash tools/gui.sh xrun python3 tools/scripts/battle_occupation_player_replay.py c
bash tools/gui.sh xrun python3 tools/scripts/battle_occupation_player_replay.py d
bash tools/gui.sh xrun python3 tools/scripts/battle_occupation_player_replay.py c --gui
bash tools/gui.sh xrun python3 tools/scripts/battle_occupation_player_receipt.py --check
```

原版 bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`，
GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
本輪沒有新 dosgolem 執行；原版證據沿 019／078／106 已保存入口，不能把 GUI 收據登錄成原版 oracle。
封存索引包含本輪輸入、輸出、日誌、觀測 Go、二進位、目前正式 Go、工具及原始資料身份。
收尾核 UID/GID、root-owned／`.md` 目錄與專案容器；不做全域清理。
