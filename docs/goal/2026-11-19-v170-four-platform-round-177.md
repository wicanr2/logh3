# 第177輪：v.1.7.0 四平台完整版與推廣片

2026-11-19。使用者要求「打包完整版，四平台與新的推廣影片」。開工前裁決四項：

- 版號 v.1.7.0-20261119。
- Android 先補 HD 圖再打包。
- 建一把本機發行金鑰。
- 推廣片沿用 v1.6.4 故事片加入 Android 段，只放本機。

| 項目 | 狀態 | 證據 |
|---|---|---|
| Android 接 HD 呈現器 | PASS，人物、星圖、介面主題、字型；不收海報 | `df086664`，模擬器截圖 |
| 封包腳本 | PASS，`tools/scripts/v170_full_release.py` 沿 v1.6.5 素材換程式並收 APK | `INPUTS.json` |
| 桌面三平台 | PASS，建置、ZIP 逐檔、AppImage 解包、macOS 雙架構 | [交付147](../playtest/147-v170-full-delivery.md) |
| Android 發行版 | PASS，本機發行金鑰簽章，素材與共用素材逐檔核對 | `ANDROID.json`、`APKSIGNER.txt` |
| 新遊戲兩陣營整場 | PASS，AppImage、Wine、行動版殼存檔逐位元組相同，也與 v1.6.5 相同 | `RESULT.json`、`mobile-e2e/` |
| 前線駐留、參謀更換 | PASS，同盟參謀第一次因腳本跟不上未走到兩入口，重跑通過 | `garrison-linux/`、`staff-linux-2/` |
| 模擬器 | PASS，戰略版面、戰場捲動、跨重開存讀檔 | `emu-saveload2/`、`promo177/emu-rec2/` |
| 推廣片 | PASS，字卡算法與剪接格先以舊版號驗證，再換新版號 | `promo/AUDIT.json` |
| 標準測試 | 17 通過、6 無測試、0 失敗 | `standard-tests.txt` |

踩過的坑：

- 行動版殼的狀態切換原本不寫 log，模擬器腳本只能照秒數操作。release 版第一次開啟比 debug 版慢，整段錯拍。改成殼印「logh3: 狀態 StateXxx」，腳本等狀態再點（`clearlog`、`waitlog`）。
- 模擬器腳本的觸控座標落在列邊界外 1 像素（裝置 146 換算回原版 121，列從 122 起），存讀檔看似沒反應。改點列中央。
- release 建置預設跑 lintVital，要另外下載 lint 套件，離線建置失敗；只有一支 Activity，關掉。
- BGM_0F 只有 127.1 秒，舊片已用到 126.9 秒，不能延長；Android 段改用該畫面在遊戲中實際的 BGM_3F。

下一步：實機驗收（需要使用者的手機）、release 版匯入流程、觸控板模式、M2。
