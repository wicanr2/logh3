# 381 索敵可見性與圖示快取的最小契約

日期：2026-10-03，第四十一輪，Issue #79。
父提交80af23933b8c31357242bdbd647fe469ae1f67ef。

## 證據身份

- 未修改GIN3.COM：770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918。
- 原版bundle：de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
- dosgolem專用副本f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，工作樹乾淨，未改執行器。
- 常式診斷來源RAM：9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d。
- IDA9.4來源RAM：9095dfa08ab36201784e6d4555b5e42312ac76bb20049614cf53bf93bc09e7ec。
- Go image為golang:1.24-bookworm；IDA image為ida-pro-9.4-idapython:locked-v1。
  Image身份沿用[研究380](380-battle-discovery-presentation.md)，沒有重建工具鏈。

61案完整收據兩次獨立執行相同，SHA-256
f03cc58211076e8bc0aa5a2a2e09414ae84716d68080fa044af62169ada01471。
固定計畫、全部注入、呼叫參數、前後完整SHIP78／FLEET122／顯示物件40 bytes、種子
存於[完整證據](../internals/data/battle-spotting-display.json)。
七份IDA匯出逐指令核對原RAM，26項附加語意由[分級索引](../internals/data/battle-spotting-display-semantics.json)合併。
COM offset0對應IDA 0x10200，原始CS:IP、dosgolem線性位址另列，保留原名及運算元。
未審查helper仍為unknown，不從自動名稱推定用途。

## 已證實的快取寫入

25DB:03E9先判force，再判Detail是否恰為3。兩者之一成立就走完整圖示支線。
Kind7判斷只在其餘支線，不能在整支函式最前面把偵察隊排除。

| 條件 | SHIP+2F | 旗艦的FLEET+6A |
|---|---|---|
| force非零或Detail3，分隊序號0 | 陣營基底加艦隊編號減1 | DirectionIndex(FLEET+4A) |
| force非零或Detail3，其餘序號 | 陣營基底加16加Kind | 不改 |
| 其餘支線Kind7 | 9 | 序號0時先寫9 |
| Detail0 | 9 | 序號0時先寫9 |
| Detail1 | 陣營基底加24 | 序號0時先寫9 |
| Detail2 | 陣營基底加16加Kind | 序號0時先寫9 |
| 其餘Detail | 保留舊圖示 | 序號0時先寫9 |

陣營0基底20，其餘50；輸出以byte保存。36案涵蓋force0／1、Detail0／1／2／3／4／255、
旗艦／一般分隊／Kind7。每案核對完整紀錄，SHIP只允許+2F、FLEET只允許+6A改動。
25DB:04E7的0503／050C直接寫+6A。既有第40輪Detail3的9→6因此得到寫入端證據。

Kind7完整圖示是帝國43、同盟73。這只證明常式分支，force仍由實際caller決定；
不能將診斷force1當成全部偵察隊的正式顯示設定。

## 已證實的界外旗標

25DB:0714把SHIP+34/+36減鏡頭後寫到顯示物件+6/+8，並設dirty為FF。
35D1:026D沿2376:0164進01E6，02B0產生界外值，01E6的024E寫物件+1E。

| 計算 | 原始契約 |
|---|---|
| 左圖塊 | 帶號16-bit的相對X加物件+12，sar4向負無限取整 |
| 上圖塊 | 帶號16-bit的相對Y加物件+14，idiv20向零截斷 |
| 右圖塊 | X加偏移加物件+A寬度減1，再sar4 |
| 下圖塊 | Y加偏移加物件+B高度減1，再idiv20 |
| +1E | 上述範圍與+1A..+1D邊界完全分離為1，其餘為0 |

固定分隊物件寬高16×16、偏移16／20、邊界1／1／28／18。
15案涵蓋四邊相鄰位置。相對X=-16界外、-15可見、447可見、448界外；
相對Y=-16界外、-15可見、359可見、360界外。
這是指定顯示物件的RAM契約，沒有以冷VGA診斷畫面宣稱UI對拍。
其他物件尺寸、任意裁切窗及完整圖塊blitter不列新逆向工作。

## 已證實的情報byte行為

十案沿既有2526:01FE診斷，觀測點及唯一候選距離100。
Kind0／7各驗Detail0／2／3／4／255。Kind7且Out清除時與一般候選相同。
Detail3只設Seen、不報告；其餘以byte加1，2→3設Hold2，4→5及255→0保留Hold。
前後完整observer78不變；candidate只改+16 bit0、+17及必要的+18。
既有UpdateSpotting額外排除Kind7並用Detail小於3夾限，與這些原版分支不同。

## 審查與停止線

### 回收後輸入刷新補證

修正Kind7候選後，正常回收測試發現索敵隊仍有Detail1，輸入刷新把標記重新設C5。
既有「已回收必須無標記」斷言只覆蓋先前沒有偵察隊情報的局面，不能當成通則。
沿原版26E5:00AC caller補兩案：pool81、FLEET12的ScoutMode=FF、SHIP Action=FF，
標記先05，分別Detail0／1，viewer0、8F1C bit0清除、地圖啟用。
完整雙跑SHA-256 d9cb659cc3bffced7dd514d0ba1d3cba507f3d75f4ed4c549657763fa06152f9。
Detail0沒有任何SHIP變動；Detail1只有+19由05→C5及+1A/+1C現在座標寫入。
兩案FLEET完整122 bytes、Out=FF及seed均不變，完整補充收據納入同一資料JSON。
26E5:0188..01C4原始指令沿[研究374](374-battle-marker-lifecycle.md)保留。
未放出仍以force0呼叫圖示／標記；已放出己方偵察隊force1，與一般分隊相同。
不因Out替代082已證實的標記狀態。這是原版行為，未將它改成新的玩家規則。

依[087](../spec/087-battle-spotting-phase.md)將候選／情報byte與03E9圖示子契約升READY後修正。
界外旗標及快取最小鏈已足夠，不繼續深挖通用renderer。
完整第三階段仍須正式逐筆快取、等待、戰鬥前續接與保存，再補正常玩家路徑及整幀差分。
常式診斷與規則回歸不能替代上述驗收，#79保持開啟。

首次IDA索引切段鍵錯誤及probe掛載路徑錯誤均為驗證環境問題。
修正後沿用相同image乾淨重跑。最初46案是探索，正式固定61案另存，不覆寫收據。

## 第42輪正式接線續證

日期2026-10-03，父提交c749176730b4c94b6f235650db1c2a98d34acaaf，工作項目#79。
前節「尚未接線」為第41輪歷史狀態。087完整READY後已接正式Game，指定94個UI邊界通過，整步尚未CONFORMED。
原始檔、工具及位址空間沿379／380／381，未改原版與dosgolem。

WORK+11／+12、+0D／+0E對應DS:8F25座位與8F21 controller。
每席依旗艦、已放出偵察隊、行星觀測；CPU更新情報，跳過UI。
顯示物件+6/+8保存世界位置減最後已畫鏡頭；camera案要求0,0但背景66,30至真正置鏡。
正式Game保留兩者，沒有把原版framebuffer作輸入。

2526:04B1原始bytes 9a73008239呼叫3982:0073，04C2原始bytes 83c302加caller字串buffer偏移。
研究338與既有TurnWindow.FleetNumber已證實一位全形、兩位半形。
初稿兩位全形改正；固定匯出workplace/ida/re79-spot-r39-report.json的SHA-256為
a2fe19073ffdd4e07bfc010e02cdad5b36f9bca00a49e6e70e3988303e015b0b。
CS:IP、dosgolem linear及IDA EA仍依379分列，未用自訂名稱作證據。

Seen與舊名Attack低byte共用SHIP+16。下輪Decay後先同步再讓移動寫回，
+17／+18按typed Detail／Hold保存，不改正式存檔格式。
九案5330筆SHIP78、680筆FLEET122與94組原生圖零差異，proven限記錄的入口與UI邊界。
22個內部停點不稱已驗。命令、來源、雜湊、排除理由及正常抽樣見
[正式收據](../internals/data/battle-spotting-game.json)、[046](../playtest/046-battle-spotting-phase.md)。
艦隊／偵察隊來源原版標頭與近距候選移鏡待補，不用正常remake圖冒充oracle。
