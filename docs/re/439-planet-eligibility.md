# 439：行星反擊的獨立判定

2026-10-06，第111輪，沿[#91](https://github.com/wicanr2/logh3/issues/91)。
第110輪只讀已過鎖定計算的父RAM。本輪重新執行原版000C及0091，確認共同閘門錯誤。
規格入口[119](../spec/119-planet-assault-eligibility.md)，事實保存於[原版向量](../internals/data/planet-eligibility.json)。

## 方法與身份

GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
完整父eng-t5.state SHA-256 `c6ce5af3a8d52acc415ac1029a29a0905e9c95e02ac0e0ba547e71300fd947c1`。
dosgolem為f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，工作樹乾淨，Go1.24.13。
沿專用工具包裝器與golang:1.24-bookworm，20000步固定IRQ，沒有CycleClock。
原始輸入唯讀，UID1000，網路關閉，有界CPU／記憶體／步數與事件筆數。

CPU與堆疊注入完整記錄：保留父暫存器，以far返回20B4:FFFE呼叫000C，在返回前停。
非基準案明示修改現存艦隊31與行星記錄，先放舊target17／bitsFFFF確認重設。
種子執行前固定12345678一次，不更換種子或挑結果。
矩陣不依賴舊目標，也不注入預期位元。

```sh
bash tools/planet_eligibility_oracle.sh a matrix2
bash tools/planet_eligibility_oracle.sh b matrix2
bash tools/planet_eligibility_oracle.sh a pipeline
bash tools/planet_eligibility_oracle.sh b pipeline
bash tools/gui.sh xrun python3 tools/scripts/planet_eligibility_receipt.py
```

輸出在workplace/oracle/planet91-r111-{matrix2,pipeline}-{a,b}/。
342份原始檔的路徑、大小與SHA-256保存在結構化原版事實內。
工具拒絕覆寫；matrix是先前探索輸出，正式引用matrix2。
分類為明示direct-entry的同輸入規則診斷；不宣稱正常玩家完整回合同態。

## 已證實控制流

原版CS:IP與IDA EA、dosgolem線性位址各自保存，不混用。
IDA9.4輸入landing48-r108-driver.com SHA-256 `eb07213f45cbcab0b002fa31570c632f686ca8e26ccd05b41f36d8e088ae5d3e`。
匯出workplace/ida/landing48-r108-driver.json仍保持原名及原運算元。

| 原版CS:IP | 原始bytes | 已證實語意 |
|---|---|---|
| 20B4:0153 | 7F25 | 超過150跳至017A，仍可判反擊 |
| 20B4:015C | 751C | Flag0成立跳至017A |
| 20B4:0169 | 750F | Action低四位不為1跳至017A |
| 20B4:016E | 26804F7001 | 登陸OR bit0 |
| 20B4:01A5 | 26804F7002 | 獨立反擊OR bit1 |
| 20B4:03D3 | 894E44 | 有相符target即設開場旗標 |
| 20B4:03D8 | 26F6477001 | bit0另控制提示字串 |
| 20B4:03E7 | 837E4400 | 開場只判target存在 |

40案兩跑120份完整RAM相同。一般行星Action0或Flag0仍得bits2。
DEF6000距離151至240仍得bits2，241不鎖；DEF0查詢半徑60，距離61只留登陸bit0。
Action11與01相同；要塞種類1或2且GunState3僅禁止反擊。
友軍與零艘旗艦不鎖。兩顆行星累積bits，最後符合者覆寫target。

## 只受反擊續行

新算完整七艦隊目標後恢復父CPU，保留新目標。以固定滑鼠move／press／release續行。
0486進場種子仍12345678，呼叫0503、0681與08B1；沒有121F占領結果。
31的target0／bits2，總損失253艘，最後種子0B719151。
七艦隊所有分隊與三顆行星的typed資料都由進場與末端RAM解析，沒有只保留旗艦。

- phase6-input-mem.bin SHA-256 `79a0d211cc3174912b666d8ede72bd03158c4b80f80400a2b465e266d43fea3e`。
- phase6-end-mem.bin SHA-256 `dd0ef6deec00d3eae7a51d1f29692f56220e7c59eb5db1326ac8c9c22fb185bc`。

本輪沒有新解RND內部呼叫序列，不擴張整回合骰序一致的門檻。
舊149的共同「再加」解釋被獨立分支及實際原版矩陣否定，149歷史保留，019追加現行勘誤。
061政策及#48正式資料確認仍沿使用者既有決定。
