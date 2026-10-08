# 334 — 戰略全滅之後轉跑 `ENDING.EXE` 是執行器造成的：BIOS ROM 尾巴是 0

> 2026-10-08 指標：下文「`ds:0000` 的 Borland 字串」實為 Datalight C 3.10 的橫幅（[`442`](442-compiler-is-datalight-c.md)），結論不變。

日期：2026-09-17。#57 的第一步（使用者裁決「先查執行器再做」）。
執行器：dosgolem-logh3 `b5485eb`（工作樹乾淨）。輸入：`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
調查由 subagent 做（筆記 `workplace/re-notes/a20-crash.md`，批次 `fcx-a20-boot`／`-cont`／`-entry`／`-rom`，反組譯
`workplace/ida/fcx-a20-2376.txt`、`fcx-a20-35D1.txt`；DOSBox 只讀 `~/cht/DOSBox-X-MCP-Debugger/dosbox-src/` 原始碼）。
主 session 核對了決定性的那一批 `fcx-a20-rom`（旗標、`report.txt`：改記憶體位址 `lin:FFFF0（FFFF0）16 bytes`、跑滿 11 億道、
主控台空、沒開 `ending.exe`；畫面 `9f32a4b4…`）。

證據等級：P ＝ proven、S ＝ strong inference、H ＝ hypothesis、U ＝ unknown。

## 結論

1. 崩潰的決定因素是 BIOS ROM 最後 16 bytes（線性 `FFFF0–FFFFF`）：**dosgolem 那裡全是 0**；DOSBox 寫的是
   `EA 5B E0 00 F0`（重置跳躍）＋ `"01/01/92"` ＋ `00 FC 55`（DOSBox-X `bios.cpp`）。P（`fcx-a20-rom`：只把那 16 bytes poke 成
   DOSBox 的值，回合照常走完、沒轉跑 `ENDING.EXE`；下一回合命令階段資料段與 `+0x6C` 清 0 的繞過批次 `fcx-turnrep-ann-clean`
   只差那 4 bytes）。
2. 不是 A20：崩潰那一刻 A20 是關的，GIN3 從開機到崩潰沒叫過 XMS、int 15h，也沒寫 `92`／`60`／`64` 埠（P，dosgolem 模型內，
   `fcx-a20-boot`／`-cont`／`-entry` 三批）。GIN3 用 EMS。
3. 在 DOSBox 語意下不會轉跑 `ENDING.EXE`（S）：A20 關時迴圈跑完、只寫壞 int 1–5 的向量；A20 開（DOSBox-X 預設 `dos=high`）
   讀到 HMA 的 0，迴圈很長但一格都不寫。js-dos 用的是 vanilla 還是 DOSBox-X 後端：U。

## 一、崩潰的路徑

`297B:034A(艦隊)` 把 `+0x6C..+0x6F`（沒進過戰場的艦隊都是 `FFFFFFFF`，S：只看了艦隊 0）當 far 指標交給 `35D1:0102`：

```
35D1:0102(p)   2376:0354(p) → 2376:01E6；p[+23] == 0 → 2376:0364(p) → 2376:02DE；在 ds:EAE0 清單找 p 刪掉
2376:01E6      存舊可見旗標與舊矩形（p[+1E]→[+23]、[+1F]w→[+24]w、[+21]w→[+26]w），算新矩形
2376:02DE      AH ← p[+2]；BX ← p[+24]w；DX ← p[+26]w → 2376:00B2
2376:00B2      字元平面 ds:D81E（40×20×6）從 BX 對應的格起，DL 列 × DH 格：
               ([si] & 0Fh) == AH → [si] |= 80h、[si+1] |= 01h；si += 6
```

（`docs/re/333` 引用 `workplace/re-notes/turn-reports.md` 寫的 `2376:03E6`／`04DE` 是 IDA 名字的位移，真正目標是 `01E6`／`02DE`，P。）

`ES:SI = FFFF:FFFF` 時，16 位元位移 `d` 的線性位址是 `FFFF0 + d − 1`：`d ≤ 0x10` 落在 ROM 尾巴（與 A20 無關），`d ≥ 0x11` 落在
1 MB 之上（A20 關 → 繞回 IVT 開頭，開 → HMA）。dosgolem 上（逐道軌跡與崩潰前傾印一致，P）：

- `p[+2]` 在 `FFFF1`：dosgolem 是 `00` → `AH = 0`；DOSBox 是 `5B`。
- `p[+24]`／`p[+26]` 在 IVT：讀到 int 3 的段 `348A`、int 4 的位移 `0010` → 迴圈 65,536 列 × 16 格。
- `AH = 0` 時「低 4 位元是 0」的格子一再被寫，掃過整個資料段，把 `ds:8F08`（計時器 ISR 轉給舊 int 08h 的向量，原本 `0080:0420`）
  改成 `0180:05A0`；#1,049,705,019 計時器跳進去（那裡全是 0），一路走到載入器、`EXEC ENDING.EXE`（P，ip-log）。
  主控台的「NULL function pointer」是 `ds:0000` 的 Borland 字串被同一個迴圈改了 bit7（P）。
- `AH = 5Bh` 時 `[si] & 0Fh` 不可能等於它，一格都不寫（S）。

## 二、對 dosgolem 的建議（沒有實作）

1. 開機把 `FFFF0–FFFFF` 填成 PC/AT 的重置進入點、BIOS 日期與機型位元組（`FC`），`F0000–FFFFF` 唯讀（寫入忽略）；`FFFFF = 55`
   是 DOSBox 自己的簽章，規格要寫明以原版執行環境為準的取捨。唯讀化之前先在既有收據集普查有沒有程式寫那一段。
2. IVT int 1–4 的預設向量：dosgolem 是 `0080:<int×4>` stub，DOSBox 是 IHSEG 的 IRET——這次只影響迴圈長度，先列觀察。
3. `.state` 沒存 `a20`／`hma`／`hmaOwned`／`a20Local`（潛在問題，這次沒觸發，P）。
4. 改了 ROM 之後要整條收據重跑（CLAUDE.md「收據綁執行器版本」）；受影響的至少有 `fcmd-occ-phz-t4`。

## 三、影響

- `fcmd-occ-phz-t4` 是崩潰殘留（`docs/re/333` 附節）；執行器修好之後重錄，才能做全滅的同狀態收據（#57）。
- 原版 `297B:034A` 拿 `+0x6C != 0` 判斷要不要釋放，而沒進過戰場的艦隊一律 `FFFFFFFF`：在 DOSBox 語意下「不崩潰」是因為 ROM 尾巴恰好
  讓迴圈不寫；remake 的全滅照「不崩潰」那一路（`fcx-turnrep-ann-clean` 的結果）是合理預設，但仍要照 #57 的裁決流程確認。
