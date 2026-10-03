# 399：正常S3首次索敵的限定資料鏈

2026-10-03，第61輪，工作[#79](https://github.com/wicanr2/logh3/issues/79)。
父提交`eb6814ab39daef76f2531e4020f64c15bf50679d`；[087](../spec/087-battle-spotting-phase.md)保持READY。
入口：[工作快照](../goal/2026-10-03-normal-spotting-round-61.md)、
[原版重播](../../tools/normal_spotting_oracle.sh)、[收據](../internals/data/normal-spotting.json)。
收據SHA-256：`4f15269fcedd71292a16a68c5bb6ec76c4303f4b0edc8d1b768a2d82909209a4`。

## 輸入身份

| 輸入／工具 | 固定身份 |
|---|---|
| 原版bundle | `org_game/银河英雄传说3SP.zip`，SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673` |
| GIN3.COM | SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918` |
| 正常S3父狀態 | 專用dosgolem `workplace/fcmd/fcmd-s3-base.state`，SHA-256 `0155f63e373582b76a3b7ee54eccbef5ae2626481e09a68550b5fddc580146d7` |
| dosgolem | `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨，未修改 |
| Go容器 | `golang:1.24-bookworm`，完整映像ID保存在收據 |
| GUI容器 | `logh3-ebiten:2`，完整映像ID保存在收據 |

原版讀取未修改父狀態，經系統、實行、到達報告進正常首戰，再正常實行第一次戰術回合。
策略seed在實行前固定D3DC167E，首戰B6E4B78A；戰術實行前固定12345678。
沒有艦隊、鏡頭、情報或畫面注入，沒有重擲。完整旗標及所有lin注入在收據。
本輪不新增IDA分析；原始定位與bytes沿087、研究379至381及388。
remake測試時Git HEAD為上述父提交，工作樹含本輪未提交的觀測器及文件修改。
實際執行來源SHA、Docker映像ID與輸出雜湊保存於JSON；正常Game命令與環境參數見051。
兩側畫布均原生640×480，沒有裁切或縮放；原版滑鼠Y到產品畫布Y加40，沿既有輸入映射。

## 原版邊界

下表是dosgolem步數與原版CS:IP，不是IDA位址。proven限固定自然局面。

| 邊界 | 原版定位 | 記憶體／畫面擷取步數 |
|---|---|---|
| 索敵入口，尚未寫標籤 | 181B:0262，483778106 | 483778107 |
| 標籤已寫、第一席查詢 | 2526:00F8，483798430 | 483798431 |
| 第二席查詢 | 2526:00F8，484103337 | 484103338 |
| 第四階段入口 | 181B:028C，484357728 | 484357728 |

兩次完整RAM、色號、PNG及DAC相同。每次16份輸出，命令、meta及report另保留。
第三階段兩席都完成，033B／04B1、報告、等待及標亮常式皆零次。
完整SHIP27456、FLEET3904、WORK230、STELLAR4032及PLANET6642 bytes在第三階段前後全部未變。
seed在四個邊界均AFCFD1BC；第三階段不消費亂數。

## 正常remake與限制

[正式bootstrap測試](../../internal/game/normalspotting_player_test.go)由Game.New、主選單138、S3帝國、正常實行到首戰。
[觀測器](../../internal/game/normalspotting_test.go)只預先固定seed，沒有停用061、原版RAM載入或直接進入戰場。
第三階段after只包一個讀取觀測，擷取後照常呼叫既有續接。

- 352筆Seen bit0、Detail與Hold在標籤及結束兩邊界符合原版。
- 正常觀測者共10筆，無報告，僅續接一次；完成後第2戰術回合，完整SHIP／FLEET保存及重讀相同。
- remake自己的第三階段五張完整表也未改變。兩次各24份輸出全部相同。
- 標籤與結束的完整307200色號及有效RGB均零差；Field、右側、底部亦零差，沒有排除區。
- 全資料仍不同：Slot原始WORK32、FLEET147、SHIP699 bytes；STELLAR與PLANET零差。
  WORK是階段中的Slot raw觀測，未經SaveSlot導出；不能把32解讀為正式存檔WORK的缺陷。
- 索敵入口原版seedAFCFD1BC、remake12345678。最早亂數消費者未定位，標unknown。
  正常CPU位置與移動旗標也不同，061保持。不能僅由畫面相同推定完整同狀態。

整組事前定為nearby-state，保持此等級。資料表差異逐byte保存在JSON，沒有補造顯示指標或遮罩。
本輪沒有新增真實cmd/gin3 X11收據；151個受控幀、#89一像素裁決及完整八步門檻保持。
索敵本身沒有新增玩家結果差異，停止追查其亂數內部。下一步沿#79補第④步艦隊戰鬥的觸發與演出證據。
正常發現報告與完整同態保留為原閘門；有玩家功能依賴時再開窄驗證。
不把整段精確骰序列為完成條件，不重新開已驗來源標頭或鏡頭分支。

## 驗證腳本勘誤

初跑錯把SHIP+16全部八位元當Seen，池105的移動bit1因此失敗。
原版首戰bit1為0，索敵入口為1；remake保持0，完整差異仍記錄。
087既有契約只定義bit0 Seen，其餘bits由移動規則維持。改正欄位檢查後同容器重跑通過。
這是新驗證腳本的欄位錯誤，沒有修改原版期望或正式產品。

收據清單計數、DAC6-bit與PNG RGB8-bit單位混用、假定實行前seed等於索敵seed亦已修正。
改用明列輸出集合、正式PNG有效色盤，以及兩側各自的階段seed不變核對。
舊log保存在workplace，未以這些錯誤認定產品缺陷。標準18套件通過。
