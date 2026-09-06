# 對拍收據 001 — 主選單（含反白）

比較等級：**`same-state`**（同一份原版 bundle、同一條輸入序列、同一個畫面狀態）

結果：**逐位元組一致**。307,200 個像素全同，16 色調色盤全同；
色號陣列與調色盤檔案的 SHA-256 兩邊相同。

## metadata

| 項目 | 值 |
|---|---|
| 原版 bundle | `org_game/银河英雄传说3SP.zip`，SHA-256 `de905aa0…9673` |
| 啟動程式 | `GIN3.COM`，SHA-256 `770ed526…d918` |
| 字型輸入 | `GRAPH.IMG`，SHA-256 `0006671aca0e216da01d817be7f6a9e2fe8f18d5d058674bacf7f32c19ec833d` |
| 執行器 | dosgolem-logh3 `4c94d1a`（branch `logh3-com-support`，工作樹乾淨）|
| 建置 image | `golang:1.24-bookworm` |
| 顯示模式 | VGA mode 12h，640×480×16 |
| 畫面 | 主選單（帝國／同盟／對戰／取進度），第 3 列反白 |

## 原版側（oracle）

```
cd ~/cht/logh3
tools/oracle.sh hilite -steps 60000000 -tick 20000 \
  -click-x 320 -click-y 200 -click-at 15000000
```

輸入序列：第 15,000,000 道指令時在 (320,200) 點一下（跳過開頭動畫），
計時器每 20,000 道指令一次，跑滿 60,000,000 道指令。滑鼠停在 (320,200)，
遊戲因此把第 3 列標成反白。

可重現性：跑到 120,000,000 道指令的畫面與 60,000,000 道**逐位元組相同**，
所以這是穩定狀態不是過場幀。

- `screen.bin` SHA-256 `599ba871926d7574c4a0de416833024d78d6f6e2c35feef099e50d925ced05d4`
- `pal.bin` SHA-256 `28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8`
- 產物在 `workplace/oracle/hilite/`，`meta.txt` 記了執行器 HEAD 與完整旗標。

## remake 側

```
cd ~/cht/logh3
tools/go.sh run ./cmd/logh3 -font /orig/GRAPH.IMG -out workplace/remake
tools/go.sh run ./cmd/oradiff \
  -oracle workplace/oracle/hilite/screen.bin -remake workplace/remake/screen.bin \
  -oracle-pal workplace/oracle/hilite/pal.bin -remake-pal workplace/remake/pal.bin
```

- `screen.bin` SHA-256 `599ba871…05d4`（與原版側相同）
- `pal.bin` SHA-256 `28b933fe…02e8`（與原版側相同）

## 差異分類

未核准差異：**0**。

## 覆蓋範圍與限制

- 這張畫面是**靜態**的：反白列有驗到，但沒有驗證切換反白的過程、游標、
  鍵盤互動與動畫相位。
- 游標沒有出現在這個狀態（滑鼠停在選單上）。滑鼠停在選單外時遊戲會畫出
  自己的游標（實測多 71 個色號 7 的像素），那部分**還沒逆向也還沒實作**。
- remake 目前只畫這一個畫面，沒有正常玩家路徑（還沒有主迴圈與輸入）。
- 視窗外框的**角落規則沒有逆向出來**，目前靠七個量自 oracle 的像素修正
  （`docs/internals/07-window-frame.md`）。**換一個視窗尺寸必須重驗**。
- 文字用的是原版字型檔；remake 不內嵌它，路徑由旗標傳入，缺檔就 skip。

## 這份收據為什麼換過

原本的收據（執行器 `30fbbb0`、第 5,000,000 道指令點擊）在修正滑鼠虛擬座標
倍率之後**不再可重現**：跳過開頭動畫的那個點擊落點整個改變。修正本身是對的
（見 dosgolem `docs/spec/012` §2），失效的是綁在舊行為上的收據。

處置：收據改用 `tools/oracle.sh` 產生並記錄執行器 HEAD，規則寫進
`AGENTS.md`「固定輸入與 metadata」。

## 回歸

`internal/screen/mainmenu_oracle_test.go`（`TestMainMenuMatchesOracle`）
每次 `tools/go.sh test ./...` 都會重跑這次比較；缺字型或缺收據時 skip，
不用代用品。
