# 041 — 第一階段來源、結果面板與士氣等待

日期：2026-10-02；工作項[#79](https://github.com/wicanr2/logh3/issues/79)。
固定身份、地址空間、注入、命令及失敗訂正：[研究369](../re/369-battle-preparation-lifecycle.md)。
事實：[欄位JSON](../internals/data/battle-preparation-lifecycle.json)；
索引：[十五項分級](../internals/data/battle-preparation-lifecycle-semantics.json)。

比較範圍是**原版局部診斷及注入局面的正常控制流重播**。
沒有remake正式畫面、差分圖或同狀態驗收，不能標整步same-state通過。

- 六案來源診斷兩次完整收據相同：
  `f1a95e05dbdc0aec055c10ae07de89ae93ea1a76f1b4d8a6d09b59c280d82c27`。
- 六組第一階段收據各兩次完整1MiB／畫面／色盤相同；全部事前固定種子`0x12345678`，
  入口、面板與終點種子bytes均維持`78563412`。
- 五個IDA函式／十五項分級合併核對；1930匯出SHA-256
  `7a84337a807478716a98a55881d43e204ebede4cfb9f1115fc40b22a9d0b32f9`，
  1AAB匯出SHA-256 `402d6a2cd17f162e2298219e834ec7fa43f5498894d20448d6df63f5df9edf32`。
  分級索引SHA-256 `b988d451bb15e1671d06d13299e085521640e383fa3ec318c1a5f6adf6f38981`。

| 收據模式 | 輸出完整記憶體SHA-256 |
|---|---|
| supply-panel | `ddfd3878bab7fce38a6e74ab09cf4859080980826a03d5c976376c171e123a22` |
| supply-end | `9a203e42181c66b72edfc4e1742dec920b05cd6a1b234e837f5b2f052c45689e` |
| repair-panel | `7dd546b48bd1f5873fd08bf88af75a099c22798ac18497f3a9be504c096599d8` |
| repair-end | `baa40c3836ef10a2c754fc1430b85d7ef006224ad9be0b570e8c68c67afb2d25` |
| morale-panel | `d11614131ec47e3b7c8d44737f6dc403ffc9d73508bf8eb90908b944cbbc9372` |
| morale-end | `7d5d2174ed239a43c8f89fcd8d10ae0d5e16917d0ad8061a9f125b472d299107` |

已目視原生640×480補給／修理面板；來源庫存、目標艘數及紅色缺額與欄位一致。
士氣恢復100、bit2報告期間保留、報告關閉才清旗標並回復防衛力。
來源零值只限固定執行器；池槽重用案例是明示診斷注入，不宣稱配置器已對拍。

`bash tools/gui.sh test -count=1 ./internal/... ./cmd/...`標準18個套件通過或無測試；
本機`workplace/battle79-life-product-tests.log` SHA-256
`6b23e17a1a612963f61bdaad463ebac1ff7f168006c790987291bc4f00091328`。
產品測試只證明既有remake回歸，本輪未新增GUI／存讀或AI／音樂實機聲明。
完整第一階段仍DRAFT，換戰場／重新開始／讀檔重設與正式接線待驗，#79保持開啟。

正式欄位SHA-256 `f8d4848a3e19232a66f92b18c2477647ea375a36379d2b5c792e3f5f48f1e650`，
另指定新輸出重生逐位元組相同。錯誤HEAD、舊錯誤堆疊收據、缺勘誤入口、既有輸出
四項拒絕護欄通過，原始收據皆未覆寫。
#79正文回讀相同且OPEN；新快照87張／11張open，267條核實零不成立、3條人工、119條一致、145條文件。
來源／輸出UID/GID1000:1000；工作樹未發現root-owned或同名.md目錄。
本輪容器皆一次性--rm結束，無本專案執行中或停止殘留；未重建image或清理其他專案。
