# 131：提督編輯器完整版與推廣影片

本輪已完成本機交付，狀態以[143](../goal/2026-10-19-admiral-delivery-round-143.md)為準。功能依[149](../spec/149-admiral-editor.md)與[驗收130](130-admiral-editor.md)，版本`v.1.5.0-20261019`。三平台七包、正常GUI、影片及README截圖的身份保存於[唯一交付索引](../internals/data/v150-delivery.json)。舊版本、遠端tag及Release保持。

## 平台驗證

| 範圍 | 結果 |
|---|---|
| 凍結來源 | 1590項含vendor，612項非vendor與目前產品逐檔一致；產品來源仍與r6標準19套件15 PASS／4無測試相同 |
| Windows amd64 | 完整／引擎ZIP與來源tar逐檔通過；Wine八條直接執行／啟動器／匯入／缺資料入口，包含中文空格路徑 |
| Linux amd64 | 完整／引擎ZIP逐檔及執行權限通過；11條入口涵蓋兩種AI、六表存讀、續回合與配樂載入 |
| macOS通用 | amd64及arm64，最低11.0、SDK15.5、僅系統依賴；ARM3253頁ad-hoc簽章全部相符；ZIP逐檔及權限通過，未真機驗收 |
| 正常GUI | 實際Linux及Windows完整ZIP解壓後，各兩次正常片頭至S1；兩陣營編輯、無效武器1拒絕、取消／套用、F7及重啟讀回通過，各11張完整960×720截圖 |

Linux正常GUI與錄影在`workplace/admiral143/editor-linux/`。Windows GUI在`workplace/music69-windows/windows143-editor-r2/`，使用eob既有Wine映像，前次缺xdotool失敗目錄保留。這次驗收沒有注入局面或seed，也沒有`-script`；Linux11入口及Windows8入口的診斷樣本另列，不能互相升格。

## 本機交付

| 平台 | 完整包 |
|---|---|
| Windows amd64 | [ZIP](../../dist-all/v.1.5.0-20261019/full/LOGH3-v.1.5.0-20261019-windows-amd64.zip) |
| Linux amd64 | [ZIP](../../dist-all/v.1.5.0-20261019/full/LOGH3-v.1.5.0-20261019-linux-amd64.zip) |
| macOS Intel／Apple Silicon | [通用ZIP](../../dist-all/v.1.5.0-20261019/full/LOGH3-v.1.5.0-20261019-macos-universal.zip) |

[七包SHA-256](../../dist-all/v.1.5.0-20261019/SHA256SUMS-platforms.json)包含三平台完整／引擎及來源包。引擎包候選留在同版本`release/`，未上傳GitHub；本機完整包仍限合法持有人使用。

## 推廣影片與README

[影片](../../dist-all/v.1.5.0-20261019/promo/LOGH3-v.1.5.0-20261019-promo.mp4)長101.016667秒，37段、1920×1080、60fps、AAC雙聲道。新版提督編輯器與三武器完整交火為正式ZIP新錄，零跳格新交火27.9667秒；另以明列「前版實戰回顧」的v1.4已驗場次補足，完整艦隊交火合計51.2833秒。要塞7.6秒亦沿既有完整場次，動畫與海報沿核准來源，不宣稱整片新拍。

全片僅BGM_0F一首連續底軌，無重啟或循環，70個遊戲音效依正式played觀測獨立重混。36個剪接點及全片PCM比對通過，AAC SNR33.96dB、峰值0.59574，沒有超幅或黑段。凍結區間按原速來源對照，分類為標題卡、原遊戲等待或低變動；不把正常停留當作遺失畫格。81個字卡／字幕安全矩形、來源時碼抽樣、CONTACT及原尺寸編輯器／要塞／片尾均已核對。聲畫收據與FFprobe保存於同一`promo/`，以[影片校驗清單](../../dist-all/v.1.5.0-20261019/promo/SHA256SUMS.json)回查。

同盟約621秒來源的全片掃描曾兩次碰到舊固定120秒時限。回查路由與實際輸入後，[驗證器](../../tools/scripts/battle_playback_audit.py)將工作時限改為`max(120, ceil(片長／2))`，並固定解碼2執行緒及1濾鏡執行緒；相同來源與黑段／凍結／逐格條件乾淨重跑通過。原始失敗日誌保留，沒有修改遊戲、來源畫格或放寬內容門檻。

README更新的[19張正式包實錄](../images/readme-v150.json)含兩陣營設定、基本／內部／四特殊能力、兩名飛行專家、AI選單、星圖、劇本與三武器；原尺寸960×720，未重畫或裁切。來源ZIP、執行中ELF、GUI收據、原始影片時碼及圖像SHA-256皆由[匯出器](../../tools/scripts/admiral_delivery.py)核對。三項未知原版欄名及新增規則的證據界線沿149。

## 重建入口

沿`tools/gui.sh`的Docker環境執行，使用既有Go模組快取離線建立vendor；來源與素材先核對雜湊。輸出沿專案`dist-all/<版本>/{full,release,promo,smoke}/`既有結構，詳見[協作規範](../../AGENTS.md)。

```sh
tools/gui.sh xrun python3 tools/scripts/windows_hotfix_package.py --version v.1.5.0-20261019 --base-version v.1.4.0-20261015
tools/gui.sh xrun python3 tools/scripts/v140_platform_delivery.py prepare --version v.1.5.0-20261019 --batch release143
tools/gui.sh xrun python3 tools/scripts/v140_platform_delivery.py linux --version v.1.5.0-20261019 --batch release143
bash tools/macos.sh macos-release143 bash /src/tools/scripts/local_test_macos_build.sh v.1.5.0-20261019 workplace/package92-v.1.5.0-20261019-release143/source
tools/gui.sh xrun python3 tools/scripts/v140_platform_delivery.py macos --version v.1.5.0-20261019 --batch release143
tools/gui.sh xrun python3 tools/scripts/v140_platform_delivery.py finalize --version v.1.5.0-20261019 --batch release143
```

## 驗證邊界

三平台完整與引擎包均驗內容、版號、來源與ZIP完整性。Linux正常X11、Windows Wine及macOS雙架構／依賴／ARM簽章分開記錄；不把Wine或靜態檢查升格真機。診斷入口與正常玩家操作分開。影片使用真實正式包畫面及既有原曲，保持全片單首連續音樂，不混入會逐剪接重啟的來源配樂。

完整版與影片含原版及衍生素材，只留本機。未授權公開發布，沒有新平台真機、人耳或DOS整段同態聲明。
