# 114：MIDI I/F

狀態：CONFORMED，2026-10-14，限定使用者指定的兩種音色選擇。[#66](https://github.com/wicanr2/logh3/issues/66)。兩陣營正式GUI與偏好／實際音訊通過[驗證121](../playtest/121-landing-and-soundfont.md)，原版七項硬體埠不宣稱同態。
原版來源為[研究432](../re/432-settings-interface.md)及[原版收據087](../playtest/087-settings-interface.md)。重製差異與本輪驗收見[121](../playtest/121-landing-and-soundfont.md)。

## 本輪決定與正式契約

- 使用者明確指定MIDI I/F改為Roland MT-32／GeneralUser GS音色選擇，並確認Gsound指GeneralUser GS.sf2。原版七項硬體埠不進重製選單，BUG!出貨不可選維持113。
- 使用原MIDI I/F所在列顯示音色庫與目前選擇；展開兩項選單，右鍵或Escape取消，選擇後關閉。未匯入的完整曲庫顯示不可選，不改播另一音源冒充成功。
- GeneralUser GS沿現行`music/hd/`的11首MIDI版OGG。MT-32曲庫位於`music/mt32/`，由原版MIDI事件與使用者ROM透過固定Munt 2.8.2離線合成。FM曲目鍵對應同名MIDI樂譜，不改音符、速度或正式存檔。
- MT-32 ROM輸入為`/home/anr2/cht/mt32`，採未修改1.07控制ROM與PCM。輸入唯讀，ROM不進Git、image與一般引擎包；合成的OGG只留本機。
- 執行期沿用OGG播放器。選擇只切曲庫，靜音維持靜音；選擇後在玩家存檔目錄以獨立`preferences.json`原子保存，不改六份原版資料或MUSIC.INI。錯誤要顯示，舊偏好與曲庫保持。
- 新增`-soundfont gs|mt32`啟動選擇；未指定沿既有`-music-quality original|hd`及已保存偏好。HD/FM與新選擇可共存，音色標籤顯示實際選定項目。
- 驗收含兩套11首完整解碼、來源雜湊與非靜音、選擇／取消／缺檔拒絕、靜音保持、偏好重開、兩陣營正常設定操作及標準測試。這是刻意重製差異，不宣稱原版七項選單same-state或實機MT-32逐波形一致。

## 已證實的原版契約

- 動作22開3667七列下拉選單，完整矩形(304,180,144,180)。文字列從(320,202)開始，間距20。
- 項目是無、Roland 300H／320H／330H／332H／334H／336H；INI+4依DS:0564查表，六位元組其他欄位保持。
- 取得列號後先關閉，再停止音樂、存INI、重新偵測及重畫設定。FF取消跳過改值與初始化。
- 右鍵取消雙跑成立。原版無MPU的本環境，選無或330H後仍為FM；不證明所有硬體狀態。

## 2026-10-05 DRAFT 歷史

- 當時使用者尚未裁決七項硬體埠與軟體MIDI映射，操作探針在`tools/scripts/midi_interface_prototype.html`。本輪由上述音色選擇決定取代。
- 確認後再定typed介面設定、軟體曲目可用性及安全寫入契約；不寫玩家原版來源，不改六份原版存檔格式。
- 正式開窗、選擇／取消、重開偏好、兩音源播放要求及原版完整畫面比較仍未接，不能由087原版收據關閉#66。
- 若選單項目或尺寸改變，重定幾何與比較等級；不以原版七項圖作兩項畫面的同態golden。
