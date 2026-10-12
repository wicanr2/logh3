# v.1.0.0-20261010 正式封包與按鈕驗收

已完成按鈕修正、三平台引擎包與本機完整版、66秒推廣片。[128](../spec/128-visible-button-activation.md)限定remake接線CONFORMED；[123](../spec/123-hd-complete-presentation.md)限定82人物、空格與邊緣細化CONFORMED。唯一來源與輸出身份見[本次索引](../internals/data/v1-release.json)。

## 交付

根目錄：`dist-all/v.1.0.0-20261010/`，沿專案AGENTS指定的`full/`、`release/`、`promo/`結構。封包名稱均為`LOGH3-v.1.0.0-20261010-<平台>.zip`。

| 目錄 | 平台及內容 | 使用方式與權利 |
|---|---|---|
| `release/` | Windows amd64、Linux amd64、macOS Intel／Apple Silicon通用包；另附引擎來源 | 玩家提供自己合法持有的PC DOS中文版原版資料，按README複製到`original/`或用`-orig`指定。無原版、HD、配樂及音色庫 |
| `full/` | 同三平台完整資料、82人物及空格、22首配樂、28段音效、固定音色庫；另附引擎來源 | 解壓後用`start-original`或`start-enhanced`啟動，兩種AI共用存檔。含原版與衍生素材，只留本機 |
| `promo/` | `LOGH3-v.1.0.0-20261010-promo.mp4`、分鏡、來源、影音檢查與抽樣圖 | 含原版衍生音畫，只留本機 |
| `smoke/` | 四條正常視窗重播、Linux／Windows抽驗及標準測試 | CLI、真實X11、Wine與macOS靜態驗證分開記錄 |

根目錄另有`BUILD.json`、`FINAL-VERIFICATION.json`及`SHA256SUMS.json`。SHA清單涵蓋八份封包、影片、建置／驗收資料與HD提案試樣。引擎授權RRSAL-1.0；第三方授權隨包保存。

存檔位置：Windows為`%APPDATA%\logh3\save`，Linux為`$XDG_CONFIG_HOME/logh3/save`或`~/.config/logh3/save`，macOS為`~/Library/Application Support/logh3/save`。防拷畫面輸入138後，可點「確認」或按Enter。

## 按鈕根因與結果

- 第二畫面的「確認」已繪製，但`StateCopyProtection`缺少左鍵分支。現在使用實際`InputBox`矩形呼叫既有Enter送出流程。空答案與界外點擊保持無作用。
- 雙人會計及訂造有可見按鈕，卻未接左鍵。現在沿實際`overlay.Boxes`進入預算／建造，確認沿既有關報告路徑。
- 預算、建造及待命提督清單返回實際單人／雙人父畫面。金額、結算、AI、存檔格式及換手規則沿既有流程。
- 修正前`TestCopyProtectionVisibleConfirm`及`TestVersusVisibleMonthlyButtons`均失敗，修正後通過。既有單人七列箭頭、百分比及存讀回歸通過。

## 最小充分驗收

| 驗收 | 結果與聲明範圍 |
|---|---|
| 標準測試 | `go test ./internal/... ./cmd/... -count=1`，19套件、14PASS、5無測試 |
| Linux正常視窗 | 實際新包從片頭與主選單進入，實體X11鍵鼠送138並點確認；帝國HD 1.5倍32張、同盟HD 2倍32張、雙人HD 1.5倍33張、原圖1.5倍32張，共129張、94080000比較像素零差。保存完整原始視窗，原圖模式另按既有方法正規化到640×480；沒有排除區域 |
| Linux完整版CLI | 11條公開入口，含存讀、兩種AI命令差異及音訊入口。CLI抽驗與上列正常視窗各自保存，不互相代替 |
| Windows完整版 | 解壓實際ZIP、逐檔manifest、版本與UTF-8 BOM／CRLF、Wine六條入口，含HD人物及戰術畫面；不稱Windows真機 |
| 引擎啟動器 | Windows Wine及Linux各四條：兩種AI的主選單與「點確認→財政變更」。使用外部原版資料，發行目錄沒有複製原版素材 |
| macOS | 通用Mach-O含amd64及arm64、最低11.0、系統依賴及ARM ad-hoc每頁hash通過。引擎及完整版程式同hash；不稱macOS真機、Developer ID簽署或公證 |
| 封包 | 六份ZIP全部CRC、檔案集合、逐檔大小及SHA-256通過；兩份引擎來源包同hash。1503項建置來源按工作樹／凍結vendor核對 |
| HD素材 | 真實FACE.AMG為83格，82人物及第82號單色空格；六張人物並排核對表已目視。其他素材33個首／中／尾樣本保持307200個邏輯中心與原色號，不稱逐張AI重繪 |

正常視窗比對對象是同版正式Game模型，證明remake內部繪製與輸入結果一致。本輪沒有新DOS oracle收據，不升格為原版same-state或全遊戲parity。

## 推廣片

66秒、1920×1080、30fps、1980幀，H.264視訊及AAC音訊。畫面取自本次正式包的正常X11與公開CLI抽驗，分鏡逐項記來源hash；原圖／HD比較是remake兩種顯示模式。配樂取本包`BGM_0F.ogg`，固定增益14dB，平均−23.8dB、峰值−8.3dB，沒有超過兩秒的靜音或0.25秒黑段。十張抽樣幀已目視，字幕安全矩形無裁切。

第1至8段刻意使用靜態畫面，第0及9段有界平移，首尾幀hash不同。凍結檢測原始輸出保留，不把靜態剪輯宣稱為連續實況。較前版改以財政、造艦與完整人物展示安排敘事，保留本作星圖與指揮介面的色彩。影片來源與權利見`promo/PLAN.json`及`RIGHTS.txt`。

## 重現入口與環境

全部建置與分析在Docker，使用既有映像：GUI為`logh3-ebiten:2`、macOS為`psychicwar-osxcross:go1.24.13-15.5-r2`、Wine為`kol-wine-verify:ubuntu-noble-20261007-r1`、影音為`eob-audio-capture:20261009-r3`。Go1.24.13，原始資料唯讀、非root寫入、網路關閉，有資源及逾時上限。

完整命令、來源與輸出hash在本次索引。以下是流水線入口，重新產生時使用新的乾淨工作區，腳本會拒絕覆寫既有封包。

1. `tools/scripts/local_test_package.py --formal`建立Windows與Linux完整包，`--library`及`--soundfont`指定本次已驗證音訊輸入。
2. `tools/macos.sh`及`local_test_macos_build.sh`建立兩架構；`local_test_macos_package.py`封裝與驗章。
3. `tools/scripts/v1_release.py`從完整包抽出引擎，`v1_release_smoke.py`驗啟動器與自備資料。
4. `hd_art_model.go`建立Game預期；`player_input_display.sh`、`hd_art_gui.py`與擷取器重播正常X11，存TRACE、原圖與差分。
5. `v1_promo.py`依序執行`prepare`、`cards`、`encode`、`audit`。由`tools/audio.sh`掛載Noto CJK字型目錄；明示字型避免montage預設字型失敗。
6. `v1_finalize.py`逐檔核對並集中交付身份。

本次從含使用者既有修改的工作樹凍結來源建置，記錄HEAD及每檔hash，沒有擅自提交其他decomp工作。未建立tag或GitHub Release。

## HD後續與仍有限制

[第122輪HD提案](../goal/2026-10-10-v1-and-hd-plan-round-122.md)按實際圖像資源安排去重、財政試樣、字型／介面、艦艇／地圖、動畫及三平台驗收。根目錄`HD-PLAN.md`為交付副本，`HD-PREVIEW.html`可切換目前4:3實圖及16:9草圖；新方向尚待使用者選擇，未進正式程式。

完整登陸分段演出、人耳、Windows／macOS真機與原有音訊循環驗收仍依#48、#69、#66保留；#16原版欄位證據未知也保持。v1.0.0為本次功能發行，不改寫這些證據限制。前版封包、影片與原始oracle不覆寫。
