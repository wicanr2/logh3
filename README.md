# 銀河英雄傳說 III SP remake

以 Go 與 Ebitengine 重製 1994 年 PC DOS 中文版，保留原版操作方式，加入陣營 HD 介面與強化 AI。

選擇銀河帝國或自由行星同盟，從財政、生產與提督任命安排戰略，再親自指揮艦隊迎戰。從亞姆力札會戰重新下令，讓你的戰術改變局勢。

**目前版本：[v.1.3.0-20261013](https://github.com/wicanr2/logh3/releases/tag/v.1.3.0-20261013)。** GitHub 提供引擎與程式來源，玩家須自備合法取得的原版遊戲資料。含遊戲資料的完整版與推廣影片僅保留本機，不加入 Git 或 Release 附件。

目前原始碼已加入 **Roland MT-32／GeneralUser GS 音色選擇**，設定中的 MIDI I/F 列改為音色庫，支援取消、靜音保持與重開保存選擇。登陸、反擊與占領也已改為獨立分段演出，保留各階段的擁有者與戰損快照。這些變更尚未納入上述發行包。[本輪驗證與曲庫使用方式](docs/playtest/121-landing-and-soundfont.md)

**要塞砲演出已完成。** 支援發射通知、射程橢圓、首次開火動畫、逐分隊中彈與戰損、下方方塊及音效等待。伊謝爾倫正常操作、原版／HD、動畫關閉、F7及演出後六表存讀已驗收。[要塞砲驗收與操作路徑](docs/playtest/123-fortress-presentation.md)

## 故事背景

遙遠的未來，人類已將生活圈擴展到銀河。實行專制政治的銀河帝國與奉行民主制度的自由行星同盟，持續交戰長達一百五十年。帝國受門閥貴族的腐敗侵蝕，同盟也陷入政客操弄與長年戰爭的消耗。兩位軍事天才的崛起，開始改變這場僵局。[世界觀來源：Production I.G 故事介紹](https://www.production-ig.co.jp/works/ginei/story.html)

帝國的萊因哈特因姊姊被皇帝納入後宮，決心與摯友吉爾菲艾斯奪取權力，進而追求統一宇宙。同盟的楊威利原本立志成為歷史學家，卻以靈活的戰術屢次挽救戰局。他被稱為「不敗的楊」與「魔術師楊」，心中盼望的仍是和平。兩人各自率領艦隊，在戰場上成為彼此最難戰勝的對手。[人物設定來源：原版人物傳](docs/lore/character-biographies.md)

亞姆力札會戰是這段歷史的重要轉折。同盟軍大舉進攻帝國領土，卻因補給困境與萊因哈特的反攻陷入危局，雙方艦隊最終集結於亞姆力札星域。[會戰背景來源：動畫本傳第 12 至 15 話介紹](https://ginei.jp/story1.php)

本作讓玩家接手兩大陣營的軍政與艦隊指揮，從亞姆力札會戰等局面重新作出選擇。人物的理想與衝突構成故事背景，戰局則交由玩家的部署、補給與臨場指揮決定。

## 下載與本機完整版

| 平台 | GitHub 引擎包 |
|---|---|
| Windows amd64 | [下載 ZIP](https://github.com/wicanr2/logh3/releases/download/v.1.3.0-20261013/LOGH3-v.1.3.0-20261013-windows-amd64.zip) |
| Linux amd64 | [下載 ZIP](https://github.com/wicanr2/logh3/releases/download/v.1.3.0-20261013/LOGH3-v.1.3.0-20261013-linux-amd64.zip) |
| macOS Intel／Apple Silicon | [下載通用版 ZIP](https://github.com/wicanr2/logh3/releases/download/v.1.3.0-20261013/LOGH3-v.1.3.0-20261013-macos-universal.zip) |

[引擎來源包](https://github.com/wicanr2/logh3/releases/download/v.1.3.0-20261013/LOGH3-v.1.3.0-20261013-engine-source.tar.gz) · [SHA-256 校驗清單](https://github.com/wicanr2/logh3/releases/download/v.1.3.0-20261013/SHA256SUMS.txt) · [發行驗證](docs/playtest/120-system-dialog-polish.md)

本機交付保留於以下位置。`dist-all/` 已由 `.gitignore` 排除，不隨儲存庫下載。

| 本機內容 | 路徑 |
|---|---|
| 三平台完整版與啟動器 | `dist-all/v.1.3.0-20261013/full/` |
| 約 125 秒有聲推廣片與正常遊玩實錄 | `dist-all/v.1.2.0-20261012/promo/` |

## 主要特點

- **保留原版操作。** 維持 4:3 畫面、立體框線、按鈕位置與原版字形，支援滑鼠停留反光。
- **兩套陣營 HD。** 帝國採藍金配色，同盟採赤焰戰線；戰略星圖重繪星球與徽章，保留星球名稱及航路。戰術地圖、人物與艦影支援 HD，部隊圖示保留原版細節。
- **隨時切換畫面。** F7 切換原版／HD，F1 開啟說明；滑鼠移到頂端可顯示功能列。顯示敵軍從設定選單切換，預設關閉。
- **強化戰略與戰術 AI。** 可選原版規則模式或四級強化難度；強化 AI 會集結兵力、協同移動並尋找包圍方向。
- **名將協同。** 楊威利採牽制與雙翼包抄，萊因哈特採集中突擊與側後封鎖。僅強化模式，兩人作為存活帶隊提督參戰時，所屬陣營全部艦隊的新移動額度為 1.5 倍，離場即解除。
- **三個額外劇本。** 取進度保留 S1「亞姆力札會戰」、S2「諸神的黃昏」、S3「巴米利恩會戰」，可正常載入及存讀。
- **四款字型。** 設定可選原版點陣、原版強化、思源宋體及思源黑體，預設原版強化。原版顯示固定使用原字形；新字集需另匯入，本機完整版已附。
- **原曲 HD 音色。** 本機完整版保留原曲，升級 11 首配樂的音色與混音；設定可切 HD／FM 或靜音。強化啟動器預設 HD，原版規則啟動器預設原音色。
- **啟動進度。** 小視窗顯示實際載入階段與人物圖張數，缺檔時顯示錯誤。
- **有聲交火。** 支援遊戲配樂、艦隊音效與普通速度戰鬥動畫。引擎包需另匯入自有音訊與 HD 素材；原版規則模式、動畫時鐘及存檔格式保持。

## 遊戲畫面

載入進度與兩陣營字型設定的實際畫面，取自 v.1.3.0 Linux 版。[截圖來源與雜湊](docs/images/readme-v130.json)

![實際七階段載入進度與人物圖張數](docs/images/readme-v130-loading.png)

| 帝國思源宋體 | 同盟思源黑體 |
|---|---|
| ![帝國設定中的四字型選擇](docs/images/readme-v130-empire-font.png) | ![同盟設定中的思源黑體與音源選擇](docs/images/readme-v130-alliance-font.png) |

以下為陣營介面、劇本與交火畫面，取自 v.1.2.0 Linux 版正常 S1 遊玩，保留完整 960×720 視窗。[截圖來源與雜湊](docs/images/readme-v120.json)

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
| 系統指令 → 設定 → 介面字型 | 循環切換四款字型 |
| 系統指令 → 設定 → HD／FM | 切換配樂音色，需已匯入 HD 曲庫 |

Windows 存檔位於 `%APPDATA%\logh3\save`；Linux 位於 `$XDG_CONFIG_HOME/logh3/save`，未設定時使用 `~/.config/logh3/save`；macOS 位於 `~/Library/Application Support/logh3/save`。兩種 AI 模式共用存檔。字型與 HD 音色選擇僅本次執行有效，可用 `-font original|classic|serif|sans` 與 `-music-quality original|hd` 指定啟動值；HD 曲目放在配樂目錄的 `hd/`。

Linux 需要 X11 或 XWayland、OpenGL 與 ALSA；Windows 需要 OpenGL 驅動；macOS 最低版本為 11.0，尚未經 Apple 公證。詳細步驟與第三方聲明見包內 `README.txt`。

## 驗證範圍

Linux 版已抽測兩陣營的讀檔、財政與生產、艦隊指令、戰場移動、交火、存讀，以及畫面、字型與音色切換。要塞砲已驗收S1同盟正常演出，試射包含友軍傷害；423張原版完整場景幀的正式渲染對拍零差。現行音樂已通過使用者實際聆聽驗收。Windows 完成 Wine 驗證，macOS 完成雙架構、依賴與 ARM 簽章檢查。Windows／macOS 真機、完整登陸／占領正常流程及同戰場讀檔續戰仍待完成。HD 是重繪模式，不宣稱全遊戲與 DOS 原版逐像素或流程完全一致。

完整結果見[發行驗證](docs/playtest/120-system-dialog-polish.md)與[戰場及影片驗證](docs/playtest/119-hero-ai-release-promo.md)。目前狀態以 [CONTEXT.md](CONTEXT.md)、[驗證矩陣](VERIFICATION-MATRIX.md)及 [GitHub worklist Issue](https://github.com/wicanr2/logh3/issues?q=label%3Aworklist)為準。

## 重製與保存

本專案同時保存台灣中文版的資料格式、規則、畫面與人物傳。逆向結論附原始定位與證據等級；原版畫面及流程對拍使用 [dosgolem](https://github.com/wicanr2/dosgolem)，規則層另有經差異測試的 C 參考實作。原版規則模式與重製新增功能分開記錄。

開發與測試方式見 [AGENTS.md](AGENTS.md)，工作歷程見 [WORKLOG.md](WORKLOG.md)。

## 文件

| 入口 | 內容 |
|---|---|
| [CONTEXT.md](CONTEXT.md)／[驗證矩陣](VERIFICATION-MATRIX.md) | 目前狀態與驗證範圍 |
| [規格索引](docs/spec/000-README.md) | 規則、操作與驗收契約 |
| [逆向資料索引](docs/internals/00-README.md)／[調查紀錄](docs/re/) | 資料格式、演算法與原始證據 |
| [人物傳](docs/lore/character-biographies.md) | 原版人物經歷保存 |
| [試玩與交付](docs/playtest/) | 正常玩家路徑、畫面及封包收據 |
| [WORKLOG.md](WORKLOG.md) | 開發、驗證與交付歷程 |
| [AGENTS.md](AGENTS.md) | 開發方式與協作規範 |

## 授權

引擎採[復古重製 source-available 授權條款 1.0（RRSAL-1.0）](LICENSE)：非商業用途免費，允許修改與再散布；實況、影片與平台分潤明示允許，商業用途另洽。第三方工具與音色庫依各自授權。

此授權不涵蓋原版遊戲檔、素材、譯文及其轉檔產物。本專案為非官方重製，與原作權利人無隸屬關係。儲存庫維持私有；研究素材與畫面不隨引擎包散布，完整版及動畫影片僅保留本機。
