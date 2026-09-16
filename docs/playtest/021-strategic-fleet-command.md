# 戰略艦隊命令同狀態對拍（清單點列 → 艦隊視窗 → 作戰 → 五項 → 確認）

日期：2026-09-15（第一片）、2026-09-16（駐留／輸送、實行、待命、占領、停用旗標）

比較等級：`same-state`（整張 640×480）。
原版全部是 dosgolem-logh3 `b5485eb`、帝國新遊戲：三十九批是正常玩家路徑、沒有
注入——冷啟動走 `dec1` 的序列（主選單 → 防拷 → 出擊命令 → 萊因哈特 → 出擊 →
確認 → 決定）到戰略地圖存檢查點，之後每一張從上一張的檢查點只多一個輸入
（`tools/fleetcmd_receipts.sh`；每批附 `plane.bin` 字元平面與 `fleet0.bin`
艦隊表第 0 筆）。停用旗標那七批（`fcmd-battle-*`、`fcmd-flag1-*`）用 `-poke`
造局面，注入的位址、步數與值寫在 `docs/re/315` §三，remake 在同一個時點改
同一個欄位。remake 走同一條路徑（`internal/game/fleetcmd_test.go`）。

`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；
四十六份 `pal.bin` SHA-256 都是
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
| `fcmd-self-done` | （從 `fcmd-self`）左鍵 (432,223) 決定 ＝ 取消命令，視窗留著 | `c22d858f12fd3ec3be1b3fbb5c3b62e4cb717f36aa92cefe7ae5410cba27a65a` | 0 |
| `fcmd-reopen` | （從 `fcmd-done`）左鍵 (450,268) 再開視窗 | `a157a513a9e6b7ae83bc4e53da77d6a628a38397d3c3588758a3c7a6fc694910` | 0 |
| `fcmd-exec` | （從 `fcmd-done`）左鍵 (544,85)、(560,85) 系統指令 → 實行 | `9c30cb446a7600e834788fe7bffb5744363b187ee7bcf9d9f53f643bad9d47ed` | 0 |
| `fcmd-garrison-pick` | （從 `fcmd-menu`）左鍵 (497,312) 駐留 | `86bb4cff7f888bceef49b365c7076d1834e35a142ff722ff5a72a5b45ebf1be0` | 0 |
| `fcmd-garrison-confirm` | 左鍵 (85,88) 亞爾提那 | `6fe0423806b87481de9f76e1eed42fd8751c0cabdeb8411e1b3e018835e7a738` | 0 |
| `fcmd-garrison-done` | 左鍵 (432,223) 決定（`+0x08` ＝ 2）| `fbe887b460912f91d2dba661e440fe9ca117d80f751e91f8e3bd27c7e8114fe5` | 0 |
| `fcmd-transport-pick` | （從 `fcmd-menu`）左鍵 (497,332) 輸送 | `eec881481fec049292a675696a24dae6723817887767f3cb6b50f1e6ce3be2fe` | 0 |
| `fcmd-transport-confirm` | 左鍵 (85,88) | `0c5bbd6b490fe75dd6296328f1b432ce6175931e1b4cf03d69dc681d29307476` | 0 |
| `fcmd-transport-done` | 左鍵 (432,223) 決定（`+0x08` ＝ 3）| `8fdb8323370b6972d0b32aea2d93c9db64648ff3a681b69883d5bb8264363852` | 0 |
| `fcmd-standby` | （從 `fcmd-menu`）左鍵 (497,352) 待命 → 指定待命行星 | `38cbde6d878320e4631516c1af029735dca4b966e29ed4b3f2b64ab2ec7c543a` | 0（排除小圖最後一列 128 px）|
| `fcmd-standby-pick` | 左鍵 (350,210) 第 1 顆 布連娜 | `3115f499e1ae0c279f5e3ad2a60458bb507f0cbef7a628af6a93e3bd5a25cf85` | 0（同上）|
| `fcmd-standby-done` | 左鍵 (400,310) 決定（`+0x07` ＝ 1、`+0x08` ＝ 5）| `f8b2665cd8cfe59160b94e7aac9b0a6dd2a1bafa95daed2f7709e85d85df7964` | 0 |
| `fcmd-standby-cancel` | （從 `fcmd-standby-pick`）左鍵 (350,310) 中止 | `052181c6ddc843357d1c4591fab5742bc9820260d333c599ea21e77bc4fb1284` | 0 |
| `fcmd-standby-rclick` | （從 `fcmd-standby-pick`）右鍵 (300,200) ＝ 決定 | `492e4e2dcb9097d606fa6b34b8ef31ac6169a6184215e8689dfe9f5d8b2588d5` | 0 |
| `fcmd-standby-reopen` | （從 `fcmd-standby-done`）左鍵 (450,268) 再開視窗 | `0b6f63b9331fecc0a7a94c4e0a649139b49b2bafec68ccd3ee6742e2da8c7690` | 0 |
| `fcmd-occupy-pick` | （從 `fcmd-menu`）左鍵 (497,292) 占領 | `64d5e94a1478c5f19bc4a205e092e4382ee2c977199ce172644813c8752845f7` | 0 |
| `fcmd-occupy-far` | 左鍵 (85,88) 亞爾提那（自己的 → 無法前往）| `572b6dacdd17b957921f865a1514fbfe5fb6b3974f8394eab76535b21bbf404c` | 0 |
| `fcmd-occupy-confirm` | （從 `fcmd-occupy-pick`）左鍵 (193,205) 亞斯提 | `ef7fc9fdc2b93c21aa7e4836a387d4e5864d297583278639593564567b646e8e` | 0 |
| `fcmd-occupy-done` | 左鍵 (432,223) 決定（`+0x07` ＝ 16、`+0x08` ＝ 1）| `6c600c4710a58ad0baa2357e6142c70b41fd5be88892445888f53420b75a3a4d` | 0 |
| `fcmd-occupy-reopen` | 左鍵 (450,268) 再開視窗 | `b834405151d53a81492911c05527eee31632b1e0bbc8250a6570ca576841251e` | 0 |
| `fcmd-occupy-phezzan` | （從 `fcmd-occupy-pick`）左鍵 (163,96) 費沙 → 參謀提醒第一頁 | `976cf5c51f631a86912f2228a2210ed991cf124e9d8d8c1d2ea8a9bdec517d28` | 0 |
| `fcmd-occupy-phezzan2` | 左鍵 (300,200) → 第二頁 | `cd3ab2b618c70beffc3adfeb6f7a51d91deeb531044412f2d1aba2025731a3c9` | 0 |
| `fcmd-occupy-phezzan3` | 左鍵 (300,200) → 「要占領中立軍費沙星系嗎？」| `ae7c80be023e67489e991247770ad7a1b85e5123518418eea7fff235f4d06a9c` | 0 |
| `fcmd-occupy-phezzan-done` | 左鍵 (432,223) 決定（`+0x07` ＝ 14、`+0x08` ＝ 1）| `1da3bb2ff0f5fbe6b5df3c760f27fa934b1b136ced2c711e2ca0acb500cb77fd` | 0 |
| `fcmd-battle-base` | 冷啟動同 `fcmd-base`，170M 步 `-poke lin:5F2DF=04`（首都 `+0x13` 有戰鬥）| `ca0ef22910190914ebdd2a60c26ecfb3522fc1035648f223c2a2db900ec6719c` | 0 |
| `fcmd-battle-win` | 左鍵 (450,268)：「戰鬥中」、沒有「變更」鈕 | `9fd2e8666cfbedfa23553178fa4cf8b01e97da3d073eca5bea34aeb30016fab5` | 0 |
| `fcmd-battle-menu` | 左鍵 (490,305) 作戰：四項停用、游標在停用項沒反白 | `ef703823a3636e51c0219f6474b12a881ecbfc369663b86a0169e4353d0e2c03` | 0 |
| `fcmd-battle-menu-hover` | 游標移到 (497,272) 移動 | `96c28c7f025ba5d35fd4dcc2f5ff105aba3667121d2e38c2393c1458954a8981` | 0 |
| `fcmd-battle-menu-click` | （從 `fcmd-battle-menu`）左鍵 (497,312) 停用的駐留 → 只關選單 | `14f9ea7d0138ef8971ffb3e74a8406d425db37aeb98ef32265235094b9d4ad6c` | 0 |
| `fcmd-flag1-menu` | （從 `fcmd-win`）446M 步 `-poke lin:60BA4=01`（艦隊 `+0x00` bit0）；左鍵 (490,305) 作戰 | `f1b0028007c68c6ddc2d1a9b154af900322a5c70bf54b6ec4025572c08c3a569` | 0 |
| `fcmd-flag1-menu-click` | 左鍵 (497,292) 停用的占領 → 只關選單 | `6c2888185595234094c54fc7d5773589f7e75602205f15e831cf5d9634a2f1d7` | 0 |

記憶體：`fcmd-done/fleet0.bin` 與 `fcmd-base/fleet0.bin` 只差 `+0x07`（FF→02）
與 `+0x08`（05→00）；remake 的 `Fleet.Target`／`Move`／`Parked` 逐一對上
（`TestFleetMoveOrderWritesFleet`）。星域 0 的 `+0x28` 在 `fcmd-base` 就是 1
（探索批次 `fcx-peek-base`：`-load-state fcmd-base.state -steps 421000000
-peek lin:5F2F4:4`）。

`fcmd-exec` 那一回合的兩條規則（軍事費流進行星、有自軍艦隊的星域換色表）與
清單記號、取消命令留著視窗、駐留／輸送的寫法，見 `docs/re/313`；「指定待命行星」
那張視窗（小圖、引線、只重畫變過的格子）見 `docs/re/314`；占領的允許表、費沙
兩頁提醒、選單停用旗標、不畫「變更」與左下角「處於戰鬥狀態」見 `docs/re/315`。
`fcmd-standby`／`fcmd-standby-pick` 排除的那一列（畫面 y 347、x 152–279）是原版
小圖的第 128 列：來源列 102 超出 102 列的地圖，讀到緩衝區後面的東西
（規格 051 已知差異）；其餘 307,072 個像素相同。

自動驗收：`internal/game` 的 `TestFleetWindowMatchesOracle`、
`TestFleetActionMenuMatchesOracle`、`TestFleetWindowClosesMatchOracle`、
`TestFleetTargetPickMatchesOracle`、`TestFleetConfirmMatchesOracle`、
`TestFleetMoveOrderWritesFleet`、`TestFleetConfirmCancelMatchesOracle`、
`TestFleetHoverHomeCountsNewFleet`、`TestFleetDecideSetsAdmiralFleetSlot`、
`TestFleetReopenShowsTarget`、`TestFleetCancelOrderKeepsWindow`、
`TestFleetGarrisonAndTransportOrders`、`TestFleetOrderThenExecuteMatchesOracle`、
`TestFleetStandbyMatchesOracle`、`TestFleetStandbyDecideWritesPlanet`、
`TestFleetStandbyCancelAndRightClick`、`TestFleetOccupyOrderMatchesOracle`、
`TestFleetOccupyPhezzanWarningMatchesOracle`、`TestFleetMenuDisabledInBattleRegion`、
`TestFleetMenuDisabledByFleetFlag`；
顯示層 `tools/gui_check.sh` 的 `fcmd-win`、`fcmd-done`、`fcmd-exec`、
`fcmd-occupy-phezzan3`（都整張）、`fcmd-standby-pick`（排除那 128 個像素）。

過程中改到的通用規則（不只這條路徑）：底線裝飾（`ui.DrawCellDecor`）、
清單被點那一列的反白、左下角與右側面板「寫進去那一刻的文字」與
「游標底下那一格變了才重寫」（規格 001 規則 4–7）、星域記錄艦隊數的重數
時機（`rules.RecountRegionFleets`）、玩家「決定」建的艦隊補上編制。
全套 `go test ./...`（除 GUI 套件）與既有 454 份收據的測試仍綠。
