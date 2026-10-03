# 059：艦隊交戰下方場景

2026-10-04，沿#48／#79。[研究407](../re/407-fleet-blocks.md)、[095](../spec/095-fleet-blocks.md)。
比較等級：same-state，只限原版已投影操作的下方合成器。
上方圖庫、文字與量表直接借原版入口畫面，不列remake Game、正常玩家路徑或整場092通過。

## 固定條件

bundle SHA-256 de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c／clean，Go1.24.13。
父提交ee5561916311c8445a2434fa6d831bf9d88a73ed，工作樹含本輪095實作；
Go image golang:1.24-bookworm、GUI image logh3-ebiten:2。
固定父提交state及場景入口state的雜湊見407與凍結收據，不稱未受控的原版正常局面。
IRQ0=20000、pinned=true、cycle_clock=false；640×480，無裁切、縮放、遮罩或允許差異區。
兩側seed依每個邊界原值比較，兩次傷害前後可不同；沒有為通過而重擲。

## 重播與比較

每個輸出目錄必須尚不存在，a／b使用同一份a父state。

```bash
bash tools/fleet_blocks_oracle.sh a
bash tools/fleet_blocks_oracle.sh b
bash tools/fleet_blocks_oracle.sh registers-a
bash tools/fleet_blocks_oracle.sh registers-b
bash tools/fleet_blocks_oracle.sh edges-valid-a
bash tools/fleet_blocks_oracle.sh edges-valid-b
bash tools/py.sh /work/tools/scripts/fleet_blocks_receipt.py
bash tools/py.sh /work/tools/scripts/fleet_blocks_probe.py
bash tools/gui.sh xrun env LOGH3_BLOCKS_OUTPUT=/src/workplace/blocks48-r70-remake \
  go test ./internal/screen -run TestFleetBlocks -count=1 -v
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
```

remake輸出目錄先在容器內建立，核1000:1000。完整重建沿凍結收據中的祖先命令，
不是把DOSBox圖換名；dosgolem沒有改版，舊收據不因此失效。
可丟棄probe與正式解析器／合成器分開，期望像素直接取原版凍結資料。
每層由原版入口及累積typed操作重建，不把上一層期望畫面當下一層輸入。

| 驗證 | 數量 | 結果 |
|---|---:|---|
| 原版整場續接雙跑 | 40層／164檔 | RAM、色號、色盤、PNG逐byte相同 |
| VGA觀測雙跑 | 56邊界 | 原版GC、序列器、stack與完整末端screen hash相同 |
| 裁切直接入口雙跑 | 20案／20全圖 | 相同，seed保持，注入全部記錄 |
| 正式合成器 | 40層＋20案 | 每張307200像素零差，下方139776像素零差 |
| 未支援狀態／錯誤素材 | 指定回歸 | 回error，不改畫布 |

下方區域為x8..631、y208..431，包含整個背景、堆疊、標籤、亮標及箭頭。
圖塊本身裁切x8..615、y208..429，幾何由正式screen函式單一來源提供。
全部60張最大通道差0、差異bbox空；沒有未核准像素差異。
上方借圖的零差不能作完整renderer通過聲明。

原版／remake／diff各60張PNG與60份色號、comparison.json存
workplace/blocks48-r70-remake/。完整傷害前scene圖已目視，RGB使用原版AC映射色盤。
comparison.json SHA-256 ddc7536c4d8173fff016f78151f20adf56ecc4a3babc1b3ad2ff9cdd4a3bc149；
361份檔案的sha256.json SHA-256 f4605ac3a34525c181f1f0f12c08376453b800e87d191daeccec5176422f7a59。
凍結390份來源的JSON SHA-256
20874578ab46987b007f746efa0e180ff0906e5cc8c23ca3afdd0f1eadc9afeb。

兩個未返回的直接注入另外保留，不算成功裁切：X624超可見域；Y376、131塊的零高度
玩家可達性未知。095回error，092接線前仍需確認；不以可返回樣本覆蓋這個限制。
正常Game／GUI、文字typed轉接、正式色盤、等待輸入與保存待092；不以局部圖替代正常玩家驗收。

標準18產品套件13PASS／5無測試，正常結束；log SHA-256
0b0aa663dc968483094e14aabe3e42f57422dee460dc37af554d2e7e1e850ca2。
指定合成器／產圖log SHA-256
1be842b58b17e799c24eaacc09e9b59db5f9c496937b585db77584831930f1b7。
