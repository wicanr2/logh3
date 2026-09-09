# 201 — `2E54:063C` 的排序鍵：等級 → 統率力 → 運用，三段降冪

階段一（`2E54:0513`，在首都新編艦隊補守軍）先把首都那一格的提督清單
依 `2E54:063C` 排序，再照順序試建。那個比較函式比的是哪一項，
一直記在 `unsolved` 裡。整支讀完了。

## 反組譯

線性 `0x2EB7C` 起（`workplace/s3ai6.bin`，段 `2E54`）：

```
2E54:063C(a, b):                  ; a、b 是「提督編號」的遠指標
  les bx,[bp+12] / mov al,es:[bx] / xor ah,ah
  mov cx,1Ch / mul cx             ; 編號 × 28（ADMIRAL 一筆的長度）
  mov di,0B77Ch / add di,ax       ; A = ds:B77C + 編號 × 28
  （對 b 做同一件事 → B）

  mov al,es:[A+06] / mov cl,es:[B+06] / cmp al,cl / jbe … / mov ax,0FFFFh / retf
  mov al,es:[A+06] / mov cl,es:[B+06] / cmp al,cl / jnb … / mov ax,1     / retf
  mov al,es:[A+0B] / mov cl,es:[B+0B] / cmp al,cl / jbe … / mov ax,0FFFFh / retf
  mov al,es:[A+0B] / mov cl,es:[B+0B] / cmp al,cl / jnb … / mov ax,1     / retf
  mov al,es:[A+17] / mov cl,es:[B+17] / cmp al,cl / jbe … / mov ax,0FFFFh / retf
  mov al,es:[A+17] / mov cl,es:[B+17] / cmp al,cl / jnb … / mov ax,1     / retf
  xor ax,ax / retf
```

`A > B` 回 `−1`、`A < B` 回 `+1`，所以是**降冪**（大的排前面）。
三個位移在 `docs/internals/08` 都是 confirmed 的欄位：

| 位移 | 欄位 |
|---|---|
| `+0x06` | **等級**（軍階由它分段推出來）|
| `+0x0B`（`+11`）| **統率力** |
| `+0x17`（`+23`）| **運用** |

> **等級高的先建艦隊；同級比統率力，再同就比運用。**

## 為什麼是這三項

`+0x06` 與 `+0x0B` 正是 `2D76:0434` 算建艦隊分數用的那兩項
（`基底 = 等級 × 2`、`× 統率力 ÷ 100`，`docs/internals/08`）。
第三段的 `+0x17` 是類別 3（運用）對應的能力。這一支不是通用比較器——
它是**建艦隊那條路自己的**排序。

## 順帶

同一段之後緊接著

```
2E54:0725:  9a 30 07 54 2e    call far 2E54:0730
            9a d8 07 54 2e    call far 2E54:07D8
            cb                retf
```

——階段三就是「算 `ds:D58E`」＋「指派」兩支，與 `docs/internals/54`
記的一致。

## 一條方法

**比較函式讀起來比它看起來便宜。** 這一支 240 個位元組、七個
`cmp`／條件跳／`retf` 的重複段，形狀一眼可辨；把三個位移抄出來去對
`docs/internals/08` 的欄位表就結束了。它在 `unsolved` 裡躺了很久，
不是因為難，是因為沒有人把它排進去。
