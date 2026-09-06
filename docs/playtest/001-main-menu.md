# 對拍收據 001 — 主選單

比較等級：**`same-state`**（同一份原版 bundle、同一條輸入序列、同一個畫面狀態）

結果：**逐位元組一致**。307,200 個像素全同，16 色調色盤全同；
色號陣列與調色盤檔案的 SHA-256 兩邊相同。

## metadata

| 項目 | 值 |
|---|---|
| 原版 bundle | `org_game/银河英雄传说3SP.zip`，SHA-256 `de905aa0…9673` |
| 啟動程式 | `GIN3.COM`，SHA-256 `770ed526…d918` |
| 字型輸入 | `GRAPH.IMG`，SHA-256 `0006671aca0e216da01d817be7f6a9e2fe8f18d5d058674bacf7f32c19ec833d` |
| 執行器 | dosgolem-logh3 `30fbbb0`（branch `logh3-com-support`）|
| 建置 image | `golang:1.24-bookworm` |
| 顯示模式 | VGA mode 12h，640×480×16 |
| 畫面 | 主選單（帝國／同盟／對戰／取進度）|

## 原版側（oracle）

```
cd ~/cht/dosgolem-logh3
DOSGOLEM_ORIG=~/cht/logh3/org_game/extracted tools/go.sh run ./cmd/probe \
  -exe /orig/GIN3.COM -root /orig -steps 60000000 -tick 20000 \
  -click-x 320 -click-y 200 -click-at 5000000 \
  -dump-ega /src/workplace/re06/screen.png \
  -dump-palette /src/workplace/re06/pal.bin
```

輸入序列：第 5,000,000 道指令時在 (320,200) 點一下（跳過開頭動畫），
計時器每 20,000 道指令一次，跑滿 60,000,000 道指令。

可重現性：同一命令連跑兩次，色號陣列與調色盤逐位元組一致
（`docs/re/06` §1）。

- `screen.bin` SHA-256 `25d9196a3b596cee06b623da0f7b15f1a5775a83987be30babc492a81431711c`
- `pal.bin` SHA-256 `28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8`

## remake 側

```
cd ~/cht/logh3
tools/go.sh run ./cmd/logh3 -font /orig/GRAPH.IMG -out workplace/remake
tools/go.sh run ./cmd/oradiff \
  -oracle workplace/re06/screen.bin -remake workplace/remake/screen.bin \
  -oracle-pal workplace/re06/pal.bin -remake-pal workplace/remake/pal.bin
```

- `screen.bin` SHA-256 `25d9196a…711c`（與原版側相同）
- `pal.bin` SHA-256 `28b933fe…02e8`（與原版側相同）

## 差異分類

未核准差異：**0**。

## 覆蓋範圍與限制

- 這張畫面是**靜態**的：沒有驗證反白、游標、鍵盤／滑鼠互動與動畫相位。
- remake 目前只畫這一個畫面，沒有正常玩家路徑（還沒有主迴圈與輸入）。
- 視窗外框的**角落規則沒有逆向出來**，目前靠七個量自 oracle 的像素修正
  （`docs/internals/07-window-frame.md`）。**換一個視窗尺寸必須重驗**。
- 文字用的是原版字型檔；remake 不內嵌它，路徑由旗標傳入，缺檔就 skip。

## 回歸

`internal/screen/mainmenu_oracle_test.go`（`TestMainMenuMatchesOracle`）
每次 `tools/go.sh test ./...` 都會重跑這次比較；缺字型或缺收據時 skip，
不用代用品。
