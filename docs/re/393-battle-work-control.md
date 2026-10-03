# 393：正常首戰WORK控制與經費保存

日期2026-10-03。沿#89與[第55輪](../goal/2026-10-03-battle-work-control-round-55.md)。
父提交b25ee53ca5cac21aa06bd8328162cca100b5e68b，原版執行器f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c／clean，沒有修改執行器。
GIN3.COM SHA-256為770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918；
bundle SHA-256為de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
IDA Pro 9.4沿用locked-v1 image，COM prefix 256 bytes加固定原版RAM切段，
原始bytes與IDA實際base 10200逐項核對。下列定位皆原版CS:IP；WORK為DS:8F14起230 bytes，線性5DA44。

## 原版重播與根因

[重播工具](../../tools/battle_work_control_oracle.sh)固定既有S3父狀態、300M seed D3DC167E，
301M／304M實行及405M關過場，435M保存、434999999倒1MiB。
父狀態SHA-256為0155f63e373582b76a3b7ee54eccbef5ae2626481e09a68550b5fddc580146d7。
兩批normal89-r55-work-a／b的RAM、色號、PNG及DAC完整相同。
RAM SHA-256為ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97，與第54輪相同。
完整旗標各批flags.txt保存；輸出路徑a／b不同，不要求flags.txt雜湊相同。

| WORK位移 | 原版首戰 | 舊remake | 已證實來源 |
|---|---|---|---|
| +02 | 1 | 0 | 12AA:001B寫DS:8F16=1；watch #405775525的0020是執行後IP |
| +03 | 26 | 32 | 0374:03FB初始化掃描星域，0476逐格增加，從0掃到26後開戰 |
| +06 word | 0 | FFFF | 12AA:0064..0078將新戰場FF參數轉0，零起算；UI顯示回合1 |
| +26..29 | 192,100 | 0,0 | 1588:0945依座位找首個對應陣營艦隊，以FLEET+5C／+5E計算鏡頭 |
| +2A..2D | 1312,1660 | 0,0 | 同一函式處理座位1；來源FLEET28的1536,1840 |
| +6E dword | 0 | 665 | 297B:1158／115C寫DS:8F82，行軍途中665→407→149→0 |
| +72 dword | 0 | 1133 | 同一寫入端，1133→917→818→556→363→185→0 |

經費四個差異byte的根因是Save保留載入raw值。不是亂數或經費扣款錯誤。
現行Slot已有Upkeep／MoveCost的typed讀取，Slot.Save尚未寫回，Game也未同步重算值。
原版1261:0067先重算士氣，008C呼叫297B:10DE，00A8／00AC保存維持費，之後才印士氣報告。
297B:10DE逐支合計維持費與移動經費，1158／115C保存移動經費。
116C仍走既有每艦隊公式；12B7／12E2保存兩種費用的低word至FLEET+0A／+0C，
12EC／12EF回讀維持費word並signed延伸。不能直接把公式的未截斷整數相加。
297B:12FB的移動經費讀FLEET+0C，1305..130B排除狀態2與5；原本只排除5的註解不完整。
這些資料流與正常逐支寫入均為proven。既有每艦隊公式及資料同步沿035／067，
本輪不由首戰0推導公式或捏造經費常數。

## 座位、回合及停止線

1588:0930依次呼叫0945(0)／0945(1)。0945讀DS:8F25[座位]的陣營，
掃DS:9898參戰艦隊，取首個對應陣營的FLEET座標。
欄=(x-224)/16、列=(y-180)/20，整數除法朝0截斷；1588:04A0限制欄0..100、列0..84。
09F6／0A0B把欄乘16、列乘20，保存到DS:8F3A[座位]。
因此[研究92](92-battle-lasts-nine-turns.md)稱此陣列為「陣營」不夠精確，本輪勘誤為座位。
12AA:0303..0317讀活動座位鏡頭；034C..0373在Input返回後保存目前鏡頭，兩側皆proven。
沒有對應陣營參戰艦隊時，原碼掃到表尾後仍取表值；後續語意unknown，不依垃圾值猜補。

12AA:00E2在完整戰術回合返回後遞增零起算回合；0107只在正常戰鬥結束回傳0時清戰鬥旗標。
存檔等系統命令的非零回傳保留旗標。0374:03E8逐格掃描，所有星域完成時+03為32。
0374掃描前一個byte為FF，不把「前一byte必須retf」硬套成邊界證明；保存原掃描範圍與直接call定位。

本輪限定正常首戰控制與現行Save／ReadWork資料鏈。
讀到完整WORK及控制值不證明讀檔後能重建戰場並續戰；完整088仍DRAFT。
顯示物件指標、FF字鍵與剩餘一像素不在此處猜補。

## 審查與驗收

初始DRAFT：上述14 bytes已有寫入點與初始值，待分級匯出及規格審查才實作。
舊產品的正常兩種AI路徑，完整WORK比較皆失敗於
02／03／06／07／26／28／2A／2B／2C／2D／6E／6F／72／73，
對應workplace/round55-before-work.log；它是原版完整230 bytes比較，沒有遮罩。
後續驗收與來源雜湊補在本節，既有第51至54輪JSON不重生。

READY審查：最後核實116C完整常式，兩種公式先取低word、再signed延伸，
與原版來源及既有公式契約一致。088第十一個子契約轉READY後才進入實作。

## 第55輪最終驗收

[完整收據](../internals/data/battle-work-control.json) SHA-256
`96660f9c7c8cea715b484d39cfe0a586f81ecc574d963b5bb3ce5c6b9b06578f`。
八份IDA匯出共785條指令，原始bytes與RAM逐項相同；11個函式邊界及56項語意保留分級。
51筆變更寫入序列兩跑相同；watch的IP按執行後解讀，函式入口與指令位址另列。
原版兩批RAM、色號、PNG與DAC完整相同，未修改執行器。

正式正常Game兩批各55份輸出相同，相對第54輪只改首戰及四個純移動存檔中的14個WORK bytes。
兩種戰略AI均經正常載入、實行、首戰、SaveSlot與重讀，完整230 bytes與原版相同。
FLEET27、SHIP214剩餘顯示物件指標保留，STELLAR4032／PLANET6642全同。
五張全圖與父輪相同，初始1 pixel、其餘四張0；五張小地圖均0，不新增遮罩。
正常cmd/gin3真實鍵鼠一倍／兩倍均與正常Game entry一致，已檢視摘要；不拿旗艦游標畫面比較entry。
15項WORK回歸、160限定PASS零skip／失敗，以及18標準套件通過。

明示診斷覆蓋座位反轉、缺匹配艦隊、狀態2／5、低word截斷、signed延伸與不足WORK；
它們是指令契約回歸，不宣稱全部原版自然局面已重播。
讀檔續戰、FF字鍵與完整088仍未完成；#89／#79保持OPEN。

重播入口如下，須使用本輪來源及未占用的輸出名稱；收據腳本拒絕覆寫既有JSON。

```bash
bash tools/battle_work_control_oracle.sh normal89-r55-work-a
bash tools/battle_work_control_oracle.sh normal89-r55-work-b
bash tools/gui.sh xrun python3 tools/scripts/battle_normal_probe_seed.py --prefix round55-normal
bash tools/gui.sh xrun go run -overlay workplace/round55-normal-overlay.json tools/scripts/battle_normal_probe.go -input-display -out workplace/round55-pass-a
bash tools/gui.sh xrun go run -overlay workplace/round55-normal-overlay.json tools/scripts/battle_normal_probe.go -input-display -out workplace/round55-pass-b
bash tools/gui.sh xrun env LOGH3_TURN_GUI_PREFIX=round55-gui LOGH3_TURN_PLAYER_DIR=round55-pass-a bash tools/turn_counter_gui.sh
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
bash tools/gui.sh xrun python3 tools/scripts/battle_work_control_receipt.py
```
