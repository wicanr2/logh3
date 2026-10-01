# 071 — 原生完整 FM／GM 匯入入口

狀態：**CONFORMED**（2026-10-01，限 Linux 完整原生匯入與正式解碼）。工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
證據：[研究 275](../re/275-fm-capture-from-opl.md)、[研究 359](../re/359-native-fm-import.md)、
[GM 階段 069](069-native-gm-import.md)、[封裝 070](070-deterministic-ogg-framing.md)。
本規格接續 069 的階段入口：無階段旗標預設完整 22 首；保留 `-gm-only`，另可抽驗 `-fm-only`。

審查：原版三條觸發路徑已由研究 275 證實；目前 clean dosgolem HEAD
`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c` 的 BGM_0F 完整 1.6B 指令擷取與保存 VGM 全檔相同。
其餘曲目依既有流程接線，每首必須先符合已鎖 VGM／OGG SHA-256 才能提交，
未通過即停止，不改參數追求碰巧通過。

## 契約

1. 完整入口選取 22 首，合法子集可由 `-tracks` 指定；階段旗標互斥，
   不准誤將 `_F` 送 GM。GM 沿用 069 的既有近似，FM 沿用原版 OPL 暫存器串。
2. 原生程式只呼叫隨包 `dosgolem-probe`、`opl2wav`、`ffmpeg`／`fluidsynth`。
   不呼叫 Docker、shell 或 Python，也不搜尋玩家 PATH。缺任一必要工具明確失敗。
3. 原版輸入只讀，在本次獨立暫存遊戲目錄複製所需來源才交給 DOS 執行器；
   OPEN_F 只在該副本將 OPEN_M 替換為 OPEN_F，曲後還原。
   不跟隨來源 symlink，不拷貝整個家目錄或輸出目錄，輸出不得位於來源內。
4. 固定原版 EXE／MDS 雜湊、probe HEAD／binary 身份、曲號、步數、tick、輸入序列、
   poke 位址／步數、錨點與 Ogg serial 由版本化鎖定 metadata 供應，原生程式不現場猜測。
   `lin:5DA64@117022200=00–07`，不接受沒有 lin 的位址映射；長度不混用十六進位。
5. BGM 八首：GIN3.COM、empty env、tick 20000、1.6B steps，沿用主選單／防拷固定輸入。
   END 兩首：ENDING.EXE、`-sound-blaster`、參數 9／1 或 9／2、1.6B steps。
   OPEN_F：GIN3.COM、替換副本、不送輸入、1.2B steps，錨仍 OPEN_M.MDS。
   每首 probe 有 15 分鐘逾時；缺錨／缺 VGM、SHA-256 不符、合成或編碼失敗均不能成功。
6. 同一 AdPlug／Vorbis 管線、44100 Hz、q5、既有全批 -1.00 dB 增益；
   Ogg serial 與 CRC 由 070 固定。完整 VGM／OGG 均需符合 SOURCES.tsv 的原有雜湊，
   解碼後確認立體聲、時長、非靜音；成功後才原子提交單曲與收據。
7. 收據保存完整 flags、原始／副本輸入身份、工具 HEAD／程式雜湊、VGM／OGG 雜湊，
   OPEN_F 標明 substituted；固定輸入重跑須至少兩次相同。原版硬體時間與合成仍為
   hardware-spec approximation，不宣稱逐週期原版硬體或人耳已驗收。

## 驗收與停止線

十一首 FM 原版擷取／合成／固定封裝後，VGM 與 OGG 全檔符合原基準；
三條觸發路徑再抽最小充分重跑，正常入口能生成指定完整曲庫，正式播放層解碼通過。
拒絕錯原版／probe、缺工具、錯曲名、缺錨與錯雜湊，原版來源未變；標準產品測試通過。
不把保存的 VGM／OGG 複製進玩家輸出、不內嵌原版素材到可公開包。
三平台第三方依賴封裝、實際聆聽與 loop 接點仍待獨立驗收，#69 未達整項判準前不關閉。

## 實際驗收

預設完整入口真實產生十一首 GM 與十一首 FM；全部 FM 的 VGM／OGG 全檔與既有
基準相同。BGM_0F、END_2F、OPEN_F 三路徑第二次獨立擷取／合成也逐位元組相同。
正式播放器完整解碼 22 首與正常玩家路徑選曲通過；互斥旗標、錯曲名、缺工具、
原版目錄內輸出、錯 probe 身份、缺原版檔案的 CLI 前置拒絕通過。
封裝錯 CRC／邊界／串接等拒絕由 070 契約測試覆蓋；標準產品測試全綠。
完整來源、版本與命令見[驗證 031](../playtest/031-native-fm-import.md) 及其 JSON 收據。
Windows amd64、macOS amd64／arm64 本體交叉建置通過，未宣稱隨包工具或實機已驗收。
