# 原畫風 HD 圖片

使用者看過萊因哈特修復試樣後，於 2026-10-06 回答「ok」，核准以原畫風修復圖片並逐張校對人物特徵。定版的第 0 號圖原檔保存，沒有重新生成或改圖。

目前82張人物與1張原版單色空格已到齊，首批5張定版原檔保持。策略前景、戰術星雲、四格天體、小地圖與單位標記已接正式遊戲。[45組艦型圖集](animation/MANIFEST.json)以原D0基底雜湊接入交火，片頭沿原幀表面重建，下方投影保留原操作序列。第125輪正常交火實錄、三平台封包及90秒雙旗故事影片的歷史驗收見[111](../../docs/playtest/111-production-hd.md)。

介面沿[135](../../docs/spec/135-original-style-hd-ui.md)重播原立體框，文字由當幀原字模的實際像素重建，保留原筆畫、中心、裁切及疊放。[字型圖集](font/MANIFEST.json)保存前版Noto Serif CJK TC來源、授權及字碼，保留載入相容用途。帝國以B為基礎，採用使用者新確認的A適度調暗藍色與黃金；同盟維持赤焰戰線。[主題資料](interface-themes.json)保存現行顏色與核准PNG雜湊，前版核准原圖保持。頂端功能列、F1說明、F7原版／HD與按鈕反光見[138](../../docs/spec/138-top-toolbar-and-hover.md)；現行敵軍顯示由設定選單操作，F8已移除。[星雲材質](material/MANIFEST.json)與[完整提示詞](material/nebula-01.prompt.txt)由內建image_gen產生，保留原PNG。現行狀態見[CONTEXT](../../CONTEXT.md)，前版兩陣營34張正常視窗及保存驗證見[114](../../docs/playtest/114-alliance-fireline.md)。

- [圖片來源與 SHA-256](MANIFEST.json)保存原圖、輸出與人物索引。
- 每張圖片旁的 `face-NN.prompt.txt` 保存內建 image_gen 的提示詞。
- [規格122](../../docs/spec/122-hd-artwork.md)定義原索引、座標、遮擋、顯示及回退。
- [完整人物與呈現契約123](../../docs/spec/123-hd-complete-presentation.md)區分人物修復與其他素材的邊緣細化。
- [正式封包與正常操作111](../../docs/playtest/111-production-hd.md)保存現行驗收；[首批108](../../docs/playtest/108-hd-artwork.md)及[前版109](../../docs/playtest/109-v1-release.md)保留歷史範圍。

圖片衍生自原版素材，沿本專案私人保存及本機測試包邊界。此目錄不表示已取得公開散布原版衍生素材的權利。

星圖採使用者核准第三稿的新繪星球與徽章，保留原功能色、航路及全部名稱。來源及提示見[strategy/MANIFEST.json](strategy/MANIFEST.json)與[strategy/PROMPTS.json](strategy/PROMPTS.json)。F7可切換原版與HD，F6保留相容入口。
