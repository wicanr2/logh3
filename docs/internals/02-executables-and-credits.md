# 執行檔、封裝器與製作團隊

## 四支執行檔

| 檔 | bytes | 型態 | 角色 |
|---|---:|---|---|
| `GIN3.COM` | 622 | 無檔頭 .COM | 啟動器（confirmed）|
| `OPEN.EXE` | 52,265 | MZ | 開頭動畫與標題（confirmed）|
| `GIN3PS.EXE` | 189,536 | MZ，**GLUE V1.70 封裝** | 遊戲本體（confirmed）|
| `ENDING.EXE` | 64,677 | MZ | 結局與製作名單（confirmed）|
| `RESSOUND.COM` | — | 無檔頭 .COM | 假說：常駐音效驅動；開機到主選單的路徑上沒有被開啟 |

### GLUE V1.70（confirmed）

`GIN3PS.EXE` 的 MZ 檔頭後面緊接著一行明文：

```
0x0024: "GLUE V1.70 -- Program Coder (By:C.C.F.)"
```

那是執行檔壓縮／封裝器，所以整支 `GIN3PS.EXE` 用 `strings` 幾乎撈不到
任何可讀內容——程式碼與資料都要等它在執行期解開之後才看得到。逆向本體
只能在執行期做（dosgolem 的 `-peek`／`-find`／trace），不能靠靜態掃檔案。

`GIN3.COM` 也沒有可讀字串；`OPEN.EXE` 與 `ENDING.EXE` 沒有封裝。

單獨執行 `GIN3PS.EXE`（不經啟動器）會印出
`TINYPROG says, "Bad program file!"` 並以離開碼 255 結束（confirmed，
`docs/re/02` §2）。TINYPROG 是另一套 DOS 的壓縮／overlay 載入器殼。
**GLUE 與 TINYPROG 的關係未知**——可能是兩層（GLUE 包住一支 TINYPROG
程式），也可能 `TINYPROG says` 就是 GLUE 解壓殼自己的訊息。跑完到主選單
之後，記憶體裡搜不到 `TINYPROG` 也搜不到 `GLUE V1`（解壓殼已被覆蓋），
所以這個問題要單獨追才有答案。

### 工具鏈（confirmed）

`OPEN.EXE` 與 `ENDING.EXE` 內含
`MS Run-Time Library - Copyright (c) 1990, Microsoft Corp`
——Microsoft C 6.0 世代的執行期程式庫。

本體用的 C runtime 有一支 `int86x` 風格的中斷派送常式：它**不發 `int nn`**，
而是 `int 21h AH=35h` 取向量、自己疊一個中斷框架、再 `iret` 跳進去
（見 [`03`](03-startup-and-memory.md)）。這是 Microsoft C 的作法，也是
逆向與模擬時最容易漏掉的一條路徑。

## 製作團隊（confirmed，取自 `ENDING.EXE`）

名單以 Big5 存在 `ENDING.EXE` 的資料區（檔案位移約 `0x67xx`–`0x74xx`），
版面是「英文職稱 ＋ 中文姓名」。

### 日文原版

```
Presented by            BOTHTEC／QUEST
Program ＆ Graphics     Micro Vision

Producer                八卷  龍一
Director                青沼  實
Game Programming        酒井  敏光
Sub Director            大橋  敦己
Scenario Designer       今野  千尋
Sub Programming         中村  勝
                        石川  惠
Graphics                杉本  佳子
                        大橋  敦己
Sound Programming ＆
Music Composition       ミラージュ
Cover Illustration ＆
Logo Design ＆
Title Animation         加藤  直之
```

### 中文版

```
Chinese Version
Distribution            亨和企業．微波軟體
Chinese Message
Translation             邵  程衍
Chinese Version
Programming             Andrew Liao
```

### 版權行

```
(C)1994  亨和企業．微波軟體
(C)1993  BOTHTEC／QUEST
(C)1993  Micro Vision
(C)1988  田中芳樹．德間書局．德間ジャパン．キティフィルム
```

1988 那一行是原作《銀河英雄傳說》的權利方（田中芳樹、德間書店、
德間ジャパン、キティフィルム）。

## 結局文字（confirmed，取自 `ENDING.EXE`）

同一個資料區裡有結局劇情文字，例如同盟結局：

```
【同盟勝利】
在自由行星同盟的長途遠征之下
終於推翻了銀河帝國。
並且在帝國領土上實行了民主政治
使各地充滿了廣大民眾的歡呼聲。
後來，完全統一後的人類社會
寫下空前未有的繁榮歷史。
```

以及角色對話（楊威利與萊因哈特的台詞，含「可以給我一杯紅茶，好嗎？」）。
文字是純文字段落，沒有明顯的長度前綴或索引表——**排版方式與索引結構
尚未確認（未知）**。

