# 303 — 戰術「資料」清單每次捲一列，視窗固定十四列

日期：2026-09-14  
輸入：`workplace/ida/seg113A.com`（SHA-256
`dffb997762ad0a281d9afed97031d2ab90356887371a48faa7bfa318e557e36b`）  
工具：IDA Pro 9.4 headless、專案 `tools/ida.sh`；下列定位採該 COM 的 IDA
位址空間 `seg000:offset`，另列檔案偏移時明寫 `file+`  
狀態：上下界、步距、可見列數與活動偏移 **proven**（原始位元組及資料流）；
換頁重建清單時歸零 **proven**（初始化位元組）

## 結論

「資料」的清單視窗固定顯示十四列。右側上、下按鈕分別把同一個活動偏移
`ds:9B86` 減一、加一；按住時處理器會反覆輪詢，所以每一輪只移一列。
偏移不會小於 0，也不會使 `偏移 + 14` 大於或等於總列數 `ds:9B88`。
因此合法範圍是：

```text
0 <= ds:9B86 <= max(0, ds:9B88 - 14)
```

清單初始化在 `file+0x2988` 先 `xor ax,ax`，接著兩個 `a3` 依序把
`ds:9B86`、`ds:9B88` 清為零：

```text
file+0x2988  31 c0           xor ax,ax
file+0x298A  a3 86 9b        mov ds:9B86,ax
file+0x298D  a3 88 9b        mov ds:9B88,ax
```

這表示三個頁籤共用目前正在顯示之表格的一個偏移；開窗或換頁重建表格時，
不是替各頁籤保留各自的位置。

## 上下按鈕處理器

IDA `seg000:4608` 起的第一組處理器使用參數 `[bp+6]` 判斷熱區編號。
編號 4 走上捲，其餘同組的編號 5 走下捲；第二組
`seg000:467E` 起只有重畫函式不同，界線完全相同。

下捲的原始資料流（`seg000:460A`–`463C`）：

```text
mov  ax,ds:9B86
add  ax,0Eh
cmp  ax,ds:9B88
jnb  stop
inc  word ptr ds:9B86
call redraw
jmp  poll_again
```

上捲的原始資料流（`seg000:463E`–`466B`）：

```text
cmp  word ptr ds:9B86,0
jz   stop
dec  word ptr ds:9B86
call redraw
jmp  poll_again
```

兩條路每輪都先呼叫 `3765:0255(0)`；回 0 就停止，非零才處理下一列，這是
按住按鈕時重複捲動的閘門。`seg000:2DA8` 的繪製迴圈則從
`ds:9B86 + 列序號` 取資料，並在列序號到 `0x0E` 時停止，獨立證實十四列
不是 remake 端推定的版面常數。

執行期熱區清單 `ds:34CE` 的第 4、5 筆都是一格寬，格座標分別是
`(25,2)`、`(25,15)`；加上視窗內容原點後，實際可點矩形是
`x=432..447`、`y=120..139` 與 `y=380..399`。這比畫出的框
`x=434..447` 向左多 2 像素，命中測試不可直接拿繪圖矩形代替熱區。

## remake 對應

`Game.battleDataOffset` 保存活動偏移；開窗及換頁歸零。上下熱區每次點擊移動
一列並夾在上述範圍，繪製只收到偏移後最多十四列。點選可見列時，明細索引
必須使用 `偏移 + 畫面列`，否則捲動後永遠會開到前十四筆。

自動測試 `TestBattleDataScrollMatchesOriginal` 固定 40 列，驗證下界 26、
單列步距、邊界不關窗、換頁歸零與明細絕對索引；
`TestBattleDataScrollHitboxesMatchOriginal` 驗證上下按鈕矩形及中間槽不誤觸。
