#!/usr/bin/env bash
# Android 建置包裝器（規格161）。一律在 Docker 內執行，主機不裝 SDK／NDK。
#
#   tools/android.sh image          # 建工具鏈映像（需要網路，只在第一次或改 Dockerfile 時）
#   tools/android.sh deps           # 補齊模組快取（需要網路，只在 go.mod 變動後）
#   tools/android.sh aar <輸出目錄>  # 離線綁定 AAR（mobile 套件）
#   tools/android.sh stage <目錄>    # 完整版素材：<目錄>/full/{original,data,music,art}（只留本機）
#   tools/android.sh apk <AAR目錄> <輸出目錄> [完整版素材目錄]
#                                   # Gradle 包 APK；第一次要 LOGH3_ANDROID_NETWORK=default 抓 AGP
#
# 踩過的坑（2026-11-19，第175輪），都已處理在下面：
# - ebitenmobile 要與 go.mod 的 Ebitengine 同版（v2.8.8），Go 要用 1.24：它依賴的
#   golang.org/x/tools v0.25.0 在 Go 1.26 編不過。
# - gomobile 執行 `go list -m -json all` 失敗時會**吞掉錯誤**、寫出空的 go.mod，
#   接著報「missing module declaration」。離線建置前要先 `deps` 把整張模組圖補進快取。
# - 只編 64 位元（arm64、amd64）：規則碼有 `1<<62` 這類假設 int 為 64 位元的寫法，
#   32 位元版即使編得過也可能算出不同結果。
# - gobind 會把 Go 的中文註解複製進 Java，javac 預設 US-ASCII，要設 UTF-8。
# - 被綁定的 mobile 套件只能露出 Java 用得到的函式；殼放在 internal/mobileshell。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${LOGH3_ANDROID_IMAGE:-logh3-android:20261119-r1}"
MODCACHE="$ROOT/workplace/gomodcache"
GOCACHE_DIR="$ROOT/workplace/android-gocache"
mkdir -p "$MODCACHE" "$GOCACHE_DIR"

common=(--rm --memory "${LOGH3_MEM:-6g}" --cpus "${LOGH3_CPUS:-4}" --pids-limit 1024
  --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)"
  -v "$ROOT:/src:ro" -v "$MODCACHE:/gomodcache" -v "$GOCACHE_DIR:/gocache"
  -e GOPATH=/tmp/gopath -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -e HOME=/tmp
  -e GOFLAGS=-mod=readonly -e GOWORK=off -e GOTOOLCHAIN=local
  -e LANG=C.UTF-8 -e LC_ALL=C.UTF-8 -e JAVA_TOOL_OPTIONS=-Dfile.encoding=UTF-8)

case "${1:-}" in
  image)
    mkdir -p "$ROOT/workplace/android-ctx"
    exec timeout 3600 docker build --network default -t "$IMAGE" \
      -f "$ROOT/tools/docker/android.Dockerfile" "$ROOT/workplace/android-ctx"
    ;;
  deps)
    # 有網路、驗 checksum。go.sum 唯讀掛載，不會被改。
    exec timeout 1800 docker run "${common[@]}" \
      -e GOPROXY=https://proxy.golang.org -e GOSUMDB=sum.golang.org "$IMAGE" sh -c '
        cd /src && (go mod download all || true)
        mkdir -p /tmp/m && cd /tmp/m && go mod init tmp >/dev/null 2>&1 &&
        go get github.com/ebitengine/gomobile/cmd/gomobile@v0.0.0-20240911145611-4856209ac325 >/dev/null
        cd /src && GOPROXY=off go list -m -json all >/dev/null && echo "離線模組圖完整"'
    ;;
  aar)
    out="${2:?用法: tools/android.sh aar <輸出目錄>}"
    mkdir -p "$out"
    out="$(cd "$out" && pwd)"
    [[ ! -e "$out/logh3.aar" ]] || { echo "$out/logh3.aar 已存在，拒絕覆寫" >&2; exit 2; }
    exec timeout 1800 docker run "${common[@]}" --network none \
      -e GOPROXY=off -e GOSUMDB=off -v "$out:/out" "$IMAGE" sh -c '
        cd /src && ebitenmobile bind -trimpath -target android/arm64,android/amd64 -androidapi 21 \
          -javapkg tw.logh3 -o /out/logh3.aar ./mobile 2>&1 | grep -v "Picked up JAVA_TOOL_OPTIONS"
        ls -l /out/logh3.aar'
    ;;
  stage)
    # 完整版：原版檔＋執行期資料。data 的檔案清單沿用桌面發行版（v.1.6.5 的 99 檔），
    # 內容取目前的 docs/internals/data；ui-locales.tsv 照桌面打包從 internal/localize/ui.tsv 複製。
    dest="${2:?用法: tools/android.sh stage <目錄>}"
    [[ ! -e "$dest/full" ]] || { echo "$dest/full 已存在，拒絕覆寫" >&2; exit 2; }
    list="$ROOT/workplace/release172/extracted/squashfs-root/usr/bin/data"
    [[ -d "$list" && -d "$ROOT/org_game/extracted" ]] || { echo "缺發行清單或原版檔" >&2; exit 2; }
    mkdir -p "$dest/full/original" "$dest/full/data"
    cp -p "$ROOT"/org_game/extracted/* "$dest/full/original/"
    for f in $(ls "$list"); do
      if [[ "$f" == ui-locales.tsv ]]; then cp -p "$ROOT/internal/localize/ui.tsv" "$dest/full/data/$f"
      else cp -p "$ROOT/docs/internals/data/$f" "$dest/full/data/$f"; fi
    done
    # 原版配樂與音效：桌面完整版音樂目錄根層的 OGG（HD 音色庫的子目錄不放）。
    music="$ROOT/workplace/release172/extracted/squashfs-root/usr/bin/music"
    if [[ -d "$music" ]]; then
      mkdir -p "$dest/full/music"
      cp -p "$music"/*.ogg "$dest/full/music/"
    fi
    # HD 圖：桌面完整版的 art 目錄，海報（畫在畫面兩側，行動版兩側是側欄）不放。
    art="$ROOT/workplace/release172/extracted/squashfs-root/usr/bin/art"
    if [[ -d "$art" ]]; then
      mkdir -p "$dest/full/art"
      (cd "$art" && tar --exclude=./poster -cf - .) | (cd "$dest/full/art" && tar -xf -)
    fi
    echo "HD 圖 $(find "$dest/full/art" -type f 2>/dev/null | wc -l) 檔"
    echo "原版 $(ls "$dest/full/original" | wc -l) 檔、data $(ls "$dest/full/data" | wc -l) 檔、music $(ls "$dest/full/music" 2>/dev/null | wc -l) 檔"
    ;;
  apk)
    aar="${2:?用法: tools/android.sh apk <AAR目錄> <輸出目錄> [完整版素材目錄]}"
    out="${3:?}"; full="${4:-}"
    aar="$(cd "$aar" && pwd)"; mkdir -p "$out"; out="$(cd "$out" && pwd)"
    [[ -f "$aar/logh3.aar" ]] || { echo "缺 $aar/logh3.aar" >&2; exit 2; }
    [[ ! -e "$out/app-build" ]] || { echo "$out/app-build 已存在，拒絕覆寫" >&2; exit 2; }
    state="$ROOT/workplace/android-state"; gradle_home="$ROOT/workplace/android-gradle-home"
    mkdir -p "$state" "$gradle_home"
    extra=(); fullarg=()
    if [[ -n "$full" ]]; then full="$(cd "$full" && pwd)"; extra+=(-v "$full:/full:ro"); fullarg=(-Plogh3FullDir=/full); fi
    net="${LOGH3_ANDROID_NETWORK:-none}"; offline=(--offline); [[ "$net" == none ]] || offline=()
    exec timeout 3600 docker run "${common[@]}" --network "$net" \
      -v "$aar:/aar:ro" -v "$out:/out" -v "$state:/state" -v "$gradle_home:/gradle-home" "${extra[@]}" \
      -e GRADLE_USER_HOME=/gradle-home "$IMAGE" sh -c '
        set -e
        [ -f /state/debug.keystore ] || keytool -genkeypair -keystore /state/debug.keystore -storepass android \
          -keypass android -alias androiddebugkey -keyalg RSA -keysize 2048 -validity 10000 \
          -dname "CN=Android Debug,O=Android,C=US" >/dev/null 2>&1
        gradle '"${offline[*]}"' --no-daemon --stacktrace --max-workers=3 -p /src/android \
          --project-cache-dir /out/gradle-project-cache \
          -Plogh3Aar=/aar/logh3.aar -Plogh3BuildDir=/out/app-build -Plogh3DebugKeystore=/state/debug.keystore \
          '"${fullarg[*]}"' assembleDebug 2>&1 | grep -v "Picked up JAVA_TOOL_OPTIONS" | tail -25
        ls -l /out/app-build/outputs/apk/debug/'
    ;;
  *)
    sed -n '2,12p' "$0"; exit 2 ;;
esac
