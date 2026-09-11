# 272 — 頭像編號直接當參數傳進去：`3958:0215`、`332E:0B8E` 讀出來了

日期：2026-09-11
執行器：dosgolem-logh3 `ef5be47`（未提交檔案 0 個）
輸入：`workplace/codeall.bin`（收據 `codeall`，`-dump-mem lin:9000:217088`，
涵蓋線性 `0x09000`–`0x3E000`；`GIN3.COM` SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`）
工具：`tools/scripts/seg_to_com.py` ＋ IDA 9.4 headless
（`tools/scripts/disasm_funcs.py`）、`farcall_to.py`、`hexat.py`
狀態：**confirmed**（純靜態，沒有另外跑執行器）

`docs/internals/09` 的「第 n 張 ＝ 提督索引 n」原本靠四條旁證撐著：收據、
制服色的分界、人物傳的內容對應、以及「提督記錄的 28 個位移沒有一欄等於
記錄索引」。缺的是繪圖端那一段程式碼。這一篇把它讀出來，結論與旁證一致，
而且多出一條**同一個區塊內**的證據。

## 怎麼查的

`docs/internals/09` 已經記著解碼器在 `348A:02F9`，`docs/spec/034` 已經記著
易主報告那條路上的呼叫端是 `3958:0215`。兩個位址都在 `codeall.bin` 的範圍
內，所以不必再跑一次原版：切段、線性反組譯就好（`rulebook/62`）。

先驗這份傾印的段位對不對得上：`348A:02F9` ＝ 線性 `0x34B99`，倒出來是
`AD 38 E0 74 13 26 88 05 03 FB …`，與 `docs/internals/09` 記的解碼器逐位元組
相同。那份反組譯來自另一批收據（HEAD `2833ec9`），兩次載入的佈局一樣。

```
tools/py.sh tools/scripts/seg_to_com.py workplace/codeall.bin 9000 3300 7000 \
    workplace/ida/fx3300.com          # 線性 0x33000–0x3A000
tools/ida.sh raw idat -A -B /work/fx3300.com
tools/ida.sh script fx3300.com disasm_funcs.py /work/d3300.txt 6795:80 06E1:1C0 1A7C:140
```

## `3958:0215` — 編號只有兩個特例，其餘原樣往下傳

線性 `0x39795`，簽章 `55 89 E5 80 7E 06 FF 75 04 C6 46 06 52 8A 46 06 3C FE 74 10`
（在傾印裡唯一）。

```
0215  55              push bp
0216  89 e5           mov  bp,sp
0218  80 7e 06 ff     cmp  byte [bp+6],0FFh    ; 參數 1 ＝ 頭像編號
021C  75 04           jnz  0222
021E  c6 46 06 52     mov  byte [bp+6],52h     ; 0FFh → 82（那一格空白）
0222  8a 46 06        mov  al,[bp+6]
0225  3c fe           cmp  al,0FEh
0227  74 10           jz   0239                ; 0FEh → 不重載，畫緩衝區裡現有的
0229  50              push ax                  ; 編號
022A  ff 36 3a f1     push word [ds:F13A]      ; 解碼緩衝區 far 指標（段）
022E  ff 36 38 f1     push word [ds:F138]      ;                   （位移）
0232  9a 01 04 2e 33  call far 332E:0401       ; 讀 FACE.AMG 第 n 筆進緩衝區
0237  89 ec           mov  sp,bp
0239  ff 36 3a f1     push word [ds:F13A]
023D  ff 36 38 f1     push word [ds:F138]
0241  8a 46 0a        mov  al,[bp+0Ah]         ; 參數 3 ＝ 列
0244  30 e4           xor  ah,ah
0246  50              push ax
0247  8a 46 08        mov  al,[bp+8]           ; 參數 2 ＝ 行
024A  30 e4           xor  ah,ah
024C  50              push ax
024D  9a dc 01 8a 34  call far 348A:01DC       ; 畫出來
0252  89 ec           mov  sp,bp
0254  5d              pop  bp
0255  cb              retf
```

三個參數都是 byte（Borland C 的 `unsigned char`，推到堆疊上占一個 word）。
**編號沒有查表、沒有加減、沒有換算**，唯一的兩個特例是：

| 傳進來的值 | 做什麼 |
|---|---|
| `0xFF` | 換成 `0x52` ＝ 82，也就是 `FACE.AMG` 那一格空白 |
| `0xFE` | **跳過載入**，直接把緩衝區裡上一次解出來的那張再畫一次 |
| 其餘 | 原樣傳給載入常式 |

`0xFF` 那條與資料對得上：提督索引的「沒有這個人」在資料裡就寫 `0xFF`
（`ADMIRAL.DT*` 的 `+0x07` 副官欄、艦隊記錄的 `+0x04`），碰到就畫第 82 張。
`docs/internals/09` 記的「83 格裡有 82 個人、第 82 格整張單色 9」不是多出來
的一格，是**設計上的空位**。

`3958:0256` 是同一支的第二個版本（簽章同上但末尾是 `74 15`），差別只在緩衝區
位移多加 `0x1000`——同時要放兩張頭像用的。傾印涵蓋的 `0x09000`–`0x3E000` 內
沒有任何 far call 打到它；範圍外還有程式碼（例如 `056A` 段），所以只能說
「這份傾印裡沒看到呼叫端」，不能說它是死碼。

## `332E:0401` → `332E:0B8E` — 索引表是 `n × 6`

`332E:0401` 只是薄薄一層包裝：

```
0401  55 89 e5        push bp / mov bp,sp
0404  ff 76 08        push [bp+8]          ; 緩衝區（段）
0407  ff 76 06        push [bp+6]          ;         （位移）
040A  ff 76 0a        push [bp+0Ah]        ; 編號
040D  1e              push ds
040E  bb af 6a        mov  bx,6AAFh        ; ds:6AAF ＝ "face.amg"
0411  53              push bx
0412  31 c0           xor  ax,ax
0414  50              push ax              ; 類別 0（選路徑前綴）
0415  9a 8e 0b 2e 33  call far 332E:0B8E
041A  89 ec           mov  sp,bp
041C  83 f8 ff        cmp  ax,0FFFFh
041F  75 02           jnz  0423            ; 成功就回
0421  eb e1           jmp  0404            ; 失敗就整段重來（無限重試）
0423  5d              pop  bp
0424  cb              retf
```

`ds:6AAF` 讀出來是 `66 61 63 65 2e 61 6d 67 00` ＝ `face.amg`；同一張字串表
接下去是 `keireki.dat`、`admiral.dt%c`、`stellar.dt%c`…，`332E` 段裡一支
包裝對一個檔。

`332E:0B8E` 是通用的「讀 `.AMG` 第 n 筆」（第一個參數選路徑前綴，
`ds:69A2 + 類別 × 32`）。索引那一段的函式內簽章是
`8A 46 5C 30 E4 B9 06 00 F7 E1 89 C3`（落在 `0BE3`，在傾印裡唯一）：

```
0B8E  55 56 57       push bp / si / di
0B91  83 ec 4c       sub sp,4Ch          ; → 參數從 [bp+56h] 起
…     組出路徑、open()，handle 進 di
0BE3  8a 46 5c       mov al,[bp+5Ch]     ; ← 編號，讀的是 byte
0BE6  30 e4          xor ah,ah
0BE8  b9 06 00       mov cx,6
0BEB  f7 e1          mul cx              ; ★ ax = n × 6
0BED  89 c3          mov bx,ax
…     lseek(di, n×6, SEEK_SET)
0C10  b8 06 00       mov ax,6
…     read(di, &rec, 6)                  ; rec ＝ u16 長度 ＋ u32 位移
…     lseek(di, rec.位移, SEEK_SET)
…     read(di, 緩衝區, rec.長度)
…     close(di) / 回 0
```

`n × 6` 直接當檔頭索引表的位移，**中間沒有任何轉換表**。這一段同時把
`docs/internals/09` 的檔案結構釘死：索引表每筆 6 bytes ＝ `u16 長度` ＋
`u32 位移`，而 `FACE.AMG` 的第一筆位移 `0x01F2 ÷ 6 ＝ 83` 就是筆數。

## `348A:01DC` — 落點換算與標頭怎麼走

`3958:0215` 的參數 2／3 傳到 `348A:01DC`（函式內簽章
`46 46 2E 89 36 D8 01 B9 01 00 51`，落在 `01FB`；入口那幾行含 far call 不適合當簽章），
它把兩個 byte 丟給 `3A8D:003A`（入口簽章
`8A E1 32 C0 8B D8 D1 EB D1 EB 03 D8 8A CD 32 ED 03 D9 81 C3 80 0C CB`）：

```
003A  8a e1 32 c0    mov ah,cl / xor al,al   ; ax = 列 × 256
003E  8b d8          mov bx,ax
0040  d1 eb d1 eb    shr bx,1 / shr bx,1     ; bx = 列 × 64
0044  03 d8          add bx,ax               ; bx = 列 × 320
0046  8a cd 32 ed    mov cl,ch / xor ch,ch   ; cx = 行
004A  03 d9          add bx,cx
004C  81 c3 80 0c    add bx,0C80h            ; ＋ 3200
0050  cb             retf
```

mode 12h 每列 80 bytes，所以

```
x ＝ 行 × 8
y ＝ 列 × 4 ＋ 40
```

`348A:01DC` 接著逐段走緩衝區的標頭（`cs:01D8` 是目前這一段的指標、
`cs:01DA` 是目的位址）：

```
01F8  c5 76 0a        lds si,[bp+0Ah]      ; si = 緩衝區
01FB  46 46           inc si / inc si      ; → +2（+0 的段數由 0228 那版讀）
01FD  2e 89 36 d8 01  mov cs:[01D8],si
0202  b9 01 00        mov cx,1             ; 01DC 固定畫一段
0205  51              push cx
0206  2e 8b 36 d8 01  mov si,cs:[01D8]     ; ← 迴圈頭
020B  8a 24           mov ah,[si]          ; +2 ＝ 畫法
020D  46              inc si               ; +3 跳過不看
020E  22 e4           and ah,ah
0210  75 05           jnz 0217
0212  e8 5e 00        call 0273            ; 0  → 0281 ＋ 02BA（→ 02F9，本篇這支）
0217  e8 60 00        call 027A            ; 非 0 → 0281 ＋ 036C（另一支解碼器）
021B  e2 e8           loop 0206
```

兩條路共用的前置是 `348A:0281`，它把 `+4`／`+6` 吃掉：

```
0281  1e              push ds
0282  b8 b3 54 8e d8  mov ax,54B3h / mov ds,ax   ; 資料段
0287  83 3e fe 8e 00  cmp word ds:8EFE,0
028C  75 f9           jnz 0287                   ; 等 ds:8EFE 歸零才往下
028E  bb 10 8f        mov bx,8F10h
0291  a0 37 8f d7     mov al,ds:8F37h / xlat     ; ds:8F10[ds:8F37]
0295  32 e4           xor ah,ah
0297  a3 fe 8e        mov ds:8EFE,ax             ; 再把它設成查到的值
029A  1f              pop ds
029B  46              inc si                     ; → +4
029C  8b 04 46 46     mov ax,[si] / si += 2      ; +4 ＝ 這一段的位元組數
02A0  03 c6           add ax,si
02A2  2e a3 d8 01     mov cs:[01D8],ax           ; 下一段的起點
02A6  2e 8b 3e da 01  mov di,cs:[01DA]           ; 落點（3A8D:003A 算的）
02AB  8b 04 46 46     mov ax,[si] / si += 2      ; +6 ＝ 目的位址位移
02B0  03 f8           add di,ax                  ; si 停在 +8（w／h）
02B2  c3              retn
```

接下來 `348A:02BA` 把 `+8`／`+10` 讀進去，然後才落到 `02F9`：

```
02D0  8a 34           mov dh,[si]      ; +8  ＝ 直行數 w
02D9  8a 14           mov dl,[si]      ; +10 ＝ 每直行的 word 數 h
02EC  bb 50 00        mov bx,50h       ; 80 ＝ mode 12h 每列位元組數
02F0  e8 06 00        call 02F9        ; ← docs/internals/09 記的那個迴圈
```

所以影像標頭的前八個位元組不是「u32 ＋ u32」，是
`u16 段數 / u8 畫法 / u8 跳過 / u16 段長 / u16 目的位移`。FACE.AMG 只有一段、
位移 0，兩種讀法數值上相同；多段的 `.A98` 要用這一份
（`348A:0228` 就是「段數讀 `+0`」的那個版本，呼叫端 `3958:01A7`）。

兩個已經量過的畫面對得上：易主報告的呼叫端傳 `(0x44, 0x23)` → `(544, 180)`
（`docs/spec/034` 量到的落點）；提督詳細資料傳 `(0x38, 0x0F)` → `(448, 100)`
（`docs/spec/012` 量到的落點）。**座標換算是獨立驗證**——它與編號無關，
卻能證明「哪個呼叫端是哪個畫面」的對應沒有猜。

## 呼叫端普查：編號從哪裡取

`farcall_to.py workplace/codeall.bin 9000 39795` 在傾印範圍內找到 **22 處**
far call。逐一反組譯呼叫前那幾行（`x`／`y` 已換算成像素）：

| 線性 | (x, y) | 編號取自 |
|---|---|---|
| `09342`、`09751` | (416, 80) | `ds:8F4F[ds:8F1D]` — 內政專任參謀 |
| `0A184`、`0ABB7` | (544, 60) | `ds:8F4D[ds:8F1D]` — 情報專任參謀 |
| `0B78E` | (448, 180) | far 指標指到的記錄 `+0x04` |
| `0D98B` | (448, 100) | 區域變數 `[bp+24h]` ★ |
| `0E28C` | (448, 60) | far 指標指到的記錄 `+0x00` |
| `0F06A`、`0F157`、`0F731`、`0F79E` | (448, 60) | 記錄 `+0x04` |
| `0F838`、`0FEAE` | (464, 80) | `ds:8F4B[ds:8F1D]` — 軍事專任參謀 |
| `11C21`、`11ECE`、`11FA7` | (544, 180) | 記錄 `+0x00`（推 word，用低位元組）|
| `11D81` | (432, 160) | 記錄 `+0x00` |
| `1232C` | (432, 180) | `ds:8F4B[bx]` |
| `15300` | (352, 100) | 記錄 `+0x04` |
| `38446` | (144, 192) | 區域變數 `[bp+22h]` ★ |
| `38627` | (544, 180) | `ds:8F4D[si]`，`si == 0FFFFh` 時改用 `[bp+0Ch]` |
| `386E0` | (352, 200) | `ds:8F4D[si]`，`si == 0FFFFh` 時改用 `[bp+32h]` |

三張 `ds:8F4B`／`8F4D`／`8F4F` 表就是六個專任參謀（`docs/internals/30`
的 `WORK +0x37`、`docs/re/61`／`re/268`），每格存的**就是提督索引**。

`38627` 那一筆是易主報告：落點 `(544, 180)` 與 `docs/spec/034` 量到的相同，
編號取 `ds:8F4D[座位陣營]` 與 `docs/re/248` 從記憶體比對出來的相同。
**兩條路各自獨立，答案一樣。**

### 最短的一條證據：`0D98B`

提督詳細資料那個畫面（落點 `(448, 100)` ＝ `docs/spec/012`）的呼叫端，
**在同一個基本區塊裡**用同一個變數做兩件事：

```
4965  ff 76 28        push [bp+28h]
4968  8a 46 24        mov  al,[bp+24h]     ; ← 同一個變數
496B  30 e4           xor  ah,ah
496D  b9 1c 00        mov  cx,1Ch
4970  f7 e1           mul  cx              ; × 28 ＝ 提督記錄長度
4972  8c d9           mov  cx,ds
4974  be 7c b7        mov  si,0B77Ch       ; ds:B77C ＝ 提督表
4977  03 f0           add  si,ax           ; 提督表 ＋ n × 28
4979  51 56           push cx / push si
497B  9a 78 00 92 0d  call far 0D92:78     ; 印這一筆的欄位
4980  89 ec           mov  sp,bp
4982  b0 0f 50        mov al,0Fh  / push ax  ; 列 15 → y = 100
4985  b0 38 50        mov al,38h  / push ax  ; 行 56 → x = 448
4988  ff 76 24        push [bp+24h]        ; ← 同一個變數，當頭像編號
498B  9a 15 02 58 39  call far 3958:215
```

`ds:B77C + n × 28` 是提督記錄（`docs/internals/08`），`FACE.AMG` 的
`n × 6` 是同一個 `n`。**「第 n 張 ＝ 第 n 筆」不再需要靠旁證推。**

`38446` 是同一個形狀的第二例：`[bp+22h]` 先 `imul 1Ch` 加進一個記錄指標，
再原樣推給 `3958:0215`。

## 順帶解出：艦隊記錄 `+0x04` ＝ 提督索引

六個呼叫端取的是「far 指標指到的記錄 `+0x04`」。`docs/internals/33` 原本只記
「`+0x04` ＝ `0xFF` 表示這一筆沒在用」，沒說非 `0xFF` 是什麼；而 `+0x05` 是
副官的提督索引，`ADMIRAL.DT*` 的 `+0x07` 也是副官的提督索引
（`docs/spec/013`）。兩者可以互相檢查：

```
提督[記錄 +0x04][+0x07]  ==  記錄 +0x05  ?
```

拿 `workplace/dm_pre.bin` 與 `chain_post.bin`（線性 0 起的 1 MB 傾印，
`ds:C074` ＝ 線性 `0x60BA4`、`ds:B77C` ＝ `0x602AC`）的 32 個艦隊槽跑一遍：
**21 筆在用的全部相符，0 筆不符**。帝國 16 個槽的值落在 0–15、同盟槽落在
40–79，與 `docs/internals/09` 記的制服分界（39／40）同一側。

所以艦隊記錄 `+0x04` ＝ **提督索引**，`0xFF` ＝ 這一槽沒在用——而
`3958:0215` 把 `0xFF` 換成第 82 張，正好讓「沒有提督」畫出那張空的。

## 結論等級

| 斷言 | 等級 |
|---|---|
| `3958:0215` 收到的編號只有 `0xFF`→82、`0xFE`→不重載兩個特例 | confirmed（反組譯位元組）|
| `332E:0B8E` 用 `n × 6` 直接定位索引表，沒有轉換表 | confirmed（反組譯位元組）|
| `x ＝ 行 × 8`、`y ＝ 列 × 4 ＋ 40` | confirmed（反組譯 ＋ 兩個量過的落點）|
| `FACE.AMG` 第 n 張 ＝ `ADMIRAL.DT*` 第 n 筆 | confirmed（`0D98B` 同區塊雙用途 ＋ `docs/re/271` 的四條旁證）|
| 艦隊記錄 `+0x04` ＝ 提督索引 | confirmed（21／21 交叉檢查 ＋ 陣營區間）|
| 影像標頭前八個位元組是 `u16 段數 / u8 畫法 / u8 跳過 / u16 段長 / u16 目的位移` | confirmed（`348A:01DC`＋`0281` 的位元組）|
| `ds:8EFE` 那段是畫面節拍 | **未知**——只讀到「等它歸零、再設成 `ds:8F10[ds:8F37]`」，誰把它減下去沒查 |
| 22 處呼叫端是全部 | **否**——傾印只涵蓋 `0x09000`–`0x3E000`，範圍外還有程式碼 |

`docs/re/271` 的結論**沒有被推翻**：對應確實是 1:1，中間沒有表。這一篇把它
從「四條旁證」升到「繪圖端的程式碼直說」。

## 一條方法

**呼叫端就在手上的傾印裡，只是沒人切那個段。** `348A:02F9`（`docs/internals/09`）
與 `3958:0215`（`docs/spec/034`）兩個位址各自被記了好幾週，`codeall.bin`
也早就涵蓋它們；缺的只是「把 `3958` 段切出來線性解一次」這一步。
整條鏈（編號 → 檔案 → 畫面）從頭到尾都是靜態的，沒有跑執行器。
