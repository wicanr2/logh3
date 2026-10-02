# 380 第三階段發現標頭、暫時標亮與置鏡

日期：2026-10-02，第四十輪，工作項目 [#79](https://github.com/wicanr2/logh3/issues/79)。
父提交 `bc3e2b1708fdd6c9c700a47d94464b032af59846`。
本輪補原版呈現證據，沒有改正式 Game 的階段順序。

## 輸入與工具

- 未修改 `GIN3.COM`，SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
- 父狀態 `battle79-move-entry-a/machine.state`，SHA-256
  `39ed892ae729c2b2bd1a0b208439d106e4913867e1b50134396fabd704b39b55`，來源鏈見[第29輪](../goal/2026-10-02-battle-movement-round-29.md)。
- dosgolem 專用副本 HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`。
  每次正式擷取前核對完整 HEAD 與 clean；沒有改執行器。
- Go image ID `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`。
- IDA Pro 9.4 image ID `sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`。
  3826／25DB／1588 切段來自第36輪入口 RAM，SHA-256
  `9095dfa08ab36201784e6d4555b5e42312ac76bb20049614cf53bf93bc09e7ec`。
  3765 切段來自第34輪入口 RAM，SHA-256
  `9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d`。
- 切段前置256個零 bytes，COM offset 0對應 IDA `0x10200`。
  原始 CS:IP、dosgolem線性位址、IDA位址分列；DS線性基底 `0x54B30`。
  [指令索引](../internals/data/battle-discovery-semantics.json)保留原始 bytes，
  [完整證據](../internals/data/battle-discovery-evidence.json)保留八份匯出及輸入雜湊。

父狀態已有固定 seed `12345678`。七個分支事先固定注入／輸入，tick=20000，沒有重擲或改原版指令。
比較只證明原版雙跑重現，沒有 remake 同狀態或正常玩家聲明。

## 七個固定分支

逐 byte 注入、步數與命令保存在[計畫](../internals/data/battle-discovery-plan.json)。
所有注入位址均為 dosgolem 線性位址，前綴為 `lin:`。

| 分支 | 注入／輸入 | 第四階段入口 | 邊界 | 結果 |
|---|---|---:|---:|---|
| detail3 | 池109的 Detail 改2 | 1256470043 | 14 | 升3、Hold=2；艦隊28的+6A由9變6 |
| camera | 格／像素鏡頭完整改為0,0 | 1256670406 | 16 | 置鏡到格68,31／像素1088,620 |
| cpu | 第一席 controller 改1 | 1227207802 | 5 | 八筆情報更新，沒有索敵 UI／等待 |
| twohuman | 第二席 controller 改0 | 1256470038 | 16 | 第一席一份報告；第二席沒有新發現 |
| swapped | controllers=0,0；陣營映射=1,0 | 1256620221 | 16 | 同一份報告改在第二席出現 |
| left | 第1230000000步左鍵240,200 | 1241469913 | 15 | 提前結束標頭等待；八筆展示仍按序完成 |
| right | 第1230000000步右鍵240,200 | 1241469915 | 16 | 同樣提前結束標頭等待 |

七組 a/b 共315份輸出逐 byte 相同。其中294份為98個邊界的 PNG、framebuffer、完整1 MiB RAM，
另21份為終態 PNG、framebuffer、色盤。逐幀保存5390筆 SHIP、686筆 FLEET、294筆 PLANET全紀錄。
每個分支的 seed 都未變。八筆候選池為109、135、137、138、139、140、141、147。

## 已證實的呈現契約

證據等級 `proven`，只限原始指令及指定受控局面。

1. `3826:0298` 建樣式1、色0x97、格位置4,7、內容20×3格的臨時窗。
   呼叫 `3617:000A` 印 caller本文，等 `DS:8F38` 訊息速度，再關窗；沒有畫頭像。
2. `25DB:0E78/0EDD` 配置40-byte暫時物件，優先值FF，寬高16×16，圖塊10。
   世界座標取 SHIP+34/+36；`0F6B` 扣鏡頭，`0F93` 移除／釋放／重畫並清指標。
   這條路沒有 Detail判斷或詳細資料文字，不能稱為詳細資料面板。
3. `1588:01A3` 的目標為 `trunc((X-216)/16)*16`、`trunc((Y-172)/20)*20`，使用帶號截斷除法。
   限界到0..1600／0..1680；上界相等也標成限界。
   沒有限界且兩軸差各小於128／60時保持原鏡頭。
4. 差至少448／360之一時走 `0183`，以 `trunc((X-224)/16)`、`trunc((Y-180)/20)`
   直接置鏡並限到格0..100／0..84。其他情況以原方位單步，旗標1重複、旗標0只走一步。
   camera從0,0到68,31，符合觀測者1320,812的直接置鏡公式。
5. `3765:006C` 用 DS:8EFE倒數，倒數歸零或事件碼1／2返回，並清倒數。
   左右鍵案例證實提前結束標頭等待；不推論其他按鍵或硬體 wall-clock。
6. Detail3在標頭／標亮入口仍保留艦隊28的+6A=9，第一筆等待時已變6。
   SHIP+2F也逐筆更新，不能在標頭出現時一次按新 Detail重畫所有候選。
   +6A的方向圖示定位見[084](../spec/084-battle-fleet-directions.md)；此次只證明更新區間，
   尚未定位本輪+6A的直接寫入指令。

## 限制與失敗分類

- camera注入改 RAM，原版先保留舊背景，直到置鏡才重畫。
  不把注入後標頭底圖當作重新繪製鏡頭0,0的畫面。
- 首次 camera探索寫 `0000`，實際只寫一個0 byte；第一次 swapped寫 `0100` 被拒。
  均為注入格式錯誤，改用空格分隔每個 byte並另建批次，舊輸出不列為正式收據。
- 首次 IDA clamp匯出漏 `--index`，沒有輸出；補參數後核 schema、雜湊與原始 bytes。
- 首次證據工具假定 FLEET不變，被 Detail3的+6A更新推翻。完整差異保留，不排除+6A。
- `2526:03F4/03F8` 讀顯示物件+1E決定是否還需移鏡。
  最小可見性契約與 `25DB:03E9` 快取寫入鏈是下一個窄任務。
- 完整第三階段仍DRAFT，正式報告／等待／戰鬥前續接、renderer及保存沒有新增通過聲明。
  第39輪座標修正與第38輪第二戰場限定成果保持，#79不關閉。

重生入口：[原版包裝器](../../tools/battle_discovery_oracle.sh)、
[旗標工具](../../tools/scripts/battle_discovery_plan.py)、[證據工具](../../tools/scripts/battle_discovery_evidence.py)。
完整資料 SHA-256 `383965f538fdf01ffad4a78b32957ad0372fd5f9a08440fe7d26210ac50507e1`。
