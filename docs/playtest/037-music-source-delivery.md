# 037 — 本作品授權與音樂來源套件驗收

日期：2026-10-02，工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
契約：[規格 077](../spec/077-music-source-delivery.md)；調查：[研究 365](../re/365-music-source-delivery.md)。
已證實（proven）：指定 Go 音樂入口能從來源套件離線重建，五種二進位與同條件對照相同；
來源重建的正常 FM 匯入器通過版本閘門與一首無損基準。不改既有音訊門檻或玩家規則。

## 唯一收據與產物

[music-source-delivery.json](../internals/data/music-source-delivery.json) SHA-256：
`e3b564733896fd65773ef9b82ee700d43b2e4745c1bf151dd91795fdbd62aa58`。
包含來源 HEAD、archive／完整逐檔 manifest、授權範本占位符、vendor 版本、五種程式、
Git 根目錄證明、測試、正常 FM 匯入及歷次失敗日誌雜湊。當輪附件尚未提交，明示 dirty。
logh3 父提交 `0f088320749fb791e5e156f85da3c424264b3185`；
dosgolem 乾淨 `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，未改執行器。

本機來源包 `workplace/music69-delivery/source-kit-vcs-root/music-source-kit.tar`，
2,361 檔、32,819,200 bytes，SHA-256：
`654a292e81afb0596b370337a00c36553786f77085873e2708decf16a07af611`。
獨立 source-kit-vcs-root-repeat 全檔相同；逐項名稱、形態、權限、大小、雜湊與解壓內容均核對。
包含 logh3 的 Go／工具來源與固定 JSON 鎖、完整必要 vendor、dosgolem 完整固定來源樹與
單一提交的淺層 Git 物件，以及 RRSAL-1.0／第三方／Go 授權。原版檔、媒體與曲庫未收錄。
dosgolem 固定樹中的 rich2.test 為既有 Go 測試程式，只為完整提交核對留存，未執行。

| 程式 | 重建與對照共同 SHA-256 |
|---|---|
| Linux music-import | `84572ec6fb023b3925c0b90ba10675ad802a9d1ad99293b0b420ce2b92490d78` |
| Windows music-import | `180549360276f500eb562d7be9942bfc07046601ad9e3a5d6d36d3def181e42d` |
| Linux music-tool | `35f9d2ad38c75b60dac2c8f982460b8d0375cb7551bbf04d0214d0080e68340f` |
| Windows dosgolem-probe | `b714939e3a82d0862647c968d9631dad8a1116705d8d3b8f1cf9263ec77afb32` |
| Linux dosgolem-probe | `0a1a7f7465944edcea218107b4acb0677a1e6b7b9dfd87ec3a6c799952fb6102` |

Go 1.24.13、CGO_ENABLED=0／GOARCH=amd64／-trimpath，兩側工具鏈與旗標相同。
logh3 用 -mod=vendor／-buildvcs=false；目前來源獨立 vendor 與包內逐檔相同。
dosgolem 用 -mod=readonly／-buildvcs=true；fsck 通過、歷史只有固定提交、完整工作樹乾淨，
兩平台 buildinfo 真正記錄 f7523f9／modified=false。對照由專用乾淨副本與同一 image 建置。
先明設 GIT_DIR／GIT_WORK_TREE 並核對 show-toplevel，來源 .git 包含 HEAD、refs、objects；
任何檢查失敗立即停止，避免向上搜尋到外層儲存庫。

來源 CLI 六個測試事件通過、零跳過，music-import 0.136 秒、music-tool 0.004 秒。
TestFMProbeIdentity 使用來源包重建的 Linux probe，正式 checkProbe 實際通過。
四組來源 archive／路徑／形態／重複／缺件／既有輸出／授權範本邊界通過。
十八個標準產品套件報告完整，成功或明示無測試；大部分產品測試使用快取。

## 來源重建正式 FM 抽驗

`tools/scripts/music_source_import_check.py` 先逐項核對驗證 032 的既有 Linux 工具包 BUILD.json，
將合成器／編碼器與隨包依賴複製到獨立中文／空白目錄，以來源重建的匯入器、啟動器及
Linux probe 替換指定程式。工具檔案及替換後身份全記進 START.json／CHECK.json。
容器無網路、沒有系統 ffmpeg，PATH=/none；原版 /orig 唯讀，正式 CLI 無診斷旁路：

```sh
music-import -source /orig -output '匯入 曲庫' -fm-only -tracks BGM_0F
```

實際絕對路徑、工具與完整 dosgolem 旗標保存在唯一收據 formal_fm_sample。
BGM_0F 真實擷取、合成與編碼完成，127.13 秒；完整 VGM／WAV／增益後 PCM 符合
[固定鎖](../internals/data/fm-import-lock.json)，Linux OGG 全檔亦符合既有基準。
這只證明重建程式能走正式匯入入口，不代替人耳、實機、循環或整項正式發行驗收。

## 重生入口與前提

使用既有 Docker 包裝器；所有輸入、輸出與掛載目錄須確認形態及 UID/GID。
已存在來源包與失敗收據不可覆寫，重跑使用完整新批次並同步調整腳本固定位置。
Git 快照為主機控制面，其餘封裝、建置、檢查與匯入全部在容器執行。
RRSAL 完整範本入口：`/home/anr2/.codex/skills/reverse-engineer-retro-game-remake/assets/project-template/LICENSE`，
SHA-256 `cb9f00c0494f0ae87a48dbe15e8911e2e3fc4c64bd47420357ac8cc7c96d8a5c`。
Go notice 取自實際 Go image，沿用[驗證 036](036-windows-combined-music-tools.md)。

```sh
# 初次準備；每個輸出先確認不存在。
git archive --format=tar --output=workplace/music69-delivery/logh3-source-reviewed.tar \
  0f088320749fb791e5e156f85da3c424264b3185 go.mod go.sum cmd internal tools docs/internals/data/fm-import-lock.json
git clone --bare --depth 1 --single-branch --branch logh3-oracle \
  file:///home/anr2/cht/dosgolem-logh3 workplace/music69-delivery/dosgolem-source-git
# 必須核對此分支 HEAD 是上列固定提交；不同則停止，不能以新 HEAD 沿用舊身份。
git --git-dir=workplace/music69-delivery/dosgolem-source-git rev-parse HEAD
git --git-dir=workplace/music69-delivery/dosgolem-source-git archive --format=tar \
  --output=workplace/music69-delivery/dosgolem-source-full.tar f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c
bash tools/gui.sh xrun sh -c 'test ! -e workplace/music69-delivery/go-vendor-candidate && GOPROXY=off go mod vendor -o workplace/music69-delivery/go-vendor-candidate'
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 bash tools/py.sh /work/tools/scripts/music_source_kit.py
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 bash tools/py.sh /work/tools/scripts/music_source_kit.py \
  --output /work/workplace/music69-delivery/source-kit-vcs-root-repeat
bash tools/gui.sh xrun bash tools/scripts/music_source_rebuild.sh \
  > workplace/music69-delivery/source-rebuild-vcs-root.log 2>&1
# 額外掛載來源必須先確認存在、形態與 UID/GID。
DOSGOLEM_GO_IMAGE=logh3-ebiten:2 \
DOSGOLEM_EXTRA_MOUNT=/home/anr2/cht/logh3/workplace/music69-delivery:/out \
DOSGOLEM_GO_CMD=sh bash /home/anr2/cht/dosgolem-logh3/tools/go.sh -c \
  'go version > /out/dosgolem-reference-go-version.txt; for platform in windows linux; do CGO_ENABLED=0 GOOS="$platform" GOARCH=amd64 go build -mod=readonly -trimpath -buildvcs=true -o /out/dosgolem-probe-"$platform"-vcs-reference ./cmd/probe || exit; done; go version -m /out/dosgolem-probe-linux-vcs-reference > /out/dosgolem-reference-vcs-buildinfo.txt'
bash tools/py.sh /work/tools/scripts/music_source_kit_test.py \
  > workplace/music69-delivery/source-boundary-vcs-root-tests.log 2>&1
bash tools/gui.sh test ./internal/... ./cmd/... > workplace/music69-delivery/source-product-tests.log 2>&1
bash tools/go.sh list ./internal/... ./cmd/... > workplace/music69-delivery/source-product-packages.txt
bash tools/py.sh /work/tools/scripts/music_source_import_check.py
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 bash tools/py.sh /work/tools/scripts/music_source_receipt.py
```

來源側空 GOMODCACHE／GOPROXY=off／GOSUMDB=off，尾端確認快取仍空。
固定 Go image：`sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8`。
日誌與完整輸出身份以唯一 JSON 為準。

## 保留反例與限制

首包 1,912 檔：`e0d4987906d619fbae9a1839205d8598dd2f1b8b4937ea9c83d9dfe4305f6968`；
一般模組／vendor 的 h1 註記差異及 JSON 鎖漏項留存。
第二包 1,913 檔：`55387b5c5c4c708fcd6072bbe3b978a33887a30a1efdb61bf15a6b14ea3738f2`；
只能證明編譯，無 VCS probe 被正式匯入器拒絕，原收據保留於 before-vcs。
第三候選 2,360 檔：`1c95d6cb7bd62ca8c6d87c8a69927bc63a56c09285b21ced7e77e83166a4cfa0`；
缺 Git refs 導致找到外層儲存庫，失敗日誌 source-rebuild-vcs-2.log 保留。
完整形成原因見研究 365，未修改舊包或把跳過升格為驗收。

LICENSE SHA-256 `ef3709a8581ce080ee02ee25bab9e0006e767e980440c1ac40eef6a480b34583`，
只有範本占位符改動。私有原版素材及 Git 歷史仍不可因此公開。
合成器／編碼器完整第三方來源另按[033](033-linux-music-sources.md)／[035](035-windows-gm-sources-unicode.md)配送；
vendor 不等於完整執行期重建。正式版號／dist-all、macOS、實機、聆聽／循環未完成，#69 保持開啟。
