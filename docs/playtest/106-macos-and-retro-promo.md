# 106：macOS完整版與復古重製推廣影片

macOS封包狀態：**已完成交付，真機待驗**。
推廣片與故事首頁：**已完成**。

版本 `v.0.1.2-20261006`。工作項目：[94](https://github.com/wicanr2/logh3/issues/94)、[95](https://github.com/wicanr2/logh3/issues/95)。
唯一交付根 `dist-all/v.0.1.2-20261006/`，完整身份見[本輪索引](../internals/data/macos-retro-promo.json)。前版與[105](105-local-full-test-package.md)的凍結來源保持。

## 下載與試玩

| 檔案 | bytes | SHA-256 |
|---|---:|---|
| [macOS Intel／Apple Silicon完整版](../../dist-all/v.0.1.2-20261006/full-local/LOGH3-v.0.1.2-20261006-macos-universal.zip) | 202206334 | ff53bf5c61058cd2db25fe96da4fe60b8fb5de5e62d344a500d7631d1bcb0e95 |
| [Windows amd64完整版](../../dist-all/v.0.1.2-20261006/full-local/LOGH3-v.0.1.2-20261006-windows-amd64.zip) | 195847244 | abd619e2074477da02d29878d749ba518268bfdb4d271f72ed4266d1c2639985 |
| [Linux amd64完整版](../../dist-all/v.0.1.2-20261006/full-local/LOGH3-v.0.1.2-20261006-linux-amd64.zip) | 196016701 | 65ce9244d87770892d55a98e693d8b9894603c98b3f2646ff1323b62d446d0f7 |
| [引擎與vendor來源](../../dist-all/v.0.1.2-20261006/full-local/LOGH3-v.0.1.2-20261006-engine-source.tar.gz) | 3379193 | 6c006bc31ea34fd72bf8a0e45b8636dffc13b08d6d57d38bdbb026860ad8d8b5 |
| [72秒推廣影片](../../dist-all/v.0.1.2-20261006/promo/LOGH3-v.0.1.2-20261006-promo.mp4) | 4731668 | 3973f08f03b9b760577e3c07b90339703b32d7b5969f2591652646a5629d1e42 |

三平台包都有原版本機資料、98份執行資料JSON、22首配樂、28段音效、固定FluidR3_GM與MIT全文、引擎RRSAL-1.0及第三方／Go授權。玩家不用另下載音色庫或安裝合成器。
遊戲與活動AI沿[120](../spec/120-active-tactical-ai.md)，只更新程式版號；本輪不修改Go規則、AI或存檔格式。

macOS最低部署版本11.0。解壓後打開 `銀河英雄傳說III.app`，預設使用原版AI。也可雙擊 `start-original.command` 或 `start-enhanced.command`；強化預設難度1。
程式與完整資料都在 `.app` 內，可整個搬移；兩個 `.command` 留在 `.app` 旁。
存檔位於 `~/Library/Application Support/logh3/save`，錯誤記錄位於 `~/Library/Logs/logh3/game.log`。

此包沒有Developer ID簽章或Apple公證。若macOS阻擋開啟，使用Apple說明的「系統設定 → 隱私權與安全性 → 仍要打開」。不要求停用安全機制。操作見[Apple支援](https://support.apple.com/102445)。
macOS原生啟動、音效裝置與人耳聆聽尚待試玩；封包檢查不取代這項驗收。

Windows與Linux沿 `start-original.cmd`／`.sh` 及 `start-enhanced.cmd`／`.sh`。
兩種AI的戰略與戰術均已啟用，電腦會下令、移動及交戰。正常戰術證據沿105，不新增雙玩家命令或原版戰場差異驗證。

## 封包與平台檢查

- 從e274312正式Go來源、完整vendor與固定Go1.24.13離線建置三平台。同一完整版號進入二進位、manifest、目錄與檔名。BUILD保存來源HEAD及當時文件／工具修改，不冒稱乾淨工作樹。
- macOS沿既有 `psychicwar-osxcross:latest`，image SHA `0f50c77c732087b104b3aadc3aecb352054aa37dee8e1d293523a66dfa02ba61`，SDK15.5；CGO兩架構均指定最低11.0，再以lipo合併。Go1.24支援macOS11的契約見[Go版本說明](https://go.dev/doc/go1.24)。
- 通用Mach-O的amd64／arm64兩片、最低部署版本與11個系統依賴逐項核對；沒有建置機的非系統動態庫依賴。Apple Silicon的ad-hoc CodeDirectory共2994頁，SHA-256全部相同。Intel原生程式未簽章，外層 `.app` 未公證。規則來源為[Apple Platform Security](https://help.apple.com/pdf/security/en_GB/apple-platform-security-guide-b.pdf)及固定Go1.24.13的 `cmd/link/internal/ld/lib.go` 中 `Link.NeedCodeSign`。
- 三個ZIP全部成員重讀，CRC、大小、SHA-256及manifest清單相同。macOS共339項含manifest，Unicode檔名保留UTF-8旗標；程式與三個啟動器保留0755。來源tar全部檔案符合BUILD記錄。
- 新Linux正式二進位11條公開CLI輸入抽驗完成，包含存讀、原版／強化戰略AI與配樂。新Windows ZIP在Wine9.0完成main／input／load／battle四條入口。部分窄樣本有公開診斷參數，不把全部抽測升格為真實滑鼠正常流程。
- 新Linux配樂經ALSA與虛擬PulseAudio輸出，monitor非零、峰值未超幅，數值保存於 `smoke/audio/AUDIO.json`；不稱人耳或macOS音訊驗收。
- 標準命令 `bash tools/gui.sh test ./internal/... ./cmd/...` 完成18套件，13PASS、5無測試。產品Go未修改，已符合來源身份的快取結果照實保留。

## 推廣影片

影片為72秒、1920×1080、30fps、2160幀，H.264與AAC立體聲。採三段流水線：收集真實音畫、建立版面素材、FFmpeg合成。
主題「星圖與作戰檔案」的青色、灰色及金色取自本作16色色盤；星點與情報線條是編輯裝飾。遊戲截圖保持完整640×480畫布與原色，字幕使用獨立安全區。

| 時間 | 內容 | 呈現方式 |
|---|---|---|
| 0–7秒 | 回到指揮席 | 標題與安全單幀縮放 |
| 7–15秒 | 帝國、同盟與兩位將領 | 世界觀雙檔案卡與安全單幀縮放 |
| 15–21秒 | 星域部署 | 完整畫面與情報側欄 |
| 21–29秒 | DOS原版與重製 | 分割對照 |
| 29–35秒 | 提督與艦隊 | 艦隊檔案與數字索引 |
| 35–41秒 | 戰術戰場 | 1280×960整數最近鄰畫面，兩側字幕 |
| 41–49秒 | 原版與強化AI | 真實兩模式畫面 |
| 49–54秒 | 保存進度 | 橫向資訊卡 |
| 54–60秒 | 配樂與音效 | 22／28數字與真實音源說明 |
| 60–65秒 | 三平台封包 | 支援架構與macOS真機待驗 |
| 65–72秒 | 試玩邀請 | 版本、AI與目前驗收限制 |

與前版的差異為世界觀敘事、十一種版面、5至8秒節奏與兩段動態。
[新舊各四幀比較](../../dist-all/v.0.1.2-20261006/promo/previous-current.png)及[十一段抽樣幀](../../dist-all/v.0.1.2-20261006/promo/samples-contact.png)已目視。
戰場、平台與結尾另檢完整1080p畫面；最終字幕零裁切。影片黑段檢查沒有超過0.25秒的黑段。
凍結檢查只報出分鏡中的九段靜態展示；兩段動態各取兩張解碼幀，身份不同。這是實際截圖剪輯，未冒稱遊戲即時錄影。

原版畫面沿dosgolem f7523f9的既有正常首戰收據，沒有重跑或改原始資料。第21秒對照為nearby-state不同回合版面展示，片中文字明示；第41秒只展示已完成的兩模式收據，不新增戰場一致性裁決。
macOS平台卡只列封包資訊，未冒充Mac實機截圖。

配樂取自包內 `BGM_0F.ogg`，混音加16dB及淡入淡出，封包OGG保持。
平均音量−21.7dB、峰值−6.3dB；沒有0.8秒以上−70dB靜音段，技術檢查非靜音且未超幅。人耳確認仍待試玩。
來源、分鏡、主題、FFprobe、音畫檢查、動態幀及權利記錄都隨片保存在 `promo/`。

## README故事來源

[README](../../README.md)最前面先介紹銀河帝國與自由惑星同盟、一百五十年戰爭、萊因哈特與楊威利，再接本作玩法及保存目的。
世界觀來源為[Production I.G官方故事](https://www.production-ig.co.jp/works/ginei/story.html)，萊因哈特的從軍動機見[集英社作品介紹](https://youngjump.jp/manga/gineiden/)。動畫與漫畫頁僅支援作品故事，未用作DOS原版畫面或規則證據。

## 重建入口

先用[完整版建置器](../../tools/scripts/local_test_package.py)建立新版本及vendor來源。全部工作沿既有Docker包裝器，輸入與輸出職責見各腳本。

```sh
bash tools/gui.sh xrun python3 tools/scripts/local_test_package.py --version <新版本> --revision a
bash tools/macos.sh <新批次> bash /src/tools/scripts/local_test_macos_build.sh <新版本> workplace/package92-<新版本>/source
bash tools/gui.sh xrun python3 tools/scripts/local_test_macos_package.py --version <新版本> --batch <新批次>
bash tools/gui.sh xrun python3 tools/scripts/retro_promo_prepare.py --version <新版本>
LOGH3_AUDIO_IMAGE=psychicwar-video:latest bash tools/audio.sh bash /src/tools/scripts/retro_promo_render.sh <新版本> <新批次> cards
LOGH3_AUDIO_IMAGE=psychicwar-video:latest bash tools/audio.sh bash /src/tools/scripts/retro_promo_render.sh <新版本> <新批次> encode
```

`audit`模式可重生抽樣及動態檢查；可選第四參數僅在PNG全檔相同且影片尺寸／幀數符合時沿用前批片段。本輪草稿有PNG時間metadata差異，因此最終十一段重新編碼。
Linux抽驗沿 `tools/playback.sh`，Windows沿 `tools/player_wine.sh`。[交付核對與身份保存器](../../tools/scripts/macos_promo_finalize.py)分為finalize與receipt；後者在文件定稿後執行，拒絕覆寫來源索引。

三平台完整版及影片含原版音畫，只留本機，不入Git或公開Release。完整登陸分段演出尚未接完；實機、人耳與完整循環仍待驗收。原版戰術參考點強推論保持，未宣稱完整原版戰場同態。
