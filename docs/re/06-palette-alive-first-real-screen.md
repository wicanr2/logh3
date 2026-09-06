# RE 06 — 調色盤修好，拿到第一張真色彩畫面

日期：2026-09-06
執行器：dosgolem-logh3 分支 `logh3-com-support`（`769b030`）
輸入：原版未修改 `org_game/extracted/GIN3.COM`（唯讀，SHA-256 見 re/01）
規格：dosgolem-logh3 `docs/spec/011-bios-palette-and-vector-stubs.md`
前篇：`05-menu-reached-heap-starved-by-exec-leak.md`

推論標籤：confirmed／強證據／假說／未知。

---

## 1. 結果（confirmed）

主選單以遊戲自己的顏色畫出來：黑底、青綠色選單框、灰白外框、白色中文字
**帝國／同盟／對戰／取進度**。`沒實作的服務（0 種）`。

證據檔（本機，gitignore）：`workplace/re06/`
- `screen.bin` 色號陣列 640×480，SHA-256
  `25d9196a3b596cee06b623da0f7b15f1a5775a83987be30babc492a81431711c`
- `pal.bin` DAC 256×3，SHA-256
  `28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8`
- `screen.png`、`report-60M.txt`

**可重現性**：同一命令連跑兩次，`screen.bin` 與 `pal.bin` 逐位元組一致，
完整報告除了輸出檔名之外完全相同。

## 2. 不是淡入沒走完，是兩個缺口疊在一起（confirmed）

re/05 §5 把黑畫面記成「調色盤淡入沒有走完，驅動來源未知」。實際是
dosgolem 這邊的兩個獨立缺口。

### 2a. int 10h AH=10h 的子功能接錯

屬性調色盤與 DAC 是兩層不同的東西（色號 → 屬性暫存器 → DAC）。

- `AL=02h` 是「設 **17 bytes** 的屬性暫存器 ＋ overscan」，被接成
  「設 **768 bytes** 的 DAC」。遊戲傳進來的是識別調色盤 `0,1,2,…,15`，
  於是 DAC 前幾筆變成 `(0,1,2)(3,4,5)(6,7,8)…`——re/04 §1 記成
  「淡入沒有進展」的那個暗值，來源就是這裡。
- `AL=10h` 是「設**單筆 DAC**」（BX ＝ 索引、DH ＝ R、CH ＝ G、CL ＝ B），
  被接成「設屬性暫存器」。反組譯 `3C5B:0085` 可以逐位元組確認介面：
  `mov bl,[bp+6]` 索引、`mov cl,[bp+8]` 藍、`mov dh,[bp+0Ah]` 紅、
  `mov ch,[bp+0Ch]` 綠、`mov ax,1010h`、`int 10h`。
- 讀回的那半（`07/08/09/15/17`）完全沒有。遊戲的淡入淡出是
  「讀回現有調色盤 → 存起來 → 塗黑 → 畫 → 用存起來的值淡回去」，
  讀回沒實作它就存到未初始化的緩衝區，淡回來的是全黑。

### 2b. 取向量後直接跳的 BIOS 呼叫完全沒有執行

`3C89:0022` 是 C 的 `int86x`，反組譯後很清楚：它**不發 `int nn`**。

```
mov al,[bp+0Ch] / mov ah,35h / int 21h   ; 取向量 → ES:BX
push es / push bx                        ; 目標 handler
（載入呼叫端給的暫存器組）
iret                                     ; 用自己疊的框架跳進 handler
```

dosgolem 的服務層掛在 CPU 執行 `INT` 指令的 hook 上，而所有向量都指向
一個裸 `iret`——所以這條路**安靜地什麼都不做**：暫存器原樣回去，沒有
錯誤、沒有 unimplemented 記錄。淡入的套用（`AX=1012h`）正是走這條。

改成每個向量各有一段 `int n` ＋ `iret` 的 stub。跳進去會重新走進 hook；
程式自己裝了 handler 時，stub 裡的 `int n` 會真的跳過去，回來再 `iret`
回呼叫端，順帶具備鏈接行為。

修正後 `AH=10h` 呼叫由 37 次增為 41 次，DAC 出現實際顏色。

## 3. 淡入的機制（confirmed，順帶解出來）

段 `3931` 是調色盤模組：

- `ds:F2D6/F2D8` 是來源調色盤的 far 指標（每筆 3 bytes，4 位元分量）。
- `ds:F2D4` 是淡入等級，`3931:017C` 依參數設成 `0`（全黑）或 `0x10`（全亮）。
- 套用迴圈：`分量 × ds:F2D4 >> 2`，16 筆，再由 `3931:022F` 組出
  `AX=1012h`（設一段 DAC）交給 `int86x`。

所以整段淡入邏輯本來就在，缺的只是它呼叫的兩個 BIOS 介面。

## 4. 這張畫面能宣稱什麼

dosgolem 跑的是原版未修改的執行檔，所以這是**原版側**的畫面，可以當之後
差分的基準。但還不是 `same-state` 對拍——對拍要兩邊，而 remake 側目前
不存在。

顏色全部來自遊戲自己寫進 DAC 的資料，不是我們補的預設調色盤。

## 5. 還沒到的地方

- 底色是色號 3（DAC 全 0 ＝ 黑）。這依賴 spec 010 §3 的 mode 0
  set/reset；硬體語意上正確，但底色的正確性目前只有「整片淺灰顯然不對」
  這個反面論證撐著，還沒有正面證據。
- 主選單還沒存成正式的 oracle 收據（run 目錄 ＋ metadata ＋ `docs/playtest/`）。
- 選單以外的畫面、資料格式、規則都還沒碰。

## 6. 重現命令

```
cd ~/cht/dosgolem-logh3
DOSGOLEM_ORIG=~/cht/logh3/org_game/extracted tools/go.sh run ./cmd/probe \
  -exe /orig/GIN3.COM -root /orig -steps 60000000 -tick 20000 \
  -click-x 320 -click-y 200 -click-at 5000000 \
  -dump-ega /src/workplace/re06/screen.png \
  -dump-palette /src/workplace/re06/pal.bin
```

約 9 秒。反組譯用 `workplace/re05/disasm.sh <起始偏移> "<hex bytes>"`，
記憶體用 `-peek "<段>:<偏移>:<長度>"`，AH=10h 的參數看報告的
`int 10h AH=10h` 段。
