# RE 07 — 中文字型解出來了：`GRAPH.IMG` 就是倚天字型

日期：2026-09-06
執行器：dosgolem-logh3 分支 `logh3-com-support`（`30fbbb0`）
輸入：原版未修改的 `org_game/extracted/`（唯讀）
結果文件：[`docs/internals/06-font-graph-img.md`](../internals/06-font-graph-img.md)
＋ [`data/font-graph-img.json`](../internals/data/font-graph-img.json)
前篇：`06-palette-alive-first-real-screen.md`

推論標籤：confirmed／強證據／假說／未知。

---

## 1. 結果（confirmed）

`GRAPH.IMG`（419,850 bytes）是完整的中文點陣字型：半形 8×15、全形 16×15，
每列 2 bytes 高位在前。索引公式、區段版面與存取方式全部釘住，見結果文件。

**內容是倚天中文系統的字型，逐位元組相同**：漢字區 13,094 個字形與
`stdfont.15` 100% 相同，半形區 3,840 bytes 與 `ascfont.15` 100% 相同。

## 2. 方法：拿畫面當已知答案去反查

主選單的九個字我們**知道字碼也知道長相**（`docs/re/06` 的 framebuffer），
所以不必猜格式：把點陣抽出來，在候選檔案裡找同一串 bytes，找到就同時得到
格式、索引與來源。

先量尺寸再找：字是色號 7、墨水框 15×14、欄距 16、列距 20——由投影量出來，
不是目測。

## 3. 三個「找不到」與它們的正對照

這一輪有三次查詢回空，每一次都不是結論，是查法的問題。

**其一：16 列。** 第一版假設字形 16 列去搜，全空。倚天的 `ascfont.15`
是 3,840 bytes ＝ 256 × 15 —— **15 列**。改成 15 列九個字全中。
量到的墨水高度 14 列本來就與 15 相容、與 16 也相容，光看畫面分不出來；
是外部參照給了答案。

**其二：`IBM16.350`。** 檔名含 16、又是整包最大的檔（1.76 MB），
是「看起來最像字型」的候選——搜下去零命中。真正的字型是名字完全看不出來的
`GRAPH.IMG`。**檔名是二手推論，位元組是一手證據。**

**其三：記憶體與 EMS 都搜不到。** 字形畫在畫面上，1 MB 位址空間與全部
EMS 頁卻找不到任何一個字的點陣。根因不是搜錯，是它**從來沒有以整塊形式
存在過**：開機只讀開頭 1,920 bytes，之後每畫一個字才 seek 過去讀 30 bytes。

## 4. 機制層級的確認（confirmed）

為了問「它到底讀了哪些位置」，dosgolem 加了檔案存取記錄（`FileOps`，
記檔名與**偏移**）。跑到主選單，九次 seek 的偏移與從畫面反推的公式
**完全相同**：

```
#25777538 seek Graph.img → 0xF780   read 30 bytes   ← 帝（公式 0x00F780）
#25778643 seek Graph.img → 0x1564E  read 30 bytes   ← 國（公式 0x01564E）
#25781661 seek Graph.img → 0x9042   read 30 bytes   ← 同（公式 0x009042）
…（九個字全中）
```

點陣對得上、機制對得上、外部參照對得上——三條獨立證據同一個答案。

## 5. dosgolem 改動（`30fbbb0`）

- `DOS.FileOps`：記 AH=3Fh／AH=42h 的檔名與偏移；probe 報告新增「檔案存取」。
- `DOS.EMSPages()`：列舉全部 EMS 邏輯頁；`-find` 一併掃過去並報 handle／頁號。
- `-find` 吃逗號分隔的多組樣式（`50f418c`）。
- 單元測試 1 支（seek 後 read 的 Pos／Len）。

## 6. 還沒解出來的

- `0x4632`–`0x6996` 那 302 個字形（假名／西里爾／框線／圈號）的字碼對應。
  已排除「以 Big5 `0xC6A1` 為起點連續存放」。測法：跑到結局畫面
  （`ENDING.EXE` 會顯示 `德間ジャパン`），看它 seek 到哪個偏移。
- `IBM16.350`（1.76 MB，整包最大）的用途。
- 其他資料檔的格式。

## 7. 重現

```
cd ~/cht/logh3
LOGH3_EXTRA_RO=~/cht/etan_font tools/py.sh tools/scripts/font_match.py \
  /ref/stdfont.15 /ref/ascfont.15 /orig/IBM16.350        # 反查來源
tools/py.sh tools/scripts/font_extract.py                 # 解格式 ＋ 出 JSON/PNG
tools/py.sh tools/scripts/font_regions.py                 # 各區段點陣圖

cd ~/cht/dosgolem-logh3
DOSGOLEM_ORIG=~/cht/logh3/org_game/extracted tools/go.sh run ./cmd/probe \
  -exe /orig/GIN3.COM -root /orig -steps 26000000 -tick 20000 \
  -click-x 320 -click-y 200 -click-at 5000000            # 看「檔案存取」段
```

產物（gitignore）：`workplace/assets/font/` 的點陣圖集。
