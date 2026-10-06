# 105：本機完整版與推廣影片

交付狀態：**已完成本機測試交付**。完整登陸分段、人耳、Windows實機及macOS仍未完成。

版本：`v.0.1.1-20261006`。工作項目：[92](https://github.com/wicanr2/logh3/issues/92)、[93](https://github.com/wicanr2/logh3/issues/93)。
唯一交付根：`dist-all/v.0.1.1-20261006/`。完整來源與輸出身份見[索引](../internals/data/local-test-package.json)，AI與原版事實見[440](../re/440-normal-first-turn-and-active-ai.md)。

## 交付內容

| 檔案 | bytes | SHA-256 |
|---|---:|---|
| full-local/LOGH3-v.0.1.1-20261006-windows-amd64.zip | 195847249 | ef3e1e4f189f31f3dcc1a0825d20e236da29a8a4ccaac9dec24d420df26b6acc |
| full-local/LOGH3-v.0.1.1-20261006-linux-amd64.zip | 196016706 | db2a21e840a012ab0c34ddb7fe6e11f67de8a86e60a2162b3510b5848f224b57 |
| full-local/LOGH3-v.0.1.1-20261006-engine-source.tar.gz | 3378441 | 411935f250ce6f412dd675a46d08fc529082c259298d5fba856c5c6e8f68493e |
| promo/LOGH3-v.0.1.1-20261006-promo.mp4 | 6337859 | 932b1301852a83559da30fea1035f8034cf0f58ee4919839ae764a0653e54562 |

包內有原版本機資料、正式遊戲程式、22首OGG配樂、28段音效、固定FluidR3_GM與完整MIT授權、引擎及第三方／Go授權、兩種AI啟動器、README與逐檔MANIFEST。引擎來源附完整vendor。

Windows解壓後用`start-original.cmd`或`start-enhanced.cmd`；Linux用相應`.sh`。強化預設難度1，可由CLI改0至3。畫布640×480，預設視窗960×720，可調大小。新版正常電腦會下令、移動、依規則傷害；原版與強化戰略及戰術均接線。

存檔：Windows為`%APPDATA%\logh3\save`；Linux為`$XDG_CONFIG_HOME/logh3/save`或`~/.config/logh3/save`。主選單預設安靜；片頭與進遊戲後播配樂。

## 實際抽驗

- 固定`logh3-ebiten:2`、Go1.24.13、離線vendor與明示版本建置Windows／Linux。BUILD保存父HEAD及dirty現況，逐檔SHA-256確認目前程式與封包來源相同，不冒稱乾淨HEAD建置。
- ZIP全部成員重讀、大小及SHA-256相同。Windows README含UTF-8 BOM及CRLF，啟動器ASCII／CRLF；非ASCII檔名檢查UTF-8旗標。
- Linux正式封包11次GUI公開CLI路徑：主選單、防拷、月初、造艦、命令、艦隊編號、載入、正常S3首戰、兩模式與配樂。兩模式均造出CPU艦隊，命令不同，公開存讀後續回合差異保持；FLEET／SHIP／STELLAR／PLANET保存讀回全同。
- Windows實際ZIP在Wine9.0解壓、逐檔核對，正式main／input／load／battle四入口完成。最後只補Go授權及MANIFEST再重讀ZIP，二進位、資料及配樂與Wine抽驗相同，不稱Windows實機。
- 正式Linux ALSA經虛擬PulseAudio輸出17.495465秒，峰值0.042450、RMS0.004542，有聲。人耳與實體裝置待使用者試玩。
- 正常公開S3兩種戰術AI各九回合，46個CPU分隊移動，目的地與命令不同。活動模式保留253艘反擊回歸；此S3無命令樣本玩家零戰損如實保留。
- 標準`bash tools/gui.sh test ./internal/... ./cmd/...`：18套件、13PASS、5無測試。

## 影片

60秒、1920×1080、30fps，H.264／AAC立體聲。十段真實遊戲畫面及標題，靜態剪輯，每段6秒為刻意停留，不冒稱即時錄影。六張抽樣幀已目視，標題及字幕未裁切。黑幀檢查沒有超過0.25秒黑段。

配樂為包內BGM_0F，混音加16dB及淡入淡出；封包OGG保持原檔。影片平均音量−21.7dB、峰值−6.3dB，數值未超幅，人耳驗收尚未完成。ImageMagick接觸表兩次abort，原因unknown；保留失敗，改容器內FFmpeg拼合，影片畫面不改。

## 重建入口與限制

建置：`bash tools/gui.sh xrun python3 tools/scripts/local_test_package.py --version <新版本> --revision <新批次>`，拒絕覆寫。
Linux抽驗：`bash tools/playback.sh <新批次> python3 tools/scripts/local_test_smoke.py --version <版本>`。
Windows：沿`tools/player_wine.sh`與`local_test_windows_smoke.py`，把實際ZIP放入其唯讀輸入／明示輸出邊界。
影片：`LOGH3_AUDIO_IMAGE=psychicwar-video:latest bash tools/audio.sh bash /src/tools/scripts/local_test_promo.sh <版本>`；既有映像SHA在索引保存，不另建工具鏈。
交付核對及證據：`local_test_finalize.py`、`local_test_receipt.py`。

完整版與影片含原版衍生素材，只留本機，不入Git、不公開Release。macOS、實體Windows、人耳與完整登陸分段演出未完成。原版戰術參考點強推論保持。雙玩家命令與戰場差異依使用者指示不再新增驗證，不把未知升格為原版一致。
