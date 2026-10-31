# v.1.6.4：P索敵、行軍與完整版推廣片

使用者要求三平台完整版，推廣影片補錄P索敵及行軍。本輪沿[158規格](../spec/158-scout-formation-lazy-hd.md)與[170已驗來源](141-scout-formation-lazy-hd.md)，凍結為`v.1.6.4-20261031`，不覆寫舊版。唯一狀態為[171](../goal/2026-10-31-full-p-scout-march-promo-round-171.md)。完整版及影片含原版與衍生音畫，只在本機交付。

## 三平台完整版

同一份凍結來源1663檔，652份共用檔。原始HD保持，83張發行人物圖由109880634縮為7392131 bytes，與原先runtime縮放後RGBA逐像素一致。字型、海報與其他素材維持來源SHA。各包附`MANIFEST.json`、版本、來源及操作說明；Windows附閃退日誌啟動檔。

| 平台 | `dist-all/v.1.6.4-20261031/full/`檔名 | bytes | SHA-256 |
|---|---|---:|---|
| Windows amd64 | `LOGH3-v.1.6.4-20261031-windows-amd64.zip` | 372492136 | `83a6324d338ae5c8296d34a7e299ada688cbe69ab3e70913e97512ab8e7d7fbb` |
| Linux x86_64 | `LOGH3-v.1.6.4-20261031-linux-x86_64.AppImage` | 372468216 | `4b7b3ab7b249effcb22a7c2ce1aacaa97e5ee17be6dc66b1c35aa5f23a5b8884` |
| macOS Intel／Apple Silicon | `LOGH3-v.1.6.4-20261031-macos-universal.zip` | 382386562 | `1301a49e7edf2897c23a7dff64cd7dcc7127b7de733b0755b8aec470c281fc79` |

Linux沿既有Type-2 runtime，944632 bytes前綴，SHA`1cc49bcf1e2ccd593c379adb17c9f85a36d619088296504de95b1d06215aebbf`。實際封包解出後核對檔案、權限與連結，等同建置AppDir，含15項共享函式庫。執行採`--appimage-extract-and-run`，素材自動尋址，未指定`-orig/-data/-art/-save`，存檔使用隔離XDG目錄。FUSE啟動未驗。

Windows為PE64 amd64，實際ZIP逐檔CRC、SHA、權限、非ASCII檔名及README BOM／CMD CRLF核對。macOS為雙架構通用app，最低11.0、SDK15.5，依賴與ARM ad-hoc簽章頁雜湊核對。macOS未作真機啟動，Wine不等於Windows真機。

## 正常玩家路徑

標準命令`go test ./cmd/... ./internal/...`在Docker完成21產品套件，16PASS、5無測試。以下都走正式封包正常入口，沒有遊戲狀態、seed或時間注入。

| 本機收據 | 驗證內容 |
|---|---|
| `workplace/release171/player-linux/RESULT.json` | AppImage同盟，三組、五語、五組名言完整300秒輪替、F1／F7、六表存讀及重啟 |
| `workplace/release171/player-empire/RESULT.json` | AppImage帝國，實際60秒名言切換、v2／v3、F1／F7、存讀及重啟 |
| `workplace/music69-windows/full171/player-windows/RESULT.json` | 實際Windows ZIP解出後Wine，同盟相同抽測及存讀；音效關閉 |
| `workplace/release171/battle-linux/RESULT.json` | AppImage讀隨包S1同盟存檔，正常進亞姆力札；P選點與返回、64點實際行軍、音效開啟兩回合到第3回合待命 |
| `workplace/release171/garrison-linux/RESULT.json` | AppImage前線駐留、換提督、當地領船、E／T拒絕、C旗艦、出擊及六表存讀 |

P選點圖可見P與250視野圈，小地圖切視野後確認目標，鏡頭還原原處。行軍依實際逐幀資料選出完整8步，旗艦從`536,960`到`600,960`；影片不只顯示下令。兩回合包含15個行星演出識別，未升格為新艦隊武器動畫驗收或DOS同態。

前線19庫存由`9000,3000,16000,2000,12000,3000,1000`改為`6000,3000,12000,2000,9000,3000,1000`，母星31保持。新編第3艦隊由提督41帶領，以C艦當旗艦並抵達18，讀回保持。Linux與Windows同盟六表逐檔相同，帝國沿自身基準核對。

README新增[兩張正式實包原視窗](../images/readme-v164.json)：索敵範圍及行軍結束。完整1280×720，沒有另繪或裁切。

## 完整推廣影片

`dist-all/v.1.6.4-20261031/promo/LOGH3-v.1.6.4-20261031-promo.mp4`沿[166已驗完整故事片](139-poster-frame-full-promo.md)，保留雙國徽底圖、百年對峙、兩位主角動畫、亞姆力札危局、三組海報與五語、兩軍戰場及完整要塞砲／光線／飛彈／戰鬥機演出。舊來源沒有冒稱新版新錄。

本版重錄正式AppImage的換提督、駐地領船與出擊，保留連續原速操作。其後加入8.62秒P索敵與2.37秒實際行軍，再接原有交戰。全片37段、126.883333秒、1920×1080、60fps、7613格。新版遊戲原窗1280×720等比縮至1664×936置中，字幕在遊戲畫布外，保留兩側人物及上下艦體。配樂沿同一首BGM_0F，不重啟、不循環；只縮短可選展示停留以容納新內容。

音畫、逐剪接點、黑幀、凍結分類、字幕安全框、來源及目視核對通過。正式影片21352320 bytes，SHA-256為`432fb2cef509043fd95734abfc0eeeaf917057ed91a3a8c9a396c069b0b6f50a`。AAC立體聲44100Hz，RMS0.08150、peak0.54011，對混音參考SNR37.3456dB，36個剪接點最低28.0066dB；非靜音且未削波，保留40個依實際演出重混的音效。全片無黑幀，30個新字幕安全框通過，保留段沿既有框。

37段分鏡表已目視；另以全尺寸抽樣國徽首尾、P範圍及返回、行軍前後、換提督、駐地領船及要塞砲。畫布、圖內名言、上下艦體及外置字幕完整，行軍前後可見位移。凍結偵測屬國徽、來源介面或動畫停留，沒有以靜圖替換實戰。沒有新增人耳驗收。詳見[機器交付索引](../internals/data/v164-full-delivery.json)及本機`promo/AUDIT.json`。

## 可重現入口

所有命令在儲存庫根目錄，透過Docker包裝器。`tools/release.sh`預設`logh3-release:20261031-r171`，沿169的Go1.24.13工具鏈補固定發行所需的numpy、Noto及SquashFS套件。定義在`tools/Dockerfile.release`，不移除169映像或借用主機runtime。macOS沿既有`cok-osxcross:go1.24.13-sdk15.5-r1`，`tools/macos.sh`固定image ID；Wine沿`yuan-wine-verify:noble-r2`及`tools/player_wine.sh`。

封包入口是`tools/scripts/scout_full_release.py`，依序prepare、windows／linux／macos、appimage與實際解包。prepare拒絕覆寫同版號目錄，需本機已驗基底及合法資料，不能只由公開來源重建完整版。Linux SquashFS走`tools/appimage.sh`，指定`LOGH3_APPIMAGE_BATCH=release171`與`LOGH3_APPIMAGE_VERSION=v.1.6.4-20261031`。macOS編譯走`tools/macos.sh`及`tools/scripts/local_test_macos_build.sh`。

```sh
bash tools/release.sh python3 tools/scripts/scout_promo.py prepare
bash tools/release.sh python3 tools/scripts/scout_promo.py render
bash tools/release.sh python3 tools/scripts/scout_promo.py grid
bash tools/release.sh python3 tools/scripts/scout_promo.py audit
# 目視通過並記錄後才執行本機安裝
bash tools/release.sh python3 tools/scripts/scout_promo.py publish
bash tools/release.sh python3 tools/scripts/scout_full_receipt.py
bash tools/release.sh python3 tools/scripts/scout_readme.py --check
```

正常索敵／行軍錄影入口為`release_battle_gui.py`，駐地錄影為`release_garrison_gui.py`，皆導出既有已驗GUI驅動器後執行正常輸入。新片源、操作事件、逐幀trace及產生器SHA由分鏡與交付索引綁定。`grid`僅修正片段容器時長四捨五入，逐視訊與AAC封包核對保持內容；不增刪影格或加速。

[169的Windows交戰閃退](140-battle-crash-linux-check.md)根因仍unknown，待對應日誌或存檔。#48同戰場讀回仍待完成。沒有新增Windows／macOS真機、FUSE、DOS全局同態、實體硬體載入速度或新的人耳驗收聲明。舊tag／Release／封包與使用者既有工作樹保留。
