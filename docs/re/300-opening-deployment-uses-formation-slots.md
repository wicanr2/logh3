# 300 — 開場佈陣直接使用艦艇既有的 `+0x05` 隊形槽

日期：2026-09-13  
輸入：`workplace/s3ai6.bin`（SHA-256
`427c15414ce8d265551a05ecd2d49e8157ecede82a80154b6319c7ced4a8d3e0`）、
`workplace/ida/seg297B.com`（SHA-256
`252b5b017484701977180498a5fcefe0f0ff4bba935234fd0a3354863dfc0322`）、
`workplace/ida/s3seg2BE1.com`（SHA-256
`bccf18d1fd4c1061a3ebb0639e145d86786d7382cb96ddcb71c8bece6cad1fbb`）  
工具：IDA Pro 9.4 headless，專案 `tools/ida.sh`，原始遊戲的執行期位址空間
`segment:offset`；1 MB 傾印與艦艇池位置使用線性位址  
狀態：呼叫鏈與槽索引來源 **proven**（原始位元組）；同盟六支的重算結果
**proven**（收據逐分隊比較）

## 結論

原版沒有一套寫死的「帝國隊形」與另一套「同盟隊形」。兩方開場形狀不同，
是因為每個艦艇記錄早已在 `+0x05` 帶著自己的 0–40 槽索引。

戰場初始化在 `297B:091D` 呼叫 `2BE1:0000(艦隊)`：

```
297B:091D  ff 76 16           push [bp+16]
297B:0920  53                 push bx
297B:0921  9a 00 00 e1 2b     call far 2BE1:0000
```

`2BE1:0000` 走艦隊 `+0x0F` 個編制槽，從 `+0x10` 起取得艦艇池索引，
再逐支呼叫 `2BE1:0059`。後者的原始定位與資料流是：

```
2BE1:0060  les bx,[bp+12]       ; 艦艇記錄
2BE1:0063  add bx,5
2BE1:0069  mov bl,es:[bx]       ; 艦艇 +0x05
2BE1:006E  shl bx,1
2BE1:0070  shl bx,1
2BE1:0072  mov ax,[bx+59BAh]    ; ds:59BA[槽].x
...
2BE1:008E  mov ax,es:[si+2]     ; ds:59BA[槽].y
2BE1:0095  les bx,[bp+0E]       ; 艦隊記錄
2BE1:0098  push es:[bx+4Ah]     ; 艦隊朝向
2BE1:00A5  call far 3A92:0001   ; 旋轉偏移
2BE1:00AF  mov ax,es:[bx+5Ch]   ; 艦隊 x
2BE1:00B3  add [bp+0],ax
2BE1:00B6  mov ax,es:[bx+5Eh]   ; 艦隊 y
2BE1:00BA  add [bp+2],ax
```

因此初始化與每回合歸隊使用的是同一張 `ds:59BA` 41 槽表、同一支旋轉函式；
[`126`](126-battle-formation-slots.md) 說「開場佈陣不是同一套」的舊結論已被
這條直接呼叫鏈推翻。

## 收據重算

`TestAllianceDeploymentMatchesOracle` 從 `s3ai6.bin` 的執行期艦隊表
（線性 `0x60BA4`）及 `ds:CFB4` 遠指標所指的艦艇池重建資料。它只取星域 26、
陣營 1 的六支同盟艦隊，以艦隊 `+0x5C/+0x5E`、`+0x4A` 和各分隊 `+0x05`
重算位置，再與艦艇 `+0x28/+0x2A` 逐支比較；六支全部相同。

remake 原先在讀取艦艇槽索引**之前**，先按分隊序號套十格帝國近似表；現改成
先讀 `+0x05`，再以 `FormationSlots.Station` 定位。這同時修掉同盟展開形狀，
也保留帝國各艦隊原有的槽位差異。
