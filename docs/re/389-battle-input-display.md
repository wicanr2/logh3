# 389 正常首戰的純移動摘要

工作[#89](https://github.com/wicanr2/logh3/issues/89)，父提交4be3b1b8f37dd1b9d16331cc2aaaafa2285bce87。
第50輪完整圖5853 pixels差異包含右側5780與旗艦圖塊73，尚未證實是點擊或移動。
本輪先取原版反例，再審088的READY子契約。工具入口[第51輪](../goal/2026-10-03-battle-input-display-round-51.md)。

## 原版反例

GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
dosgolem專用副本HEAD f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，乾淨，Go1.24-bookworm。
[重播工具](../../tools/battle_input_display_oracle.sh)固定正常S3首戰435M，再440M只送一次move事件。
正式base／flagship／squad／empty／panel各a、b，完整RAM、色號、PNG及DAC逐byte相同。
兩側seed沿先前固定D3DC167E；此原版Input期間保持B6E4B78A。

| DOS游標 | 畫面游標 | DS:9B7A | 1CC5:008E | 結果 |
|---|---|---|---|---|
| 240,200初始 | 240,240 | 0050 | 入場一次 | 第13艦隊舒坦梅茲 |
| 248,208 | 248,248 | 0050 | 零次 | 同旗艦內摘要保持 |
| 249,184 | 249,224 | 0052 | 一次 | 第2部隊，3000艘 |
| 100,100 | 100,140 | FFFF | 零次 | 恢復Input摘要 |
| 500,250 | 500,290 | FFFF | 零次 | 恢復Input摘要 |

原版FLEET3904／SHIP27456／WORK230 bytes全部未變，純移動沒有12AA:02D4命令入口。
空白對照完整畫面差6033 pixels，其中右側相對初始差5780。這足以否定舊「只在命令選單」實作。
原始entry.bin SHA-256 `ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97`。

## IDA資料流

[匯出工具](../../tools/scripts/battle_input_display_ida.py)使用IDA 9.4 locked-v1、UID1000。
COM前置256 bytes，IDA實測原段基底10200；原版CS:IP與線性位址另列，逐byte回查上述RAM。
函式自動名稱僅供定位，語意來源為原指令與獨立動態實驗。

- proven：12F1:04C3取事件物件，04FA比較DS:9B7A，04FC相同則退出。
  051F原bytes `9a0900c51c`呼叫1CC5:0009，0530寫新的物件編號。
- proven：1CC5:0030比較SHIP陣營與目前活動陣營；0049呼叫008E完整摘要。
  0071呼叫25DB:0AF8。敵方改走04CC，其文字與正常敵方動態尚未驗，不擴張聲明。
- proven：12F1:0539在FFFF時直接返回；07D1清反白物件，07D6將快取設FFFF。
  這條路會重畫右側戰場摘要。
- proven：25DB:0BD2建現在位置與集結點兩個透明物件，圖塊10、寬高16、優先FF／FE。
  0C43／0C4A讀SHIP+34／36；0CCD／0CD4讀非旗艦SHIP+28／2A；
  0CFA／0D03讀旗艦FLEET+64／66。不必翻譯DOS物件管理器或捏造far指標。

切段SHA-256：12F1 `0379c0c9852b419014643aa47d49eec2d54c6ad8946a63f828ed4990556abc95`；
1CC5 `b87bc806227d13b1f3741452bb7b8e37bbaf9b8175e5e3d4f612ae48df21799e`；
25DB `d1004845670387a15986f585658051965a2cb899f1f499d062b7b7ba8bd34b7a`。
重播、完整雜湊及後續實作驗收補入本輪結構化收據。完整088維持DRAFT。

## 工具邊界

save-state排在最後一步時不執行，擷取上限須超過該步；explore舊輸出保留，不當正式收據。
IDA最初沒有自動函式邊界，明示原始序言及前一函式retf後，手動建立窄函式；
所有正式匯出逐byte驗RAM，不以退出碼證明成功。

## 正式限定驗收

依088 READY接純移動快照與兩個透明反白物件，艦隊行動及標籤士氣讀活戰術記錄。
五組原版RAM／全圖／色盤獨立雙跑相同，正式正常玩家55份輸出雙跑相同。
初始首戰差由5853減至2 pixels；旗艦／分隊各1，空白／右面板全圖0。
分隊文字區五案均0／20480，未解像素仍在完整差分，沒有遮罩。
WORK14／FLEET27／SHIP276／STELLAR0／PLANET0與380／799/3/1、七艦隊／55池／B6E4B78A保持。
61限定PASS零skip、18產品套件及正式X11一倍／兩倍正常首戰通過。

[完整收據](../internals/data/battle-input-display.json) SHA-256
`c1f387c1e3729450f4b7b3f9f8babb21b0b8f49d1b0762b1f6e0cca648cfb5c6`。
它保存535條原始指令、22項分級語意、全部原圖與逐區數字。
初始差在(506,93)與(255,255)，分隊差在(506,92)，仍unknown；完整088及#89／#79不升格。
敵方／行星／特殊艦種與模態退出未新增驗收。
