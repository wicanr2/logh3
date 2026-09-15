# 戰略艦隊命令同狀態對拍（清單點列 → 艦隊視窗 → 作戰 → 移動 → 確認）

日期：2026-09-15

比較等級：`same-state`（整張 640×480；`fcmd-done` 排除清單第 0 列 (336,300) 288×20）。
原版全部是 dosgolem-logh3 `b5485eb`、帝國新遊戲的正常玩家路徑，沒有注入：
冷啟動走 `dec1` 的序列（主選單 → 防拷 → 出擊命令 → 萊因哈特 → 出擊 → 確認 →
決定）到戰略地圖存檢查點，之後每一張從上一張的檢查點只多一個輸入
（`tools/fleetcmd_receipts.sh`；每批附 `plane.bin` 字元平面與 `fleet0.bin`
艦隊表第 0 筆）。remake 走同一條路徑（`internal/game/fleetcmd_test.go`）。

`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；
十五份 `pal.bin` SHA-256 都是
`28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8`。
座標是滑鼠座標（畫面 y ＝ 滑鼠 y ＋ 40）。

| 批次 | 多的那一步 | 畫面 SHA-256 | 差異 |
|---|---|---|---:|
| `fcmd-base` | （＝ `dec1`）| `3c8e6d5a149ba65457b3367aafa3373935c14ab33be4ad8a59da5ea9a3326434` | 0 |
| `fcmd-win` | 左鍵 (450,268) 清單第 0 列 | `0b29ecd9ed10a83ebd2a20a2576e661f55f8a5d95426c87c3e2fbe9a40138a29` | 0 |
| `fcmd-menu` | 左鍵 (490,305) 作戰 | `fd89cc88117b3d204d03dda9641fb81071a0d859313e30a20eccfcfb30729d1d` | 0 |
| `fcmd-menu-hover` | 游標移到 (497,272) | `65c34dbe1b853fe0d52c5295659cde0fcc8456b17ad1e16a90b04fd89c0bf79a` | 0 |
| `fcmd-pick` | 左鍵 (497,272) 移動 | `65942142b8c70a67b6499a19d78b9b60db0989cd99c97d4773cb48b77f078406` | 0 |
| `fcmd-confirm` | 左鍵 (85,88) 亞爾提那 | `c4070fae01487a9b10b3d7387d4b0423166c2f8439ff97164aa09ec3526d423a` | 0 |
| `fcmd-done` | 左鍵 (432,223) 決定 | `8875adb56b58ece719dfec553279d4537b110e0cc56b42b0e38d1f640e2f2c2a` | 0（排除清單第 0 列）|
| `fcmd-cancel` | 左鍵 (383,223) 中止 | `2603248e33142d7fe2d8a2570b2e1a06f7d34fb79ab9a3e07cf2b15e04370f30` | 0 |
| `fcmd-far` | 左鍵 (300,270) 巴拉特 | `acf2a7719e05b1e20b0097ca5a745eed6de1ec9a78ad8025bb9146249b6b06a0` | 0 |
| `fcmd-self` | 左鍵 (32,32) 瓦爾哈拉 | `f74d52c79d50acb1c523a5e1bef0a270c673226606e7db9060a51a1e1517cf3f` | 0 |
| `fcmd-pick-rclick` | 右鍵 (300,200) | `fa23a0ceb923de55215b957174614d2ce08f96c6dbfc345715418b7c709e0fd4` | 0 |
| `fcmd-menu-rclick` | 右鍵 (300,200) | `0adf02d86c0d8b54d48ced4af0fa2ab5467b2954650027d46670e6a5d8e493b6` | 0 |
| `fcmd-win-rclick` | 右鍵 (300,200) | `492e4e2dcb9097d606fa6b34b8ef31ac6169a6184215e8689dfe9f5d8b2588d5` | 0 |
| `fcmd-close` | 左鍵 (490,350) 結束 | `fb618854519c2548470ea2da1ff47f34424a0aefd2f16d89cbe217fdac577f66` | 0 |
| `fcmd-hover-home` | （從 `fcmd-base`）游標移到 (32,32) | `305f4de0b79b7a1b91c61720b78222cbb21af0f35dd5c41c465f6e341335ee69` | 0 |

記憶體：`fcmd-done/fleet0.bin` 與 `fcmd-base/fleet0.bin` 只差 `+0x07`（FF→02）
與 `+0x08`（05→00）；remake 的 `Fleet.Target`／`Move`／`Parked` 逐一對上
（`TestFleetMoveOrderWritesFleet`）。星域 0 的 `+0x28` 在 `fcmd-base` 就是 1
（探索批次 `fcx-peek-base`：`-load-state fcmd-base.state -steps 421000000
-peek lin:5F2F4:4`）。

`fcmd-done` 清單第 0 列左端多出來的移動圖示是規格 051 列出的已知差異
（那張圖與規則還沒查），那一列以外整張相同。

自動驗收：`internal/game` 的 `TestFleetWindowMatchesOracle`、
`TestFleetActionMenuMatchesOracle`、`TestFleetWindowClosesMatchOracle`、
`TestFleetTargetPickMatchesOracle`、`TestFleetConfirmMatchesOracle`、
`TestFleetMoveOrderWritesFleet`、`TestFleetConfirmCancelMatchesOracle`、
`TestFleetHoverHomeCountsNewFleet`；顯示層 `tools/gui_check.sh` 的
`fcmd-win`（整張）與 `fcmd-done`（排除同一列）。

過程中改到的通用規則（不只這條路徑）：底線裝飾（`ui.DrawCellDecor`）、
清單被點那一列的反白、左下角與右側面板「寫進去那一刻的文字」與
「游標底下那一格變了才重寫」（規格 001 規則 4–7）、星域記錄艦隊數的重數
時機（`rules.RecountRegionFleets`）、玩家「決定」建的艦隊補上編制。
全套 `go test ./...`（除 GUI 套件）與既有 454 份收據的測試仍綠。
