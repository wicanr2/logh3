# 銀河英雄傳說 III SP remake

以 Go 與 Ebitengine 重製 1994 年 PC DOS 中文版，保留原版操作方式，加入陣營 HD 介面與強化 AI。

選擇銀河帝國或自由行星同盟，從財政、生產與提督任命安排戰略，再親自指揮艦隊迎戰。從亞姆力札會戰重新下令，讓你的戰術改變局勢。

**目前版本：[v.1.2.0-20261012](https://github.com/wicanr2/logh3/releases/tag/v.1.2.0-20261012)。** GitHub 提供引擎與程式來源，玩家須自備合法取得的原版遊戲資料。含遊戲資料的完整版與推廣影片僅保留本機，不加入 Git 或 Release 附件。

## 下載與本機完整版

| 平台 | GitHub 引擎包 |
|---|---|
| Windows amd64 | [下載 ZIP](https://github.com/wicanr2/logh3/releases/download/v.1.2.0-20261012/LOGH3-v.1.2.0-20261012-windows-amd64.zip) |
| Linux amd64 | [下載 ZIP](https://github.com/wicanr2/logh3/releases/download/v.1.2.0-20261012/LOGH3-v.1.2.0-20261012-linux-amd64.zip) |
| macOS Intel／Apple Silicon | [下載通用版 ZIP](https://github.com/wicanr2/logh3/releases/download/v.1.2.0-20261012/LOGH3-v.1.2.0-20261012-macos-universal.zip) |

[引擎來源包](https://github.com/wicanr2/logh3/releases/download/v.1.2.0-20261012/LOGH3-v.1.2.0-20261012-engine-source.tar.gz) · [SHA-256 校驗清單](https://github.com/wicanr2/logh3/releases/download/v.1.2.0-20261012/SHA256SUMS.txt) · [發行驗證](docs/playtest/119-hero-ai-release-promo.md)

本機交付保留於以下位置。`dist-all/` 已由 `.gitignore` 排除，不隨儲存庫下載。

| 本機內容 | 路徑 |
|---|---|
| 三平台完整版與啟動器 | `dist-all/v.1.2.0-20261012/full/` |
| 約 125 秒有聲推廣片與正常遊玩實錄 | `dist-all/v.1.2.0-20261012/promo/` |
| 封包、聲畫及發布驗證 | `dist-all/v.1.2.0-20261012/smoke/` |

推廣片包含雙旗 HD 片頭、S1 讀檔、三個額外劇本、名將 AI 比較、F7 即時切換，以及約 52 秒正常速度艦隊交火。指定 1988 動畫片段的原生畫質為 480p，保留來源標示；成片與相關素材僅留本機。

## 主要特點

- **保留原版操作。** 維持 4:3 畫面、原版字形、立體框線與按鈕位置；修正確認、財政、訂造及建造操作，加入滑鼠停留反光。
- **兩套陣營 HD。** 帝國採藍金配色，同盟採赤焰戰線；戰略星圖重繪星球與徽章，保留星球名稱及航路。戰術地圖、人物與艦影支援 HD，部隊圖示保留原版細節。
- **隨時切換畫面。** F7 切換原版／HD，F1 開啟說明；滑鼠移到頂端可顯示功能列。顯示敵軍從設定選單切換，預設關閉。
- **強化戰略與戰術 AI。** 可選原版規則模式或四級強化難度；強化 AI 會集結兵力、協同移動並尋找包圍方向。
- **名將協同。** 楊威利採牽制與雙翼包抄，萊因哈特採集中突擊與側後封鎖。僅強化模式，兩人作為存活帶隊提督參戰時，所屬陣營全部艦隊的新移動額度為 1.5 倍，離場即解除。
- **三個額外劇本。** 取進度保留 S1「亞姆力札會戰」、S2「諸神的黃昏」、S3「巴米利恩會戰」，可正常載入及存讀。
- **有聲交火。** 支援遊戲配樂、艦隊音效與普通速度戰鬥動畫。引擎包需另匯入自有音訊與 HD 素材；原版規則模式、動畫時鐘及存檔格式保持。

## 遊戲畫面

以下八張取自本版 Linux 正式程式的正常 S1 遊玩實錄，保留完整 960×720 視窗。[截圖來源與雜湊](docs/images/readme-v120.json)可核對程式、操作紀錄及影片時碼。

| 帝國藍金介面與星圖 | 同盟赤焰戰線介面與星圖 |
|---|---|
| ![帝國HD戰略星圖，保留星球名稱與航路](docs/images/readme-v120-empire-map.png) | ![同盟HD戰略星圖與艦隊清單](docs/images/readme-v120-alliance-map.png) |

| 三個額外劇本 | 強化模式戰術戰場 |
|---|---|
| ![正常取進度清單中的S1、S2、S3](docs/images/readme-v120-scenarios.png) | ![同盟視角第4回合戰場，顯示敵軍已開啟](docs/images/readme-v120-tactical-ai.png) |

| 原版顯示 | 同一戰場切回 HD |
|---|---|
| ![正常遊戲按F7切為原版顯示](docs/images/readme-v120-switch-original.png) | ![同一戰術狀態按F7切回HD](docs/images/readme-v120-switch-hd.png) |

| 中子光線砲交火 | 戰鬥機交火 |
|---|---|
| ![正常戰鬥實錄中的中子光線砲與HD艦影](docs/images/readme-v120-beam.png) | ![正常戰鬥實錄中的戰鬥機攻擊與HD艦影](docs/images/readme-v120-fighters.png) |

## 怎麼跑

1. 下載並解壓對應平台的引擎包。
2. 找到自己原版資料中同時包含 `GIN3.COM`、`GRAPH.IMG`、`FACE.AMG` 的目錄，依下表放入引擎包。原始來源可另外保留。
3. 啟動原版規則或強化模式。防拷畫面輸入 `138`，按確認或 Enter。

| 平台 | 原版資料位置 | 原版規則／強化模式啟動器 |
|---|---|---|
| Windows | `gin3.exe` 旁的 `original/` | `start-original.cmd`／`start-enhanced.cmd` |
| Linux | `gin3` 旁的 `original/` | `start-original.sh`／`start-enhanced.sh` |
| macOS | `.app/Contents/Resources/original/` | `start-original.command`／`start-enhanced.command` |

Linux 也可直接指定既有原版資料目錄：

```sh
./start-enhanced.sh -orig "/path/to/game"
```

啟動器可加 `-music "/path/to/music"` 匯入自行準備的 OGG 配樂與音效；加 `-graphics hd -art "/path/to/art"` 使用自有 HD 素材。未匯入音訊時，引擎包保持安靜。本機完整版已附該次本機交付的素材。

| 操作 | 功能 |
|---|---|
| F1 | 說明 |
| F7 | 原版／HD 顯示切換 |
| 滑鼠移到頂端 | 顯示功能列 |
| 系統指令 → 設定 → 顯示敵軍 | 遊戲中開關敵軍顯示 |

Windows 存檔位於 `%APPDATA%\logh3\save`；Linux 位於 `$XDG_CONFIG_HOME/logh3/save`，未設定時使用 `~/.config/logh3/save`；macOS 位於 `~/Library/Application Support/logh3/save`。兩種 AI 模式共用存檔。

Linux 需要 X11 或 XWayland、OpenGL 與 ALSA；Windows 需要 OpenGL 驅動；macOS 最低版本為 11.0，尚未經 Apple 公證。詳細步驟與第三方聲明見包內 `README.txt`。

## 驗證範圍

正常玩家路徑已抽測讀檔、財政與生產操作、艦隊指令、戰場移動、交火、存讀及顯示切換。兩陣營的原版規則／強化模式從相同初始艦隊比較，兩位名將各有正常實戰錄影；普通速度演出的格序、音效閘門及結果時點另有收據。

Linux 封包已實際啟動，Windows 以 Wine 驗證，macOS 完成雙架構、依賴與 ARM 簽章檢查。Windows／macOS 真機、人耳與完整登陸／占領演出仍待驗證。HD 是重繪模式，不宣稱全遊戲與 DOS 原版逐像素或流程完全一致。

詳細證據見[發行驗證119](docs/playtest/119-hero-ai-release-promo.md)。目前狀態以 [CONTEXT.md](CONTEXT.md)、[驗證矩陣](VERIFICATION-MATRIX.md)及 [GitHub worklist Issue](https://github.com/wicanr2/logh3/issues?q=label%3Aworklist)為準。

## 重製與保存

本專案同時保存台灣中文版的資料格式、規則、畫面與人物傳。逆向結論附原始定位與證據等級；原版畫面及流程對拍使用 [dosgolem](https://github.com/wicanr2/dosgolem)，規則層另有經差異測試的 C 參考實作。原版規則模式與重製新增功能分開記錄。

開發、測試與素材處理使用既有 Docker 工具鏈。將合法原版資料放入 `org_game/extracted/` 後：

```sh
tools/gui.sh build ./cmd/gin3
tools/gui.sh run ./cmd/gin3 -orig /orig -ai 1
tools/gui.sh test ./cmd/... ./internal/...
```

## 文件

| 入口 | 內容 |
|---|---|
| [CONTEXT.md](CONTEXT.md)／[驗證矩陣](VERIFICATION-MATRIX.md) | 目前狀態與驗證範圍 |
| [規格索引](docs/spec/000-README.md) | 規則、操作與驗收契約 |
| [逆向資料索引](docs/internals/00-README.md)／[調查紀錄](docs/re/) | 資料格式、演算法與原始證據 |
| [人物傳](docs/lore/character-biographies.md) | 原版人物經歷保存 |
| [試玩與交付](docs/playtest/) | 正常玩家路徑、畫面及封包收據 |
| [WORKLOG.md](WORKLOG.md)／[AGENTS.md](AGENTS.md) | 工作歷程與協作規範 |

## 授權

引擎採[復古重製 source-available 授權條款 1.0（RRSAL-1.0）](LICENSE)：非商業用途免費，允許修改與再散布；實況、影片與平台分潤明示允許，商業用途另洽。第三方工具與音色庫依各自授權。

此授權不涵蓋原版遊戲檔、素材、譯文及其轉檔產物。本專案為非官方重製，與原作權利人無隸屬關係。儲存庫維持私有；研究素材與畫面不隨引擎包散布，完整版及動畫影片僅保留本機。
