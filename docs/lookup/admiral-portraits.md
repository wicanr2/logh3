# 提督頭像與能力值對照表（人工查詢用）

這份表給 [#16](https://github.com/wicanr2/logh3/issues/16) 拿紙本「登場人物初期資料表」或其他來源逐人比對用：找出 `+20`／`+21`／`+22`／`+24` 在紙本上叫什麼。粗體欄是還沒有名字的欄位，其餘十欄已由提督詳細資料那張畫面標名（[`docs/internals/08`](../internals/08-admiral-dt.md)）。

- 數值來源：`ADMIRAL.DTX`（新遊戲樣板），82 筆 × 28 bytes，欄位是記錄內位移。`+20` 在三格存檔裡會變，`+21`／`+22`／`+24` 四個版本一致。
- 頭像：`FACE.AMG` 第 n 張 ＝ 提督索引 n（confirmed，[`docs/internals/09`](../internals/09-face-amg.md)），檔案在 [`assets/original/face/`](../../assets/original/face/)，像素與 PNG 雜湊在 `docs/internals/data/face-amg.json`。
- 姓名與陣營：`KEIREKI.DAT`（`docs/internals/data/keireki.json`）。
- 比對時的已知事實：公開的 Win 版能力表只有那十個已命名欄位，數值逐欄吻合，沒有粗體那四欄（[`docs/re/287`](../re/287-admiral-remaining-fields.md)）。紙本若有第十一欄以上，先拿萊因哈特（第 0 列）對 `+20`＝120、`+21`＝140、`+22`＝130、`+24`＝95。

由 `tools/py.sh tools/scripts/admiral_lookup.py` 產生，不要手改。

| # | 頭像 | 姓名 | 陣營 | 統率力 | 移動 | 回轉 | 營運 | 光線砲 | 飛彈 | 戰鬥機 | 防禦 | 占領 | **+20** | **+21** | **+22** | 運用 | **+24** |
|---:|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 0 | <img src="../../assets/original/face/face-00.png" width="48" alt="0"> | 萊因哈特 | 帝國 | 138 | 110 | 118 | 100 | 119 | 114 | 110 | 117 | 126 | 120 | 140 | 130 | 115 | 95 |
| 1 | <img src="../../assets/original/face/face-01.png" width="48" alt="1"> | 吉爾菲艾斯 | 帝國 | 110 | 108 | 105 | 120 | 112 | 118 | 111 | 110 | 116 | 112 | 90 | 122 | 113 | 95 |
| 2 | <img src="../../assets/original/face/face-02.png" width="48" alt="2"> | 米達麥亞 | 帝國 | 120 | 110 | 120 | 102 | 113 | 102 | 98 | 102 | 113 | 108 | 100 | 120 | 120 | 90 |
| 3 | <img src="../../assets/original/face/face-03.png" width="48" alt="3"> | 羅嚴塔爾 | 帝國 | 118 | 105 | 110 | 95 | 112 | 106 | 105 | 113 | 115 | 120 | 80 | 128 | 114 | 90 |
| 4 | <img src="../../assets/original/face/face-04.png" width="48" alt="4"> | 畢典菲爾特 | 帝國 | 115 | 100 | 92 | 80 | 126 | 130 | 110 | 69 | 93 | 73 | 135 | 80 | 96 | 65 |
| 5 | <img src="../../assets/original/face/face-05.png" width="48" alt="5"> | 謬拉 | 帝國 | 108 | 94 | 100 | 93 | 97 | 102 | 93 | 130 | 108 | 105 | 125 | 98 | 108 | 75 |
| 6 | <img src="../../assets/original/face/face-06.png" width="48" alt="6"> | 梅克林格 | 帝國 | 95 | 98 | 98 | 112 | 90 | 91 | 99 | 108 | 110 | 95 | 88 | 95 | 102 | 85 |
| 7 | <img src="../../assets/original/face/face-07.png" width="48" alt="7"> | 魯茲 | 帝國 | 94 | 92 | 92 | 99 | 100 | 96 | 90 | 100 | 90 | 93 | 95 | 100 | 103 | 40 |
| 8 | <img src="../../assets/original/face/face-08.png" width="48" alt="8"> | 瓦列 | 帝國 | 99 | 102 | 97 | 89 | 95 | 100 | 97 | 96 | 93 | 101 | 91 | 98 | 110 | 80 |
| 9 | <img src="../../assets/original/face/face-09.png" width="48" alt="9"> | 艾齊納哈 | 帝國 | 95 | 97 | 96 | 102 | 92 | 98 | 95 | 101 | 91 | 98 | 100 | 112 | 69 | 70 |
| 10 | <img src="../../assets/original/face/face-10.png" width="48" alt="10"> | 坎普 | 帝國 | 101 | 90 | 94 | 88 | 90 | 91 | 126 | 95 | 81 | 78 | 110 | 90 | 98 | 20 |
| 11 | <img src="../../assets/original/face/face-11.png" width="48" alt="11"> | 法倫海特 | 帝國 | 93 | 96 | 110 | 92 | 103 | 110 | 97 | 81 | 91 | 88 | 87 | 99 | 100 | 30 |
| 12 | <img src="../../assets/original/face/face-12.png" width="48" alt="12"> | 舒坦梅茲 | 帝國 | 90 | 92 | 94 | 100 | 95 | 92 | 96 | 103 | 89 | 81 | 100 | 105 | 104 | 35 |
| 13 | <img src="../../assets/original/face/face-13.png" width="48" alt="13"> | 克斯拉 | 帝國 | 83 | 97 | 93 | 112 | 93 | 88 | 94 | 99 | 121 | 85 | 95 | 90 | 97 | 50 |
| 14 | <img src="../../assets/original/face/face-14.png" width="48" alt="14"> | 連內肯普 | 帝國 | 87 | 98 | 92 | 108 | 95 | 91 | 93 | 92 | 83 | 82 | 115 | 106 | 106 | 10 |
| 15 | <img src="../../assets/original/face/face-15.png" width="48" alt="15"> | 托爾奈森 | 帝國 | 99 | 108 | 94 | 90 | 98 | 102 | 82 | 94 | 92 | 76 | 103 | 90 | 90 | 5 |
| 16 | <img src="../../assets/original/face/face-16.png" width="48" alt="16"> | 奧貝斯坦 | 帝國 | 90 | 108 | 90 | 118 | 98 | 100 | 92 | 102 | 99 | 134 | 85 | 110 | 75 | 35 |
| 17 | <img src="../../assets/original/face/face-17.png" width="48" alt="17"> | 錦茲 | 帝國 | 86 | 105 | 116 | 108 | 82 | 86 | 82 | 85 | 82 | 91 | 100 | 105 | 95 | 60 |
| 18 | <img src="../../assets/original/face/face-18.png" width="48" alt="18"> | 拜耶爾藍 | 帝國 | 105 | 102 | 102 | 110 | 95 | 100 | 95 | 90 | 88 | 93 | 110 | 98 | 103 | 85 |
| 19 | <img src="../../assets/original/face/face-19.png" width="48" alt="19"> | 列肯道夫 | 帝國 | 89 | 106 | 100 | 103 | 88 | 91 | 85 | 108 | 79 | 92 | 90 | 109 | 93 | 50 |
| 20 | <img src="../../assets/original/face/face-20.png" width="48" alt="20"> | 迪爾克先 | 帝國 | 103 | 107 | 95 | 112 | 79 | 78 | 82 | 71 | 83 | 108 | 97 | 92 | 94 | 40 |
| 21 | <img src="../../assets/original/face/face-21.png" width="48" alt="21"> | 瑞文茲 | 帝國 | 90 | 91 | 102 | 106 | 89 | 85 | 72 | 101 | 88 | 113 | 100 | 97 | 92 | 40 |
| 22 | <img src="../../assets/original/face/face-22.png" width="48" alt="22"> | 列佛爾特 | 帝國 | 93 | 91 | 103 | 97 | 99 | 88 | 95 | 100 | 87 | 85 | 94 | 92 | 82 | 25 |
| 23 | <img src="../../assets/original/face/face-23.png" width="48" alt="23"> | 庫典森 | 帝國 | 77 | 89 | 91 | 105 | 81 | 84 | 105 | 90 | 83 | 87 | 98 | 86 | 90 | 45 |
| 24 | <img src="../../assets/original/face/face-24.png" width="48" alt="24"> | 萊博爾 | 帝國 | 94 | 98 | 106 | 99 | 87 | 82 | 93 | 85 | 96 | 92 | 107 | 111 | 107 | 25 |
| 25 | <img src="../../assets/original/face/face-25.png" width="48" alt="25"> | 克里希 | 帝國 | 113 | 92 | 104 | 100 | 97 | 85 | 82 | 96 | 80 | 85 | 100 | 93 | 83 | 40 |
| 26 | <img src="../../assets/original/face/face-26.png" width="48" alt="26"> | 魯比茲 | 帝國 | 88 | 82 | 93 | 98 | 94 | 97 | 82 | 84 | 80 | 96 | 91 | 108 | 81 | 50 |
| 27 | <img src="../../assets/original/face/face-27.png" width="48" alt="27"> | 山德斯 | 帝國 | 82 | 110 | 95 | 102 | 83 | 81 | 78 | 81 | 90 | 117 | 92 | 91 | 84 | 45 |
| 28 | <img src="../../assets/original/face/face-28.png" width="48" alt="28"> | 奈西巴哈 | 帝國 | 87 | 90 | 82 | 95 | 100 | 87 | 94 | 84 | 92 | 85 | 100 | 105 | 100 | 25 |
| 29 | <img src="../../assets/original/face/face-29.png" width="48" alt="29"> | 布連塔諾 | 帝國 | 93 | 98 | 106 | 103 | 81 | 79 | 78 | 108 | 87 | 81 | 94 | 100 | 78 | 20 |
| 30 | <img src="../../assets/original/face/face-30.png" width="48" alt="30"> | 布魯丹巴 | 帝國 | 87 | 91 | 93 | 100 | 82 | 83 | 82 | 81 | 88 | 90 | 104 | 101 | 81 | 40 |
| 31 | <img src="../../assets/original/face/face-31.png" width="48" alt="31"> | 菲爾納 | 帝國 | 82 | 93 | 88 | 108 | 85 | 79 | 72 | 81 | 103 | 95 | 83 | 90 | 76 | 10 |
| 32 | <img src="../../assets/original/face/face-32.png" width="48" alt="32"> | 修特萊 | 帝國 | 80 | 112 | 111 | 126 | 79 | 78 | 82 | 81 | 78 | 95 | 100 | 92 | 85 | 10 |
| 33 | <img src="../../assets/original/face/face-33.png" width="48" alt="33"> | 皮羅 | 帝國 | 98 | 94 | 97 | 106 | 95 | 98 | 82 | 105 | 80 | 95 | 100 | 97 | 93 | 80 |
| 34 | <img src="../../assets/original/face/face-34.png" width="48" alt="34"> | 貝爾肯格林 | 帝國 | 102 | 96 | 95 | 106 | 93 | 107 | 84 | 92 | 75 | 98 | 98 | 102 | 105 | 75 |
| 35 | <img src="../../assets/original/face/face-35.png" width="48" alt="35"> | 費塞尼加 | 帝國 | 88 | 85 | 94 | 120 | 81 | 76 | 98 | 99 | 108 | 110 | 93 | 100 | 97 | 20 |
| 36 | <img src="../../assets/original/face/face-36.png" width="48" alt="36"> | 阿姆斯道爾夫 | 帝國 | 78 | 100 | 105 | 100 | 92 | 96 | 94 | 95 | 90 | 99 | 91 | 102 | 80 | 40 |
| 37 | <img src="../../assets/original/face/face-37.png" width="48" alt="37"> | 格雷布納 | 帝國 | 111 | 87 | 91 | 96 | 92 | 95 | 88 | 93 | 82 | 73 | 105 | 90 | 95 | 35 |
| 38 | <img src="../../assets/original/face/face-38.png" width="48" alt="38"> | 烏爾拉 | 帝國 | 72 | 86 | 103 | 112 | 86 | 91 | 87 | 83 | 84 | 112 | 112 | 100 | 79 | 55 |
| 39 | <img src="../../assets/original/face/face-39.png" width="48" alt="39"> | 狄克爾 | 帝國 | 91 | 82 | 98 | 118 | 82 | 98 | 90 | 76 | 91 | 80 | 110 | 103 | 107 | 30 |
| 40 | <img src="../../assets/original/face/face-40.png" width="48" alt="40"> | 亞典波羅 | 同盟 | 110 | 109 | 103 | 95 | 109 | 107 | 98 | 103 | 100 | 104 | 89 | 108 | 108 | 90 |
| 41 | <img src="../../assets/original/face/face-41.png" width="48" alt="41"> | 派特 | 同盟 | 80 | 97 | 98 | 92 | 98 | 90 | 96 | 86 | 83 | 74 | 115 | 85 | 107 | 70 |
| 42 | <img src="../../assets/original/face/face-42.png" width="48" alt="42"> | 路菲普 | 同盟 | 87 | 102 | 95 | 95 | 92 | 100 | 85 | 94 | 86 | 83 | 108 | 88 | 92 | 30 |
| 43 | <img src="../../assets/original/face/face-43.png" width="48" alt="43"> | 培特雷 | 同盟 | 88 | 93 | 94 | 92 | 101 | 91 | 95 | 83 | 94 | 85 | 110 | 82 | 88 | 10 |
| 44 | <img src="../../assets/original/face/face-44.png" width="48" alt="44"> | 比克古 | 同盟 | 120 | 105 | 108 | 109 | 110 | 120 | 104 | 107 | 112 | 108 | 87 | 112 | 112 | 95 |
| 45 | <img src="../../assets/original/face/face-45.png" width="48" alt="45"> | 慕亞 | 同盟 | 85 | 92 | 94 | 91 | 97 | 103 | 98 | 86 | 96 | 79 | 133 | 85 | 87 | 20 |
| 46 | <img src="../../assets/original/face/face-46.png" width="48" alt="46"> | 赫伍得 | 同盟 | 95 | 95 | 100 | 98 | 101 | 82 | 90 | 85 | 85 | 77 | 105 | 88 | 99 | 35 |
| 47 | <img src="../../assets/original/face/face-47.png" width="48" alt="47"> | 阿普頓 | 同盟 | 98 | 100 | 95 | 88 | 95 | 92 | 87 | 102 | 92 | 80 | 90 | 89 | 101 | 45 |
| 48 | <img src="../../assets/original/face/face-48.png" width="48" alt="48"> | 亞爾•沙列姆 | 同盟 | 88 | 90 | 97 | 92 | 101 | 103 | 98 | 83 | 75 | 86 | 88 | 90 | 103 | 40 |
| 49 | <img src="../../assets/original/face/face-49.png" width="48" alt="49"> | 伍蘭夫 | 同盟 | 125 | 98 | 102 | 82 | 118 | 119 | 98 | 90 | 105 | 98 | 90 | 115 | 114 | 90 |
| 50 | <img src="../../assets/original/face/face-50.png" width="48" alt="50"> | 魯格拉希 | 同盟 | 87 | 90 | 98 | 80 | 95 | 103 | 92 | 93 | 105 | 86 | 140 | 80 | 96 | 25 |
| 51 | <img src="../../assets/original/face/face-51.png" width="48" alt="51"> | 波羅汀 | 同盟 | 106 | 97 | 111 | 89 | 110 | 98 | 104 | 107 | 96 | 92 | 130 | 110 | 113 | 85 |
| 52 | <img src="../../assets/original/face/face-52.png" width="48" alt="52"> | 楊威利 | 同盟 | 136 | 112 | 120 | 88 | 125 | 123 | 114 | 113 | 116 | 112 | 82 | 138 | 110 | 100 |
| 53 | <img src="../../assets/original/face/face-53.png" width="48" alt="53"> | 莫頓 | 同盟 | 100 | 99 | 98 | 95 | 97 | 85 | 102 | 99 | 86 | 96 | 100 | 98 | 105 | 80 |
| 54 | <img src="../../assets/original/face/face-54.png" width="48" alt="54"> | 卡爾先 | 同盟 | 92 | 102 | 96 | 103 | 94 | 98 | 91 | 90 | 81 | 85 | 110 | 72 | 91 | 60 |
| 55 | <img src="../../assets/original/face/face-55.png" width="48" alt="55"> | 費雪 | 同盟 | 102 | 116 | 115 | 110 | 90 | 102 | 88 | 100 | 95 | 90 | 100 | 90 | 120 | 50 |
| 56 | <img src="../../assets/original/face/face-56.png" width="48" alt="56"> | 尤里安 | 同盟 | 96 | 84 | 93 | 99 | 92 | 96 | 108 | 85 | 72 | 108 | 89 | 115 | 75 | 80 |
| 57 | <img src="../../assets/original/face/face-57.png" width="48" alt="57"> | 拉歐 | 同盟 | 76 | 83 | 97 | 107 | 95 | 88 | 93 | 114 | 76 | 88 | 100 | 109 | 101 | 30 |
| 58 | <img src="../../assets/original/face/face-58.png" width="48" alt="58"> | 巴格達胥 | 同盟 | 83 | 74 | 84 | 83 | 81 | 80 | 82 | 80 | 106 | 82 | 81 | 80 | 76 | 10 |
| 59 | <img src="../../assets/original/face/face-59.png" width="48" alt="59"> | 古嚴 | 同盟 | 101 | 81 | 79 | 90 | 108 | 110 | 88 | 72 | 72 | 68 | 132 | 76 | 84 | 35 |
| 60 | <img src="../../assets/original/face/face-60.png" width="48" alt="60"> | 邱吾權 | 同盟 | 93 | 106 | 102 | 119 | 93 | 98 | 108 | 117 | 79 | 112 | 91 | 120 | 105 | 60 |
| 61 | <img src="../../assets/original/face/face-61.png" width="48" alt="61"> | 拉普 | 同盟 | 105 | 100 | 103 | 110 | 98 | 102 | 98 | 105 | 104 | 105 | 86 | 117 | 92 | 90 |
| 62 | <img src="../../assets/original/face/face-62.png" width="48" alt="62"> | 雅拉肯 | 同盟 | 92 | 104 | 80 | 95 | 98 | 96 | 84 | 72 | 108 | 70 | 120 | 83 | 97 | 65 |
| 63 | <img src="../../assets/original/face/face-63.png" width="48" alt="63"> | 法菲爾 | 同盟 | 104 | 100 | 104 | 114 | 82 | 88 | 85 | 81 | 77 | 100 | 113 | 105 | 93 | 55 |
| 64 | <img src="../../assets/original/face/face-64.png" width="48" alt="64"> | 馬里涅汀 | 同盟 | 88 | 74 | 85 | 102 | 92 | 104 | 90 | 77 | 87 | 101 | 102 | 93 | 90 | 75 |
| 65 | <img src="../../assets/original/face/face-65.png" width="48" alt="65"> | 陳 | 同盟 | 79 | 108 | 103 | 109 | 93 | 85 | 78 | 90 | 86 | 98 | 109 | 113 | 99 | 50 |
| 66 | <img src="../../assets/original/face/face-66.png" width="48" alt="66"> | 史多克斯 | 同盟 | 83 | 74 | 78 | 116 | 81 | 80 | 72 | 80 | 98 | 72 | 144 | 81 | 80 | 5 |
| 67 | <img src="../../assets/original/face/face-67.png" width="48" alt="67"> | 可那利 | 同盟 | 78 | 83 | 88 | 106 | 84 | 87 | 91 | 76 | 72 | 76 | 110 | 105 | 88 | 40 |
| 68 | <img src="../../assets/original/face/face-68.png" width="48" alt="68"> | 菲列特利加 | 同盟 | 102 | 106 | 85 | 128 | 69 | 65 | 78 | 77 | 72 | 93 | 98 | 79 | 75 | 5 |
| 69 | <img src="../../assets/original/face/face-69.png" width="48" alt="69"> | 馬利諾 | 同盟 | 85 | 105 | 104 | 106 | 98 | 93 | 86 | 86 | 75 | 81 | 119 | 82 | 85 | 85 |
| 70 | <img src="../../assets/original/face/face-70.png" width="48" alt="70"> | 沙尼亞 | 同盟 | 74 | 84 | 89 | 96 | 80 | 120 | 83 | 88 | 77 | 78 | 106 | 97 | 91 | 65 |
| 71 | <img src="../../assets/original/face/face-71.png" width="48" alt="71"> | 姆萊 | 同盟 | 83 | 86 | 97 | 120 | 86 | 93 | 86 | 87 | 98 | 113 | 108 | 96 | 80 | 30 |
| 72 | <img src="../../assets/original/face/face-72.png" width="48" alt="72"> | 先寇布 | 同盟 | 110 | 82 | 93 | 79 | 83 | 82 | 75 | 101 | 145 | 88 | 93 | 90 | 86 | 80 |
| 73 | <img src="../../assets/original/face/face-73.png" width="48" alt="73"> | 史路茲卡利達 | 同盟 | 76 | 84 | 91 | 112 | 81 | 82 | 78 | 112 | 76 | 93 | 109 | 111 | 89 | 60 |
| 74 | <img src="../../assets/original/face/face-74.png" width="48" alt="74"> | 波布蘭 | 同盟 | 93 | 76 | 83 | 72 | 70 | 73 | 138 | 70 | 77 | 81 | 83 | 84 | 75 | 25 |
| 75 | <img src="../../assets/original/face/face-75.png" width="48" alt="75"> | 卡介倫 | 同盟 | 90 | 82 | 82 | 128 | 78 | 75 | 91 | 102 | 93 | 97 | 100 | 85 | 78 | 75 |
| 76 | <img src="../../assets/original/face/face-76.png" width="48" alt="76"> | 林滋 | 同盟 | 87 | 81 | 77 | 88 | 72 | 73 | 75 | 83 | 138 | 85 | 122 | 87 | 77 | 15 |
| 77 | <img src="../../assets/original/face/face-77.png" width="48" alt="77"> | 布魯姆哈爾特 | 同盟 | 83 | 74 | 79 | 90 | 71 | 70 | 72 | 80 | 126 | 72 | 132 | 86 | 76 | 10 |
| 78 | <img src="../../assets/original/face/face-78.png" width="48" alt="78"> | 伊旺•高尼夫 | 同盟 | 91 | 73 | 76 | 100 | 76 | 75 | 126 | 82 | 73 | 85 | 100 | 80 | 75 | 20 |
| 79 | <img src="../../assets/original/face/face-79.png" width="48" alt="79"> | 派特里契夫 | 同盟 | 118 | 79 | 73 | 106 | 81 | 86 | 89 | 98 | 97 | 87 | 121 | 84 | 95 | 70 |
| 80 | <img src="../../assets/original/face/face-80.png" width="48" alt="80"> | 梅爾卡茲 | 同盟 | 127 | 95 | 109 | 92 | 110 | 98 | 109 | 108 | 110 | 112 | 100 | 121 | 112 | 90 |
| 81 | <img src="../../assets/original/face/face-81.png" width="48" alt="81"> | 舒奈德 | 同盟 | 101 | 100 | 113 | 108 | 90 | 105 | 83 | 92 | 98 | 99 | 111 | 95 | 102 | 40 |

第 82 張頭像是沒有人物的單色格（不對應任何提督），不列入。
