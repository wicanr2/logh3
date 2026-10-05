# Windows 正常曲庫播放觀測

2026-10-05，第105輪，沿 [#69](https://github.com/wicanr2/logh3/issues/69)。
父提交 `62e925a8ee081623144a17317f85fca62579b25b`。
本輪確認正常 Windows 玩家入口能輸出配樂、換曲及停止；完整播放驗收仍缺。

## 輸入與範圍

- 玩家程式沿第103輪來源重建，SHA-256
  `6af940c72f340a5ecc49b2c6dc092e0e6dca077c1136abada3cad09b49e8ffa9`。
  沒有修改產品 Go、112／116、音色、MIDI I/F、AI或正式規則。
- Windows 曲庫為 `combined-utf8-full/生成 曲庫` 的22首 OGG，逐首身份符合
  [既有合併曲庫索引](../internals/data/windows-combined-music-tools.json)。
- 原版、畫面資料及曲庫唯讀。來源玩家從預設片頭開始，真實X11輸入跳片頭、輸入138、
  關報告、開設定、選BGM 1及關配樂；無-script、direct-entry或局面注入。
- e／f各3600更新、33項X11命令，末態StateSettings。每批六張完整640×480視窗符合
  第104輪來源Linux基準，合計12張3686400像素零差。這是remake跨平台回歸，沒有新增原版same-state收據。
- 實際路徑為Windows WASAPI→Wine ALSA→PulseAudio虛擬輸出。端點PID與玩家行程一致，
  float32le、44100 Hz、雙聲道、100%音量、沒有重取樣。

## 實際結果

原檢查固定依序取第一個RMS大於0.001的曲段，要求連續逐位元組相同至少3秒，
並核對玩家狀態的時間區間。e的失敗未被後面的額外片段取代。

| 批次 | OPEN_M | BGM_0F | BGM_1F | 原首段門檻 |
|---|---:|---:|---:|---|
| e | 2.486757秒 | 額外觀測4.575873秒 | 額外觀測6.391542秒 | FAIL，片頭不足3秒 |
| f | 8.793084秒 | 7.970000秒 | 6.358413秒 | PASS，只限所列連續內段 |

e片頭另找到3.153175秒連續內段，但首段FAIL保持。固定曲段之間可見20／10毫秒全零區間；
BGM_0F有10.317毫秒全零區間。另有0.405896秒混合區間未完整解析，保存原錄音。
f所識別的三個連續內段沒有這些缺口；沒有外推整首或循環無缺口。

兩批主選單及關閉配樂各至少3秒皆全零。e／f錄音峰值均為0.999969482421875，沒有達到1.0的取樣。
錄音包含服務等待時間，總長102.271905／90.831565秒不當成曲長。
兩份完整monitor SHA-256分別為：

```text
e f44eee4c668aaba3069aa1709c36d6c66dfc5ded75106c211f2e858c5b37ebd7
f 786c5cbddeef5f2451848085bd52c66c8b1c1b0182061974d52b36988c8252e5
```

`ANALYSIS.json`保存所有固定1–8秒特徵與中間區間，不只保存通過片段。
`FIXED-GATE.json`保存原命令、退出碼與失敗文字。e沒有PCM-CHECK；f有PCM-CHECK。
兩批`*-observed.wav`各保存5秒實際輸出，未移除缺口。人耳狀態仍false。

## 環境診斷與限制

早期a–d保留。1秒片頭特徵接近靜音，不能可靠識別曲目，改為按固定振幅條件取特徵。
Linux及Windows的正式Vorbis解碼來源四首逐位元組相同，排除這四首的跨平台解碼差異。
正式signed16與輸出float的轉換是`int16/32768`，依鎖版來源核實。

增加Wine容器CPU配額未解決缺口。c有CPU限流；d限制Mesa繪圖執行緒後沒有CPU限流，仍失敗。
因此CPU配額不足不是充分解釋。[Mesa環境契約](https://docs.mesa3d.org/envvars.html#envvar-LP_NUM_THREADS)
僅用於設定軟體繪圖執行緒，沒有修改產品渲染規則。

| 固定8秒PCM對照 | GUI | 完整來源是否連續出現 |
|---|---|---|
| 直接paplay→同一Pulse monitor | 無 | 是，8秒全檔相同 |
| Windows鎖版Oto→Wine Pulse | 無 | 否，所取特徵只連續0.775986秒 |
| 相同Windows Oto→Wine ALSA→Pulse | 無 | 是，8秒全檔相同 |

這些是窄幅環境診斷，不能取代正常玩家收據。缺口已縮到Wine音訊路徑，精確成因unknown。
未反組譯DOS音效driver或追求硬體wall-clock一致。

既有Wine映像缺ALSA的Pulse外掛。新增音訊revision延伸原父映像，保留Wine9；
50個新增套件全部鎖版，沒有升級既有套件。父映像及初次probe保留供反例重現。
[Wine9的driver選擇](https://github.com/wine-mirror/wine/blob/wine-9.0/dlls/mmdevapi/main.c)
由隔離prefix的`Software\Wine\Drivers\Audio`設定；遊戲FM／MIDI設定不變。
[Ubuntu固定外掛版本](https://packages.ubuntu.com/noble/libasound2-plugins)為1.2.7.1-1ubuntu5。

| 工具 | 實際映像ID |
|---|---|
| logh3-windows-audio:1 | sha256:28adb713b85470848a83d297279520f1e02bb2ea7f74c411008cb07c253633e0 |
| logh3-windows-audio:105-probe | sha256:0527c1d89a17404c1a02f2af91e75d8bd617e430343761a913b9e6638b0e4c63 |
| logh3-ebiten:3 | sha256:ecb3612ac7c18d0cd225fd55b7b186052a8bf0029e82aa786a23f4b88f196fa8 |
| colonization-windows-verification:20261001-r1 | sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201 |

## 重跑入口

沿用既有父映像，建置時需官方套件下載；播放時無網路、無實體音效裝置。
兩個終端使用相同的新批次名稱，先等服務印READY，再啟動玩家。

```bash
bash tools/build_windows_audio.sh
LOGH3_GUI_IMAGE=logh3-ebiten:3 bash tools/gui.sh xrun \
  python3 tools/scripts/wine_playback_sink.py <新批次>
LOGH3_WINE_CPUS=4 LOGH3_MESA_THREADS=1 LOGH3_WINE_AUDIO_DRIVER=alsa \
  bash tools/player_wine_audio.sh <新批次> \
  python3 /tools/scripts/windows_music_player.py <新批次>
bash tools/gui.sh xrun python3 tools/scripts/windows_music_compare.py <新批次>
bash tools/gui.sh xrun python3 tools/scripts/windows_music_analysis.py <新批次>
```

服務180秒、Wine包裝器3分鐘、Xvfb與Pulse皆有擁有程序及清理。
服務、玩家、來源、每項輸入及UI時間戳保存在各批次JSON；聲音時間軸只作明示近似對帳。
正式Vorbis對照由[解碼器](../../tools/scripts/windows_music_decode.go)產生，
完整身份及診斷輸入見[本輪來源收據](../internals/data/windows-music-playback.json)，
用[收據核對器](../../tools/scripts/windows_music_playback_receipt.py)重驗。

標準18套件13PASS／5無測試，Game177.805秒、gin3 104.550秒。
第103輪20469來源與第104輪1096來源／164原版輸入重核保持。
Windows實機、公開保存／讀回、完整播放／循環、人耳、macOS及正式dist-all仍缺，#69 OPEN。
