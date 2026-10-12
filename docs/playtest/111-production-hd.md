# 111：原配置 HD 正式交付

2026-10-11，第125輪。原配置HD呈現及`v.1.1.0-20261011`本機交付驗收通過；結果由本頁及[唯一來源索引](../internals/data/production-hd.json)保存。

使用者選A「典雅艦橋」，只加強原操作介面，並要求地圖與動畫HD。正式顯示1920×1440，保持640×480邏輯、4:3、視窗／按鈕位置、疊窗、操作與存檔。契約見[130介面](../spec/130-hd-drawing-journal.md)、[131地圖](../spec/131-hd-map-materials.md)、[132策略與符號](../spec/132-hd-strategy-and-symbols.md)、[133動畫](../spec/133-hd-animation-surfaces.md)。

## 交付入口

新版本`v.1.1.0-20261011`位於`dist-all/v.1.1.0-20261011/`，沿專案`full/`、`release/`、`promo/`。前版保持，不建立或移動公開tag／Release。

- 本機完整版含原版資料、HD、22首配樂、28段音效與固定音色庫。解開整個ZIP，執行`start-original`或`start-enhanced`。兩者指AI，均預設HD。
- 引擎包不含原版或HD衍生素材，依包內README匯入自己持有的DOS資料；可用`-graphics hd -art <資料夾>`指定自備HD素材。
- 存檔沿原使用者設定目錄。Windows為`%APPDATA%\logh3\save`，Linux為`$XDG_CONFIG_HOME/logh3/save`或`~/.config/logh3/save`，macOS為`~/Library/Application Support/logh3/save`。

## 呈現範圍

| 項目 | 實作與來源 |
|---|---|
| 人物 | 原82人與第82號空格，索引保持；生成原PNG與提示詞見`assets/hd/MANIFEST.json` |
| 操作介面 | 順序繪製紀錄、深藍金色面框、13945字碼Noto字型、21個原字模回退、箭頭、八向標記、游標與單位框 |
| 策略圖 | 原32區域遮罩與當幀色號，原前景位置、32星名／節點、航路與點選保持 |
| 戰術地圖 | 32張地圖、53種實際格號、73組行星及8組灰色天體四格；星雲按世界座標接續，小地圖沿原bitmap |
| 戰鬥動畫 | 12張AI圖集覆蓋45組D0基底。原幀未變艦身取新圖，受損、光束、飛彈、戰機與爆炸沿原變更；來源雜湊不符即回退 |
| 片頭／投影 | 374來源格、421播放格沿原鏡頭重建表面；投影沿原圖塊、數量、透明、裁切及遮罩 |

這次完成原配置HD呈現。不把1438動畫格稱為逐幅AI新繪，不宣稱全遊戲原版same-state一致。#48完整登陸分段流程、人耳、Windows／macOS真機與原有音訊限制各自追蹤。

## 驗證與正常實錄

最終核對器通過凍結來源、六ZIP逐檔、八封包雜湊、程式身份、素材與以下收據。標準19套件為15PASS／4無測試。正常X11與離線診斷分列，不能互相取代。

- `workplace/hd125/x11-empire-validated/`及`x11-alliance-validated/`直接使用新Linux完整包，各32張完整視窗，共44236800像素零差。未通過取圖保留原圖、差圖與數字，不排除游標或其他像素。
- `maps-r3/MAPS.json`為正式Battle消費端離線診斷，32張64鏡頭，原色重播相同；不是原版same-state。
- `animations-r11/ANIMATIONS.json`核254段1438格、45個基底雜湊、421片頭播放格與157個HD抽樣。每段獨立解碼試片不代表連續正常演出。
- S1實錄`workplace/hd125/s1-hd-guarded/`約359.2秒。由片頭、讀檔清單、S1、帝國、正常設定、實行進戰場；戰鬥命令依目的地與朝向提交，回指旗艦顯示「戰鬥」。五回合依實際「輸入」文字觀測後推進。沒有`-script`、direct-entry或局面注入。
- `OPERATOR-EVENTS.json`記一次正常X11移鼠，移開遮住「輸入」文字的游標。同次錄影未中斷，未更改局面。
- 實錄約291至347秒可見兩組艦隊交戰、飛彈、光線砲、命中、損毀與投影。90秒成片剪接較後回合，不稱90秒內連續完成戰役。帝國／同盟雙旗分居兩側，中央保留完整4:3；遊戲配樂後配，不稱現場音效同步。
- 成片90秒、1920×1080、30fps、2700幀，34張逐秒交火抽樣均不同。40個字幕安全矩形通過；無超過0.25秒黑段，平均−23.6dB、峰值−8.3dB。九段分鏡及交火已目視，九個完整遊戲區抽樣與原實錄平均通道誤差均小於4。凍結偵測含故事卡、讀檔、操作及原演出等待格，逐段分類保存在`promo/SOURCE-CHECK.json`，不宣稱全片無靜止。
- Linux完整包11入口、Windows Wine六入口、兩平台引擎啟動器各四條通過；Linux／Wine各核個人第0格、S1至S3及保存後再載S3。macOS通用架構、ARM簽章頁、系統依賴及ZIP權限通過。

完整版ZIP在`full/LOGH3-v.1.1.0-20261011-{windows-amd64,linux-amd64,macos-universal}.zip`，引擎ZIP同名放`release/`。影片為`promo/LOGH3-v.1.1.0-20261011-promo.mp4`；完整實錄為`promo/RAW-S1-HD.mp4`。`DELIVERY.json`、`SHA256SUMS.json`與`smoke/PACKAGE-CHECK.json`保存最終身份及驗收結果。

## 可重現入口

所有工作在既有Docker執行。標準命令為`bash tools/gui.sh test ./internal/... ./cmd/...`。

| 工作 | 工具及方法 |
|---|---|
| 介面 | `hd_art_model.go`提供正常公開輸入預期圖；`hd_art_gui.py`直接操作新完整包與全視窗核對 |
| 地圖／動畫 | `hd_map_check.go`、`hd_animation_check.go`枚舉實際archive，走正式renderer |
| 正常S1 | `s1_gameplay_record.py`及`tools/s1_record_display.sh`；本輪PLAN、TRACE、OPERATOR-EVENTS記輸入 |
| 完整包 | `local_test_package.py --formal --complete-hd`；`tools/macos.sh`、`local_test_macos_build.sh`及`local_test_macos_package.py` |
| 引擎及試玩 | `v1_release.py`、`local_test_smoke.py`、`local_test_windows_smoke.py`、`v1_release_smoke.py`、`bundled_scenarios_smoke.py` |
| 影片 | `s1_story_promo.py prepare|encode|audit --version v.1.1.0-20261011 --root /work/hd125/story --scenes /work/hd125/story/SCENES.json --font-dir /usr/share/fonts/opentype/noto`，走`tools/audio.sh`及既有HD影片image |
| 最終核對 | [hd_finalize.py](../../tools/scripts/hd_finalize.py)集中封包、凍結來源、輸入／輸出雜湊與證據；[hd_promo_source_check.py](../../tools/scripts/hd_promo_source_check.py)核成片遊戲區與原實錄，不把原演出停格誤判為合成凍結 |

完整版及影片含原版音畫、AI衍生圖及旗幟照片，僅留本機。引擎依RRSAL-1.0，字型及第三方授權隨包保留。macOS雙架構與ARM ad-hoc簽章不能代替真機執行或Apple公證。
