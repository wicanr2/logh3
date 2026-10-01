# 070 — 固定 Ogg 單串流封裝序號

狀態：**CONFORMED**（2026-10-01，限封裝層）。工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
證據：[研究 359](../re/359-native-fm-import.md)。標準來源：
[Xiph Ogg framing](https://xiph.org/ogg/doc/framing.html)。
審查：相同波形重編產物只差 serial／CRC，探針恢復完整舊 SHA-256；
不改 Vorbis 音訊、樣本數、時間位置、分段或排序，沒有新增音色／音量取捨。

後續裁決：[規格 074](074-cross-platform-fm-pcm.md) 允許跨平台 Vorbis 音訊封包差異，
改驗完整 VGM／WAV／增益後 PCM。此處仍只負責固定 serial／CRC，
不得把封裝修正當作已消除音訊差異；下列 Linux 歷史驗收保留。

1. typed input：完整 OGG bytes 與已鎖定 uint32 serial；輸入上限 64 MiB。
2. 逐頁驗 OggS、版本 0、段表與 payload 完整邊界、原始 CRC。
   頁序從 0 連續，單一舊 serial，首頁 BOS、尾頁 EOS；拒絕多串流、串接、截斷與壞 CRC。
3. 唯一變更：頁頭偏移 14–17 改固定 serial，22–25 依標準重算 CRC。
   CRC 用 direct polynomial 0x04C11DB7，起始與最終 XOR 均 0，計算時 CRC 欄位清零。
   其餘 bytes 全部保留。錯誤時不改呼叫者輸入，也不輸出部分成功結果。
4. 固定 serial 是每曲封裝鎖定值，來自版本化 metadata；不得從執行時鐘取值或重擲。
   尚未通過完整 OGG SHA-256 的曲目不能用固定 serial 掩蓋真正封包差異。
5. 驗收：重編 BGM_0F 完整輸出與既有 OGG 逐 byte／SHA-256 相同，解碼 PCM 不變；
   任意封裝碰撞、壞 CRC、截斷、錯頁序、錯 BOS／EOS 都拒絕；全產品測試通過。

停止線：本規格只完成 Ogg 封裝層；不證明 VGM 來源已重生，不改 FM 基準，
不新增原版 driver 逆向，亦不等於完整玩家匯入或平台／聆聽驗收。

驗收：真實重編 BGM_0F 恢復完整基準 bytes／SHA-256，CRC、截斷、錯頁序、
BOS／EOS 與串接拒絕通過；輸入不變。標準產品測試全綠，完整原生入口首曲
也由原版重錄後通過相同 VGM／OGG 閘門。十一首全批仍由規格 071 另行驗收。
