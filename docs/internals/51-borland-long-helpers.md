# Datalight C 的 32 位元輔助函式（`3CC9` 那一段）

證據等級：**confirmed**，逐行讀出來的（[`docs/re/146`](../re/146-battle-damage-helpers.md)）。
檔名沿用舊名，因為有連結指向它。

`GIN3PS.EXE` 的 C runtime 是 Datalight C 3.10 Large（橫幅在 `ds:0000`，
見 [`docs/re/442`](../re/442-compiler-is-datalight-c.md)）。`long` 的乘除比較
都呼叫執行期輔助函式。看不懂這幾支就每一條算式都讀不下去，所以先釘死。

## 呼叫慣例

**兩個運算元放 `AX:BX` 與 `CX:DX`（都是「高位字:低位字」），結果回
`AX:BX`。**

| 位址 | 做什麼 |
|---|---|
| `3CC9:0009` | `AX:BX × CX:DX → AX:BX`（32 × 32 → 32） |
| `3CC9:0020` | `AX:BX ÷ CX:DX → AX:BX`，**無號**；餘數留在 `CX:DX`；除以 0 回 `0xFFFFFFFF` |
| `3CC9:0070` | 同上但**有號**：先取絕對值、呼叫 `0020`、再依號數補回去 |
| `3CC9:00B9` | 比較。高位字不同就直接帶著旗標回（呼叫端接 `jg`／`jle`）；相同時 `AX ← −1／0／1` |
| `3CC9:00CC` | 亂數：`ds:94DC` 的 32 位元種子 × `0x41C64E6D` ＋ `0x3039`，回高位 word `& 0x7FFF`（[`docs/re/87`](../re/87-original-rng.md)，remake 的 `rules.Rand`）。**呼叫端多半寫成 `3CD5:000C`**——同一個線性位址 `0x3CD5C` 的另一種段寫法 |

## 認出來的樣子

編譯器產生的呼叫幾乎都是這個形狀：

```
mov  bx, <運算元一的低位>
xor  ax, ax                  ; 或 cwd／mov ax,<高位>
mov  cx, <運算元二的高位>
mov  si, <運算元二的低位>
xchg dx, si                  ; 只是拿 SI 當暫存把值搬進 DX
call far 3CC9:xxxx
```

`xchg dx, si` 沒有別的意思，看到就跳過。回值要低位字時取 `BX`，
要整個 32 位元就是 `AX:BX`。

## 順帶

`39B5:000C`／`39B5:005C`（要塞砲的橢圓判定用）與 `3870:0643`／`3870:0614`
（焦距）走的是同一批慣例的定點數版本，還沒逐支讀完。
