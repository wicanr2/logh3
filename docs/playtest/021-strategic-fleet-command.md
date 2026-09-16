# 戰略艦隊命令同狀態對拍（清單點列 → 艦隊視窗 → 作戰 → 五項 → 確認）

日期：2026-09-15（第一片）、2026-09-16（駐留／輸送、實行、待命、占領、停用旗標、變更、艦隊組成編輯器、副官、右下角清單的分頁與捲動、星系情報視窗、艦隊走到之後）

比較等級：`same-state`（整張 640×480）。
原版全部是 dosgolem-logh3 `b5485eb`、帝國新遊戲：一百四十六批是正常玩家路徑、沒有
注入——冷啟動走 `dec1` 的序列（主選單 → 防拷 → 出擊命令 → 萊因哈特 → 出擊 →
確認 → 決定）到戰略地圖存檢查點，之後每一張從上一張的檢查點只多一個輸入
（`tools/fleetcmd_receipts.sh`；每批附 `plane.bin` 字元平面與 `fleet0.bin`
艦隊表第 0 筆）。停用旗標那七批（`fcmd-battle-*`、`fcmd-flag1-*`）用 `-poke`
造局面，注入的位址、步數與值寫在 `docs/re/315` §三；艦隊組成編輯器的
`fcmd-edit-poor*` 五批把運補艦存量 poke 成 0（`docs/re/317` §九）；remake 在
同一個時點改同一個欄位。remake 走同一條路徑（`internal/game/fleetcmd_test.go`、
`fleetedit_test.go`）。

`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；
一百五十八份 `pal.bin` SHA-256 都是
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
| `fcmd-change-menu` | （從 `fcmd-win`）左鍵 (490,325) 變更 → 副官／編組 | `ef930a43722cd909a7b78f844b4131a63ac513af20c0432da31f07fb1a640d84` | 0 |
| `fcmd-change-menu-hover` | 游標移到 (497,352) 編組 | `5076b5d2cfac178e9ff1ceb1eabf9d865b21c3a481d14af5d9de914e88e0550b` | 0 |
| `fcmd-change-menu-rclick` | （從 `fcmd-change-menu`）右鍵 (300,200) | `0adf02d86c0d8b54d48ced4af0fa2ab5467b2954650027d46670e6a5d8e493b6` | 0 |
| `fcmd-change-adj` | （從 `fcmd-change-menu`）左鍵 (497,332) 副官 → 艦隊編成視窗的變體 | `5954644f1a34b3bd4b66967ac3decd73078cc637c179ccd8ced41f7fb6c5aac6` | 0 |
| `fcmd-change-adj-close` | 左鍵 (480,300) 結束 → 回視窗、清單那一列不反白 | `cc68e5e842ecc076fc187c8f85b2526345a6ecedead05501d9ae8c17613ebf56` | 0 |
| `fcmd-change-adj-rclick` | （從 `fcmd-change-adj`）右鍵 (300,200) | `e6368b493b5cf533d4752c0a8b91493d0e642b4aa5e2e89b6c6cb10d048b9619` | 0 |
| `fcmd-change-org` | （從 `fcmd-change-menu`）左鍵 (497,352) 編組 → 這一支的艦隊組成 | `06219d971be8b3a6e43a4d8c5d6cb56dc0fbc46d9de0b1a97f868d697ce61449` | 0 |
| `fcmd-change-org-cancel` | 左鍵 (510,280) 中止 | `1327506be7002eac56052b3363e76e3176390d87088b19cba63e7e2020f9d4c4` | 0 |
| `fcmd-change-org-decide` | （從 `fcmd-change-org`）左鍵 (510,300) 決定 | `25f4ea051d0e227f7a30a343607d2dfc4dd4a1f50880cb0b6522351c51d50b82` | 0 |
| `fcmd-change-org-rclick` | （從 `fcmd-change-org`）右鍵 (300,200) → 不動 | `015473d94ca5e1966f4a902b05b6fa44cc238afc453e9287679a2eba656b290b` | 0 |
| `fcmd-fok-cancel` | 冷啟動同 `fcmd-base`，艦隊組成按中止 (510,280) → 回地圖、存量 78000 | `284f41b4d9872ad3c873c3febbb9703096d9ef94801eb98be83e9ea9b83dc09f` | 0 |
| `fcmd-fok-rclick` | 冷啟動同上，艦隊組成按右鍵 → 不動 | `7fea9d30203f0f052457bf5a5c12e86f48002fe00b5539e3e7f3ce639d5a0298` | 0 |
| `fcmd-edit-place1` | （從 `fcmd-change-org`）左鍵 (176,200) 槽 19 放戰艦 | `9d73928a392583cdd565d35fa041969221a207648b2ee068997c3488beefc4a2` | 0 |
| `fcmd-edit-row2` | 左鍵 (360,170) 巡航艦列 | `9d96001b90d918d194cbb97306527db67ed6e1988af7717b66c20ce45b744871` | 0 |
| `fcmd-edit-place2` | 左鍵 (240,200) 槽 21 放巡航艦 | `474b900fcdf37f051d3543d9a828fbae3980fe2a95185211cb5e11a378dd4600` | 0 |
| `fcmd-edit-replace` | （從 `fcmd-edit-place2`）左鍵 (176,200) 槽 19 換成巡航艦 | `41b4923f11f9b86c5833b8671ff7b3605f92830176f9f08f20ca19a36a822f07` | 0 |
| `fcmd-edit-remove` | （從 `fcmd-edit-place2`）右鍵 (176,200) 拆槽 19 | `764247354046d3ec879b4aa1c8329e1e51c8e1838f8896dd03a8f9061aacc33f` | 0 |
| `fcmd-edit-centre-msg` | （從 `fcmd-edit-place2`）左鍵 (208,200) 巡航艦放中央 → 訊息框 | `dcef66450939da92e294fdabc0e7807bdbb582268cebfba2a79ad41dea25bc41` | 0 |
| `fcmd-edit-centre-ok` | 左鍵 (432,223) 確認 | `2065261289159a5e5d5a611d0d67ec87ac9f248a12111591313e1112a0f82b8f` | 0 |
| `fcmd-edit-flag-rclick` | （從 `fcmd-edit-place2`）右鍵 (208,200) 中央 → 「無法撤離旗艦。」| `78ab08e24f20ebf502ed1fc3bfce66ddf72193472292affc05d3b821d3766427` | 0 |
| `fcmd-edit-cancel` | （從 `fcmd-edit-place2`）左鍵 (510,280) 中止 → 存量倒回 | `1327506be7002eac56052b3363e76e3176390d87088b19cba63e7e2020f9d4c4` | 0 |
| `fcmd-edit-decide` | （從 `fcmd-edit-place2`）左鍵 (510,300) 決定（`+0x0F` ＝ 3）| `8669e400742fc8c77d491de64a141f69a7b41177988bf4eeb12582f998c2cf6a` | 0 |
| `fcmd-edit-formation` | （從 `fcmd-edit-remove`）左鍵 (510,270) 陣形 | `6c7661b90d970b44d64cfbfcff480a2f5cf16eecd422c2721bbf1fe3ffe680ca` | 0 |
| `fcmd-edit-form-hover` | 游標移到 (505,350) 密集 | `c4710bca22436ab2a0975d3a2e0d9453d5e8dd57bca4e2f5e2e298b29a78bdb0` | 0 |
| `fcmd-edit-form-rclick` | （從 `fcmd-edit-formation`）右鍵 (300,200) | `23061a2af590bf70b05a32a22fb613a7cd5323630dc150b31da16de3797e4bea` | 0 |
| `fcmd-edit-form-pick` | （從 `fcmd-edit-formation`）左鍵 (505,210) 圓錐 → 8/8、81％ | `9fe58c6fb2994f5b949b86fd9bcde37dccdfc907e577f3348572a6188bc4fb62` | 0 |
| `fcmd-edit-full-msg` | 左鍵 (144,136) 槽 0 → 「無法超過…」| `47893b787168e13a540cf6a4ccf8fa07da065c84d9dcc097e31e417851b9a290` | 0 |
| `fcmd-edit-form-decide` | （從 `fcmd-edit-form-pick`）左鍵 (510,300) 決定（`+0x0F` ＝ 8）| `6f570476544b5913c0c7c95c592aee6c344b269e20d8cbd9cfd5e2a1c2fdf8b9` | 0 |
| `fcmd-edit-poor` | （從 `fcmd-change-menu`）471M 步 `-poke lin:5F310=00,00,00,00`（運補艦 0）；左鍵 (497,352) 編組 | `1c3a8296c24da680500fd0107c035f72cf5da79d54832903a287375156212f27` | 0 |
| `fcmd-edit-poor-row5` | 左鍵 (360,230) 運補艦列 | `9af313ffd78da0b6b7926c32e369cf33d0264ff22469cdc1a09829c14a941890` | 0 |
| `fcmd-edit-poor-msg` | 左鍵 (176,200) → 「船艦數量不足。」| `50420c27d3df5d430e1f949a39919a3004003dec2bde394d3453d7dc2f0da0b8` | 0 |
| `fcmd-edit-poor-form` | （從 `fcmd-edit-poor`）左鍵 (510,270) 陣形 | `3fce8834fd1bcd4b62e77fa99ef0371eaf720048ec33d36de5001aa327a23e6a` | 0 |
| `fcmd-edit-poor-pick` | 左鍵 (505,210) 圓錐 → (5,29) 換成修護艦 | `4d671bbed080d7400b324cfc78f4967576272322f9fdc5fc13a6c211e2401b4a` | 0 |
| `fcmd-fok` | 冷啟動同 `fcmd-base`，350M 截（艦隊組成剛開）| `767fb3660a681deb410882d1305beff0faa6039d99eb2a71710e3f59891252b3` | 0 |
| `fcmd-fok-place1` | 左鍵 (176,200) | `28b13e7e6fd81f15a74ef191ae7c509fa4aa0b503b97e7c2d2ba62c5f2f09faf` | 0 |
| `fcmd-fok-row2` | 左鍵 (360,170) | `5cb6ae1486b6a2ca1b3a8181f02bfa946530bcad41578d2dda8e2a2483f6d78c` | 0 |
| `fcmd-fok-place2` | 左鍵 (240,200) | `5a4337a027ad3f6288dc220a8ee0c72223f2ea4fb639afd9a6137adec0284a68` | 0 |
| `fcmd-fok-decide` | 左鍵 (510,300) 決定 → 地圖（`+0x0F` ＝ 3）| `b7a824f3153d94cab8d1434efdf618d9c8b727594fbf94dfbb3b5fc42eafc4ce` | 0 |
| `fcmd-fok-form` | （從 `fcmd-fok-row2`）左鍵 (510,270) 陣形 | `a29dc2dbdbf3162756b49abe83ba424e45b54d0ed1bc06815d99092a026c35fb` | 0 |
| `fcmd-fok-form-pick` | 左鍵 (505,210) 圓錐 | `4e44f62495e4560137250a1bbf9fee9d6179b9496b156f7c1e52ef61e12a1cd3` | 0 |
| `fcmd-adj-menu` | （從 `fcmd-change-adj`）左鍵 (430,300) 副官 → 變更／卸任 | `992c7d5d728f89d348b4327bc4b1377c8dfcd1790d4ecfc8ab002bdbb99dd0db` | 0 |
| `fcmd-adj-menu-hover` | 游標移到 (430,330) 卸任 | `3846cb00a461d31e02bd6077010f1167975f288bf09852a9e351557c0e89c279` | 0 |
| `fcmd-adj-menu-rclick` | （從 `fcmd-adj-menu`）右鍵 (300,200) | `0c4a19a70a470e30a6c841e72cfb3e7ad55dded98b4e19153a330af1731446fc` | 0 |
| `fcmd-adj-dismiss` | （從 `fcmd-adj-menu`）左鍵 (430,330) 卸任 | `3f60ad34533bdf53546ce0fdd18821abce9e9477342effd3e4b7f248095279f4` | 0 |
| `fcmd-adj-dismiss-close` | 左鍵 (480,300) 結束（艦隊 `+0x05` ＝ 0xFF、營運 100）| `24e8bac68504b2e294198c4f9a855ce25cf71230921f727673c0a8c93e3a1e97` | 0 |
| `fcmd-adj-change` | （從 `fcmd-adj-menu`）左鍵 (430,310) 變更 → 待命中副官資料 | `6748f8b04c66a0fb8d037cf47008e9cf28e42180151b333502dccc831c1cb1db` | 0 |
| `fcmd-adj-change-hover` | 游標移到 (450,150) | `dd1462f225f3071db44d9a91315f550064f33e6553be6f4e586bff8c46d7318a` | 0 |
| `fcmd-adj-change-rclick` | （從 `fcmd-adj-change`）右鍵 (300,200) → 原版清單留著、之後沒反應 | `1226bffc349b72a9a9e1aaa0beee1f03786dc283f0c2cbcc3b627b83ff90b652` | 不比（規格 013 已知差異）|
| `fcmd-adj-change-pick` | （從 `fcmd-adj-change`）左鍵 (450,70) 米達麥亞 → 詳細資料 | `2bb79ace513cef01ebef704a3ee569e566f3db3d831f18c82aa95d2cdb13a249` | 0 |
| `fcmd-adj-change-down` | （從 `fcmd-adj-change`）左鍵 (617,350) ▼ | `25febf16d6f2cc0762a1ef550a937f8ffdfc554838918689440eaa5611dcdbcf` | 0 |
| `fcmd-adj-change-down2` | 左鍵 (617,350) ▼（到底）| `4be3777def5805563995723d2d6c5842d486cd60945458e7b060c550f2c82a0e` | 0 |
| `fcmd-adj-change-up` | （從 `fcmd-adj-change-down`）左鍵 (617,50) ▲ | `eac58f1a4403f64eb9a90b40509672266a80a583b21717b34738c8301d23e85f` | 0 |
| `fcmd-adj-change-pick2` | （從 `fcmd-adj-change-down`）左鍵 (450,70) 瑞文茲 | `a7a345f7d333b9f69a319211548a326a8a996f48a6f14ff7c477a04635c58a67` | 0 |
| `fcmd-adj-pick-rclick` | （從 `fcmd-adj-change-pick`）右鍵 (300,200) → 回視窗不換 | `e86b3b2483162125eb8d468f11b4229d1ac85d4cabc94906b24910cce0c8e902` | 0 |
| `fcmd-adj-pick-cancel` | （從 `fcmd-adj-change-pick`）左鍵 (448,280) 中止 | `733061c4df33427d823ad5a48cfbe2b7762838555aeb8bf2388cf029a51ffac6` | 0 |
| `fcmd-adj-pick-confirm` | （從 `fcmd-adj-change-pick`）左鍵 (496,280) 出擊 ＝ 決定 | `2a227faa14d70af9512b2c37c9f7a8d61db824a0ca2c4e1db76bc34c175859b5` | 0 |
| `fcmd-adj-confirm-close` | 左鍵 (480,300) 結束（艦隊 `+0x05` ＝ 2、營運 102）| `eca31b764ff84958d9fd9a4df95e5d9a90033e71fa4b45a569a7f20c9506b5cc` | 0 |
| `fcmd-fadj` | 冷啟動 `fadj` 的序列，360M 截（艦隊編成按副官）| `1c9a944d192e828583459931b27ed1c811c1b3b929ef2adffd028ec3bafab454` | 0 |
| `fcmd-fadj-dismiss` | 左鍵 (430,330) 卸任 | `7558fdb7df8e0b9627931d5db6ef8865d63f6373d9b54848f576eddd48eea40e` | 0 |
| `fcmd-fadj-change` | （從 `fcmd-fadj`）左鍵 (430,310) 變更 | `328d349c5aca298b5e6acce3b7b2152eaef416b800efe20f4b4485a70063c2b2` | 0 |
| `fcmd-fadj-change-pick` | 左鍵 (450,70) | `d56711025b0e546bffb22ea5c7e8a1071ffecd86cc1aefe05f2ebef02d2979d3` | 0 |
| `fcmd-fadj-pick-confirm` | 左鍵 (496,280) | `42a69af008a6476f05e98e025c558362cb2934093b808bb52d275a6dc059cf27` | 0 |
| `fcmd-fadj-confirm-fleet` | 左鍵 (480,310) 確認、(510,310) 決定 → 地圖（艦隊 `+0x05` ＝ 2）| `3c8e6d5a149ba65457b3367aafa3373935c14ab33be4ad8a59da5ea9a3326434` | 0 |
| `fcmd-tab-planet` | （從 `fcmd-base`）左鍵 (415,250) 行星分頁 → 瓦爾哈拉四顆行星 | `01903bc07e685dcc739ded0459bb9ceb7e8ff4512b75928f301a65b048d12331` | 0 |
| `fcmd-tab-planet-row` | 左鍵 (450,268) 列 0 → 奧丁的行星情報視窗（列反白） | `6cd8f243d632eb0733278a12849eef40cad995e242c96590ec491367e8308673` | 0 |
| `fcmd-tab-planet-ok` | 左鍵 (462,289) 確認 → 關窗、反白還原 | `e5e1fd8efa46580d888e81fdf19bba0da6f99393120c529d6fb339e0b79540a6` | 0 |
| `fcmd-tab-planet-rclick` | （從 `fcmd-tab-planet-row`）右鍵 → 關窗 | `9082a46e674a6312009ae9e079eb80ba9c8badc2da97f870928c37dba3267c30` | 0 |
| `fcmd-tab-planet-scroll` | （從 `fcmd-tab-planet`）左鍵 (617,250) 欄 17 列 6 → 只有游標動 | `06013047e4bd2ae5951f2b047765f0d8889c1d923037c89e9910d3cf86ae1144` | 0 |
| `fcmd-tab-enemy` | （從 `fcmd-base`）左鍵 (490,250) 敵艦隊分頁 → 空頁 | `d8e4f9d44e57efc58149c921ac4e6d0dc5ada0cdcaed3f16cb56029e63d995cd` | 0 |
| `fcmd-tab-enemy-row` | 左鍵 (450,268) 空列 → 不動 | `841a878e045ca37535cad8a6e90009c022f27fcb962591e3dd97e54ed3cfecb0` | 0 |
| `fcmd-tab-admiral` | （從 `fcmd-base`）左鍵 (560,250) 提督分頁 → 35 位、▲▼ | `ae52019660673172e44de5fce149300f5d5afc8172c5c1588466241016fb2c50` | 0 |
| `fcmd-tab-admiral-down` | 左鍵 (617,348) ▼ 按住 0.3M 步 → 位移 5 | `edbd2bec6acd24323119e2463036fdace17dbf19f87b09e78e5c6c82cc212d44` | 0 |
| `fcmd-tab-admiral-down2` | 再一次 ▼ → 位移 10 | `54da800fa4212c6ba4587b26c4b9160204466008c5b5a9cb83294dfd8d6c9d05` | 0 |
| `fcmd-tab-admiral-up` | （從 `fcmd-tab-admiral-down`）左鍵 (617,268) ▲ → 位移 0 | `c69aab032dc8527306384da7733891d89d7f71fb41f43ca4f47643dccfe32159` | 0 |
| `fcmd-tab-admiral-end` | （從 `fcmd-tab-admiral`）▼ 按住 3M 步 → 到底、位移 30 | `8f2450590f344cb9e30075c5295a00d97cca1b06bcafd3f190b315dfde0f8d6e` | 0 |
| `fcmd-tab-admiral-row` | （從 `fcmd-tab-admiral`）左鍵 (450,268) 列 0 → 吉爾菲艾斯的詳細資料 | `3f8261509a1221a7aa9c7ce53b31574485dcf416d9597ff4c648d9df66344725` | 0 |
| `fcmd-tab-admiral-cancel` | 左鍵 (448,280) 中止 → 關窗、反白還原 | `9c606e8fc4aa107ddc6de757b2c47cb854b52bfe504bb3b2c2c7e4d1775ef8ad` | 0 |
| `fcmd-tab-admiral-rclick` | （從 `fcmd-tab-admiral-row`）右鍵 → 關窗、清單不動 | `f00e0276f84654f8ccfcee56b7e4141a9c14dc15fa848501a33e1291cb94e39f` | 0 |
| `fcmd-tab-admiral-sortie` | （從 `fcmd-tab-admiral-row`）左鍵 (496,280) 出擊 → 艦隊編成 | `13f30e0535a539895feb0e6acb8c6d2055fc93d2eed382443673e9eeab908132` | 0 |
| `fcmd-tab-sortie-rclick` | 右鍵 → 艦隊組成（出擊那條路的艦隊編成右鍵 ＝ 確認） | `8bfbe45a8027622dd05438186ba5f6e495c7cbad6084cb97eb1a8f85fb4ee2ae` | 0 |
| `fcmd-forg-rclick` | 冷啟動同 `fcmd-fok`，艦隊編成那一下換成右鍵 → 艦隊組成（雜湊 ＝ `fcmd-fok-rclick`） | `7fea9d30203f0f052457bf5a5c12e86f48002fe00b5539e3e7f3ce639d5a0298` | 0 |
| `fcmd-tab-fleet2` | （從 `fcmd-tab-admiral-sortie`）左鍵 (480,310) 確認、(510,310) 決定 → 第２艦隊；提督分頁還選著、少兩位 | `dab8c85922316bea893ea1d39179ff0f4c94bd75a676abec73f2d391ef0104fd` | 0 |
| `fcmd-tab-fleet2-list` | 左鍵 (350,250) 艦隊分頁 → 兩支 | `e52262bc8ad747db34b77158666ea9133d2e961602c23eca7cb94645149d9e8a` | 0 |
| `fcmd-tab-fleet6` | （從 `fcmd-tab-fleet2`）再出擊四位（列 0、出擊、確認、決定 × 4）→ 六支 | `25597c8ba654be0f45f7cb50559086ffbba530c6a534adfaee4b56ba99fa961e` | 0 |
| `fcmd-tab-fleet6-list` | 左鍵 (350,250) 艦隊分頁 → 六支、▲▼ | `34dbeb5fa11add3bde3afd6b966060929612f7ffca88c20cdc79701051b09df3` | 0 |
| `fcmd-tab-fleet6-down` | 左鍵 (617,348) ▼ → 第 2–6 艦隊、位移 1 | `fe80039ea5d459e42c319e91cdf7aa02ee891edc320619cefa7efe9b94b75e7c` | 0 |
| `fcmd-tab-fleet6-up` | 左鍵 (617,268) ▲ → 位移 0 | `e5bc4d16b3ba80a03dd1be488e3f4c399b542bc2faa0c5a9e466fae52c5a25d2` | 0 |
| `fcmd-tab-fleet6-row` | （從 `fcmd-tab-fleet6-down`）左鍵 (450,348) 列 4 → 第６艦隊的視窗 | `43f8f82c5404f860cf8b6c666c9e6feaa24229c964a045642056459c12be5a77` | 0 |
| `fcmd-tab-fleet6-close` | 左鍵 (490,350) 結束 → 地圖、位移仍 1 | `602d0a3aa64727b59bcedbbb922aa76c7605c9d8b099ec787133e5496c6dafd7` | 0 |
| `fcmd-rinfo` | （從 `fcmd-base`）左鍵 (32,32) 瓦爾哈拉（面板看的那一格）→ 星系情報視窗 | `24f4f7b8107457fe4875d5560a13cf34ad80eefd5a5af6163d1927d44905dba9` | 小圖最後一列 |
| `fcmd-rinfo-close` | 左鍵 (270,350) 結束 → 回地圖、左下角那塊空 | `55d1083fb4627488bb401a2caf921a89c8e643b4824c353c27cc930ee3cbcc62` | 0 |
| `fcmd-rinfo-rclick` | （從 `fcmd-rinfo`）右鍵 → 回地圖 | `492e4e2dcb9097d606fa6b34b8ef31ac6169a6184215e8689dfe9f5d8b2588d5` | 0 |
| `fcmd-rinfo-outside` | （從 `fcmd-rinfo`）左鍵 (100,100) → 視窗不動 | `a80c7282d9737b3b21599cf9e098f703309b980abe73cb602ca72970589f1b83` | 小圖最後一列 |
| `fcmd-rinfo-fezzan` | （從 `fcmd-base`）左鍵 (163,96) 費沙 → 換選：右側面板中立軍星系費沙、清單空 | `722c17e6d47cd4f219fb43c431213ebe3c78379d46a827c1bc3eae4e1f06a8cc` | 經濟力數字 |
| `fcmd-rinfo-fezzan2` | 再點費沙 → 視窗（破折號版） | `e652632554f2f6eb078d050b82fe6990cb7693ac22e60362407ee29fe2b60779` | 小圖最後一列、經濟力數字 |
| `fcmd-rinfo-ballat` | （從 `fcmd-base`）左鍵 (300,270) 巴拉特 → 換選：同盟軍星系、成長率／徵稅額／稅率 `----` | `3ae6d451f1b02ea513f7ca28ca63aecd26bc7c47af802104a17516f058bd003c` | 經濟力數字 |
| `fcmd-rinfo-ballat2` | 再點巴拉特 → 視窗：五顆行星（魯德米拉在要塞格） | `aab8ea9ed75bc373bdce1016443a3a7a2b0ac64b3ecfdabf6a20518f23867c09` | 小圖最後一列、經濟力數字 |
| `fcmd-rinfo-ballat2-close` | 左鍵 (270,350) 結束 → 回地圖、面板仍看巴拉特 | `cd04db95e2f03aa8e1fc4a3b347d9a34d66c479c6e94dd58482417c6980bdf4d` | 經濟力數字 |
| `fcmd-map-rclick` | （從 `fcmd-base`）**右鍵** (300,270) 巴拉特 → 什麼都不換（`ds:9588` 仍 00、面板與清單仍瓦爾哈拉），左下角跟著游標寫巴拉特 | `9e1111bcb6157795ce3b7e9b28cdf3805c640a81e0747fe40b6105bd3566cfa7` | 0 |
| `fcmd-occ-t1` | （從 `fcmd-occupy-done`，120M 步一回合）系統指令 (544,85)、實行 (560,85) → 艦隊 0 → 2 | `6a456b0bc7e4b0995f0dbd71cb7a68f6e69df2610b640efde420c5c68fc4fd61` | 0 |
| `fcmd-occ-t2` | 再實行 → 7 | `cf3e80c92368b3de99437a984e7ac381acacaa900ea7c34f5d4639784323a3f0` | 0 |
| `fcmd-occ-t3` | 再實行 → 8 | `acdbd29f87363898c851ed4ede8a50f0d86c1d5d9032803caf620389222e8555` | 不比（叛亂分岔） |
| `fcmd-occ-t4` | 再實行 → 13 | `c559c0c1ce7bf04d0459cc302ef7b28da798a46f797ec9f9adf4a48bc6adeb2e` | 不比 |
| `fcmd-occ-t5` | 再實行 → 16 亞斯提：停駐、bit4、占據報告、打行星（4200→3768、`+3C` 67） | `8f5855bc3f73975d2f0bf63c5bda9722b715aeb4d551482415ab1b718ce53e92` | 不比（只比記憶體） |
| `fcmd-occ-t5b` | 左鍵 (300,200) 關報告 | `bd9495843825a14760b72128bbf63da36c31d5465094687fc7f2785eb3228532` | 不比 |
| `fcmd-occ-t6` | 再實行：行星回 4200、`+3C` 48 | `37540979b158f06bb8f2abb42bb5ea24a33f38aeed74da99a8cda7144497570d` | 不比 |
| `fcmd-occ-phz-done` | （從 `fcmd-occupy-phezzan3`）決定 (432,223) → 占領費沙 | `1da3bb2ff0f5fbe6b5df3c760f27fa934b1b136ced2c711e2ca0acb500cb77fd` | 不比（只比記憶體） |
| `fcmd-occ-phz-t1` | 系統指令、實行 → 2 | `24fbd482ed93d44218882061bc1455128880d9223de46026b914c67580d1d025` | 不比 |
| `fcmd-occ-phz-t2` | 再實行 → 6 | `16f29f72098733fa10ddfe145775f1c1959fbb4cb9d47e3d1c9b7a44db2fea3d` | 不比 |
| `fcmd-occ-phz-t3` | 再實行 → 14 費沙：停駐、bit4、8800→8792、`+3C` 47、停在士氣報告 | `006288b49ad1e138d1556932f4a9ad984aac320cf7905c2e688b65ea7c8b38c9` | 不比 |
| `fcmd-occ-phz-t3b` | 左鍵 (300,200) 關報告 | `4a7b2f0b91b3e1a0caba34bb5742e73cc5807f15c326922061bde3e821230884` | 不比 |
| `fcmd-occ-phz-t4` | 再實行：8792 不動 | `4f7988030a00d082fe445e00a2ac5dab502300ff1b80e8592dd569867b60ef74` | 不比 |

記憶體：`fcmd-done/fleet0.bin` 與 `fcmd-base/fleet0.bin` 只差 `+0x07`（FF→02）
與 `+0x08`（05→00）；remake 的 `Fleet.Target`／`Move`／`Parked` 逐一對上
（`TestFleetMoveOrderWritesFleet`）。星域 0 的 `+0x28` 在 `fcmd-base` 就是 1
（探索批次 `fcx-peek-base`：`-load-state fcmd-base.state -steps 421000000
-peek lin:5F2F4:4`）。

`fcmd-exec` 那一回合的兩條規則（軍事費流進行星、有自軍艦隊的星域換色表）與
清單記號、取消命令留著視窗、駐留／輸送的寫法，見 `docs/re/313`；「指定待命行星」
那張視窗（小圖、引線、只重畫變過的格子）見 `docs/re/314`；占領的允許表、費沙
兩頁提醒、選單停用旗標、不畫「變更」與左下角「處於戰鬥狀態」見 `docs/re/315`；
「變更」的小選單、兩張子視窗與艦隊組成的中止／右鍵、待命艘數的扣法（開局 78000）
見 `docs/re/316`；艦隊組成編輯器（放、拆、陣形、決定寫回、每一步的存量）見
`docs/re/317`——二十九批各附 `-peek lin:5F2FC:28` 的七個存量與 `fleet0.bin`；
副官的變更／卸任（候選、寫回提督記錄的哪幾格、清單捲動與右鍵）見 `docs/re/318`
——二十三批各附提督 0／2／32 的記錄。
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
`TestFleetMenuDisabledByFleetFlag`、`TestFleetChangeMenuMatchesOracle`、
`TestFleetChangeAdjutantWindowMatchesOracle`、`TestFleetChangeOrganiseWindowMatchesOracle`、
`TestFleetCompositionCancelAndRightClick`；編輯器（`fleetedit_test.go`）
`TestFleetEditPlaceMatchesOracle`、`TestFleetEditReplaceMatchesOracle`、
`TestFleetEditRemoveMatchesOracle`、`TestFleetEditFlagshipKindMessageMatchesOracle`、
`TestFleetEditCancelAndDecideMatchOracle`、`TestFleetEditFormationMatchesOracle`、
`TestFleetEditPoorStockMatchesOracle`、`TestFleetEditSortiePathMatchesOracle`
（二十九批整張，另比七個存量與 `fleet0.bin` 的組數與每組的艦種／槽）；副官
（`fleetadj_test.go`）`TestFleetChangeAdjutantMenuMatchesOracle`、
`TestFleetAdjutantDismissMatchesOracle`、`TestFleetAdjutantChangeMatchesOracle`、
`TestFleetAdjutantSortiePathMatchesOracle`（二十二批整張，另比提督記錄的 +07／+09
與艦隊的副官／營運）；右下角清單的分頁（`listtabs_test.go`，規格 052）
`TestListTabPlanetMatchesOracle`、`TestListTabEnemyMatchesOracle`、
`TestListTabAdmiralMatchesOracle`、`TestListTabFleetScrollMatchesOracle`、
`TestFleetOrgRightClickIsConfirm`（二十六批整張，另比四個分頁的筆數與位移）；星系情報視窗
（`regioninfo_test.go`，規格 053）`TestRegionInfoMatchesOracle`、`TestRegionInfoOtherRegionsMatchOracle`、
`TestMapRightClickKeepsPanel`（十批，小圖最後一列與換看別家時的經濟力數字事先排除，理由在規格 053）；
艦隊走到之後（`occupation_test.go`，規格 054）`TestOccupyAstarteArrivalMatchesOracle`（t1／t2 整張，t1–t5 艦隊記錄與星域旗標）、
`TestOccupyPhezzanArrivalMatchesOracle`、`TestPlanStepBranches`（十三批 `fcmd-occ-*`，每批另附艦隊 0／星域表／行星表的 `-peek`）；
顯示層 `tools/gui_check.sh` 的 `fcmd-win`、`fcmd-done`、`fcmd-exec`、
`fcmd-occupy-phezzan3`、`fcmd-change-org`、`fcmd-edit-full-msg`、`fcmd-fok-decide`、
`fcmd-adj-confirm-close`、`fcmd-tab-planet-row`、`fcmd-tab-admiral-down`、
`fcmd-tab-fleet2`、`fcmd-occ-t2`（都整張）、`fcmd-standby-pick`、`fcmd-rinfo`（各排除小圖最後一列那 128 個像素）。

過程中改到的通用規則（不只這條路徑）：底線裝飾（`ui.DrawCellDecor`）、
清單被點那一列的反白、左下角與右側面板「寫進去那一刻的文字」與
「游標底下那一格變了才重寫」（規格 001 規則 4–7）、星域記錄艦隊數的重數
時機（`rules.RecountRegionFleets`）、玩家「決定」建的艦隊補上編制。
全套 `go test ./...`（除 GUI 套件）與既有 454 份收據的測試仍綠。
