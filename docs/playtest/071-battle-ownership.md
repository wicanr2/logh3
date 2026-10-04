# 071：戰術星域歸屬與保存

2026-10-04，第82輪。正式父提交58291ecad91f4efbfdfce26a3988707200044e51。
沿[#79](https://github.com/wicanr2/logh3/issues/79)，只處理062子步8-2。
[規格106](../spec/106-battle-region-ownership.md)保持READY，完整062保持DRAFT。
[研究416](../re/416-battle-ownership.md)保存原版控制流、注入與追加勘誤。

## 實作與原版證據

正式8-1完成後重算三方行星數。新舊星域歸屬不同時依帝國、同盟順序通知人類座位。
報告使用原版四句Big5格式及情報參謀頭像。每次點擊只結束目前一則；
兩座位通知都返回後才寫新歸屬、續接回合尾巴。
人類自己拿下費沙時，先登記待辦再顯示自己的報告。

原版來源為GIN3.COM，SHA-256
770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918。
專用dosgolem固定f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，工作樹乾淨、未修改。
[原版v2](../internals/data/battle-ownership-v2.json)凍結244份來源，SHA-256
306f7612cb536cdb441bba2e245baecb9786fbc84a9210b066b06a09934ac262。
156案雙跑相同，產品逐案比較完整126-byte星域、104則報告、費沙時點、WORK及seed。
14組自然呼叫的RAM、framebuffer、色盤及PNG雙跑相同；保存解碼全部70欄位相同。
原始.state因gob列舉順序可有不同雜湊，不宣稱原檔逐byte相同。

## 畫面比較

六張受控報告包含四張完整首報與兩張第二座位報告。
原版首報complete固定到1189100000步，已命中頭像及訊息等待。
早期1188500000步的四張中間幀保留，不用作完成畫面。
原版完整PNGs、framebuffer及16色有效RGB保持，不裁切來源、不遮罩頭像或游標。

| 比較範圍 | 每張像素 | 不同像素 | 最大通道差 | 差異邊界 |
|---|---:|---:|---:|---|
| 報告視窗32,180,416,140 | 58240 | 0 | 0 | null |
| 完整640×480診斷底圖 | 307200 | 248681 | 見各張統計 | 見各張統計 |

矩形由正式TurnWindow與FrameTileW／FrameTileH／FrameCellOriginY計算。
remake在空白畫布上用正式報告renderer重畫；沒有抄原版背景像素。
局部標為same-state controlled report window。完整底圖差異明列為診斷背景不同，
不宣稱全戰場同態。逐張remake、全圖差分及統計保存在產品來源索引。

初稿無頭像推論被畫面差分推翻。106退回DRAFT，補3826:0405參數資料流與完整首報，
再審查為READY。原版第二參數為陣營，頭像取情報參謀13／58；第三參數0在此不消費。
v1保留為已推翻來源，測試只綁v2。

## 正式回合與保存

兩案由公開Game輸入到S3首戰後，明示注入三顆行星歸屬0／1。
正式finishBattleTurn走8-1、8-2及一則人類報告；等待期間Owner與回合不提前更新。
關閉後只續接一次，seed12345678保持，完整星域只有+12及+16–18改變。

隨後以公開Click開戰場選單，選「存進度」，點第一個存檔列。
保存時為StateSaveGame。修正前兩案分別保存行星64／63的舊歸屬；
原版直接改行星記錄，remake的typed副本只在EndBattle寫回，造成星域與行星分家。
正式SaveSlot的戰場分支及進入保存選單前均補同步。
修正後兩案六份DT0重讀，完整星域表與當前行星歸屬相同，沿用既有格式。

這是明示歸屬注入的Game選單診斷，沒有冒稱玩家實際完成占領或X11存檔選單。
LoadSlot仍返回命令階段；[088](../spec/088-normal-battle-player-entry.md)列出的戰場讀檔續戰未知保持。

## 正常GUI回歸與重生

正式cmd/gin3用真實X11鍵鼠經主選單、S3、陣營選擇、系統指令及實行到首戰。
一倍、兩倍皆符合本輪獨立Game基準的完整307200-byte邏輯畫面。
本批使用-intro=false，沒有-script、直接進戰場或局面注入。
視窗尺寸640×480／1280×960已核，沒有擷取全部原生X11像素，
也沒有從本次正常入口宣稱占領GUI或原版同態通過。

第一批拿第47輪舊基準比較失敗；該圖早於後續顯示修正。
失敗輸出保留，按目前程式重生Game基準後兩倍率通過，產品未遷就舊圖。
GUI映像logh3-ebiten:2，ID為
sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8。

```sh
bash tools/gui.sh xrun env LOGH3_OWNERSHIP_OUTPUT=/src/workplace/<新批次> \
  go test ./internal/game -run '^TestBattleOwnership' -count=1 -v
bash tools/gui.sh xrun go run tools/scripts/battle_ownership_player_entry.go \
  -out workplace/<新Game基準>
bash tools/gui.sh xrun env LOGH3_TURN_GUI_PREFIX=<新GUI批次> \
  LOGH3_TURN_PLAYER_DIR=<新Game基準> bash tools/turn_counter_gui.sh
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
```

本次實際批次為ownership79-r82-ui-v8、ownership79-r82-player-entry-v2及round82-normal-gui-v3。
上方命令的LOGH3_OWNERSHIP_OUTPUT、-out、LOGH3_TURN_PLAYER_DIR及LOGH3_TURN_GUI_PREFIX
分別代入這些名稱；其餘參數與命令保持。[產品凍結器](../../tools/scripts/battle_ownership_game_receipt.py)核對各批次。
[產品來源索引](../internals/data/battle-ownership-game.json)記錄程式、輸入、log、PNG與保存雜湊。
索引共307份來源，SHA-256
4ea853fd7d5fa0e97d19745ae21b38ba19ab6148141fe8d79d737be02fc3e3e3。
原版重生入口為tools/battle_ownership_oracle.sh；全部旗標、注入及dosgolem版本由原版v2回查。
兩側測試seed執行前固定；正常首戰GUI沒有固定或比較RNG。

## 驗收限制

156案、六張報告矩形與兩案選單保存均零失敗、零跳過。
標準18套件13通過、5無測試。正常GUI兩倍率通過。
8-3–8-5、正常玩家占領GUI、完整八步及原版整段同態仍缺，#79保持OPEN。
061 CPU被動、核准零距離及#89一像素待裁決保持；本輪未改音訊。
原始遊戲、保存dump及PNG均留本機，沒有加入發行包。
