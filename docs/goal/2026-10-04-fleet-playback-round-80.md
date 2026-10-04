# 第80輪：正式播放器與正常GUI音訊

2026-10-04，父提交134dcae343fd08b6931b78c63957f34433673a37。
沿open [#48](https://github.com/wicanr2/logh3/issues/48)、[#69](https://github.com/wicanr2/logh3/issues/69)及[#79](https://github.com/wicanr2/logh3/issues/79)。
89張worklist／9張open，269條核對零失效；第79輪屬於進展，四案GUI收據已推送。

依104 READY核正式音效Context、音效一次性生命週期、配樂與音效獨立控制及正常GUI交戰。
主機無/dev/snd，既有GUI映像缺PulseAudio；以既有GUI2為基底建立明確的播放驗證revision，
固定PulseAudio及ALSA插件版本，使用有時鐘的虛擬輸出與monitor錄實際PCM，不偽造實體音效卡。
先重現正式裝置判斷，再依既有「裝置可用才播放」契約修正已證實的誤判。
測試來源用玩家自備原版的既有真實OGG，沒有合成替代素材；原版衍生音訊仍只留本機。
解碼、正式後端輸出、畫面／音效同步與真人聆聽分開記錄，虛擬裝置不能升格真機或人耳。
不改第79輪工具及收據，不改#89 pending排序、不猜八步缺口、不新增完成或發行聲明。
