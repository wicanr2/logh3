# RE 36 — 開頭動畫：語法解出來了，像素對應還沒

日期：2026-09-07
執行器：dosgolem-logh3 `8c8af90`
狀態：**分段與語法 confirmed**（374 格逐格對上長度表、整個檔案吃光）；
**寫到哪一格 未解**。

## 先找到解碼器

`docs/internals/19` 量到的段值（`0110`／`01E6`／`2527`）在這一輪的執行裡
整體位移了 `+0x239`。用 `-vram-sites` 重找一次（要跑到 52,000,000 步，
動畫的第一格在 34,700,000 步才畫出來）：

```
041F:0206 / 0218 / 022A / 0241  ×90800 各   ← 四個 plane 的 rep movsw
0349:09E6                       ×56000      ← 清畫面
```

`-regs-at 041F:0206` 拿到來源 `DS:SI = 2760:0016`（線性 `27616`）。
對那裡設監看，寫入者是 **`041F:02D2`** 與 `041F:02A8`——解碼器就在同一段。

## 解碼器（手工解碼，`lin:41F0` 起 1 KB）

外層（`041F:01B2`）一格做四次：

```
lds si,cs:[000E]        ; 壓縮資料
mov bx,cs:[001C]        ; 目前吃到哪
add si,bx / sub bx,bx
les dx,cs:[0012]        ; 緩衝區 A ＝ plane 0
call 024B               ; 解一個 plane
add dx,46F0 / inc bx    ; ← 每個 plane 之間 inc bx：跳過一個分隔位元組
call 024B               ; plane 1
add dx,46F0 / inc bx
call 024B               ; plane 2
inc bx
les dx,cs:[0016]        ; 緩衝區 B ＝ plane 3
call 024B
inc bx
add cs:[001C],bx        ; 累加吃掉的位元組
```

內層（`041F:024B`）：

```
xor di,di / sub cx,cx
mov word cs:[0248],0100        ; 終止界限，之後會被改成 0050
loop:
  cmp di,cs:[0248] / jz ret
  mov ax,[bx+si] / inc bx      ; al ＝ 這個位元組、ah ＝ 下一個
  cmp ah,al / jnz single       ; 兩個不一樣 → 單一位元組
  inc bx                       ; 一樣 → 吃掉第二個，這是跳脫
  mov cx,[bx+si] / inc bx      ; 讀計數
  test ax,ax / jz skiprun      ; 位元組是 0 → 跳過，不寫
  test cl,80 / jnz one         ; 計數的第一個位元組有高位 → 一個位元組
  xchg ch,cl / inc bx          ; 否則兩個位元組，big-endian
  jmp xorrun
one: and cx,7F
xorrun:  add di,dx / xor es:[di],al / sub di,dx
         add di,50 / cmp di,46F0 / jnb wrap / loop xorrun
skiprun: add di,50 / cmp di,46F0 / jnb wrap / loop skiprun
wrap:    sub di,46EF / loop … / mov word cs:[0248],0050 / jmp loop
single:  add di,dx / xor es:[di],al / sub di,dx / add di,50 / … / jmp loop
```

整理成語法：

| 樣子 | 意思 |
|---|---|
| `XX YY`（XX ≠ YY）| 把 `XX` **XOR** 進目前這一格，只吃掉 `XX` 一個位元組 |
| `XX XX n`（`n` 有 bit7）| 把 `XX` XOR 進 `n & 0x7F` 格 |
| `XX XX n1 n2`（`n1` 沒有 bit7）| 同上，格數 ＝ `(n1 << 8) \| n2` |
| `00 00 …` | 一樣的兩種計數，但**只前進不寫** |

一格 ＝ 四個 plane 的資料，**每個 plane 後面接一個分隔位元組**
（外層那四個 `inc bx`）。

## 語法是對的：374 格逐格對上長度表

照這個語法從檔頭一路解到底：

- 每個 plane 都**正好**走完 80 × 227 ＝ 18,160 格才停；
- 每一格四個 plane ＋ 四個分隔位元組加起來，與 `IBMFLS16.350` 的
  374 個長度**逐一相符**；
- 全部吃掉 1,764,911 bytes ＝ `IBM16.350` 的大小，一個位元組不差。

第 0 格 78 bytes 拆開來是 `20 ＋ 1 ＋ 17 ＋ 1 ＋ 20 ＋ 1 ＋ 17 ＋ 1`。

## 但是寫到哪一格還不對

拿原版的緩衝區當底比對，我的輸出**一個位元組都沒對上**：

- 第 34,706,000 步（第一次搬到 VRAM 之前）倒出來的緩衝區有 3,105 個
  非零位元組——那是動畫開始前就在那裡的底圖；
- 第 35,300,000 步（第一格解完之後）倒出來的差了 **3,316** 個位元組；
- 拿第一份當底、把 374 格裡任何一格套上去，都沒有比「不套」更像。

監看到的第一個寫入是 `27616: 00 → 20`（第 0 格、值 0x20），
而照上面的語法，第一個 token 是「跳過 10,559 格」——對不起來。

所以 `add di,50` / `sub di,46EF` 那個走法（一欄一欄往下）**至少有一處
讀錯了**，或者解碼器不是從檔頭開始（`cs:[001C]` 在動畫開始前已經被推過）。

## 下一步

- 把 `cs:[001C]` 與 `cs:[000E]` 在第一次解碼時的值倒出來，就知道
  是不是從檔頭開始的（`-regs-at 041F:01B2`）。
- 把 `041F:0248` 附近那幾個 `cs:` 變數一起倒出來對。
- 語法既然對得上整個檔案，錯的多半只在「格 → 位移」這一段。
