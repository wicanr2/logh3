#!/usr/bin/env bash
# #95：②版面素材→③FFmpeg；單幀zoompan有精確總幀數，最多2CPU。
# LOGH3_AUDIO_IMAGE=psychicwar-video:latest bash tools/audio.sh bash /src/tools/scripts/retro_promo_render.sh VERSION REV cards|encode|audit
set -euo pipefail
export MAGICK_THREAD_LIMIT=2
export OMP_NUM_THREADS=2
source /src/tools/scripts/retro_promo_theme.sh
version="${1:?需版本}"
revision="${2:?需影片批次}"
mode="${3:?需cards、encode或audit}"
[[ "$version" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
[[ "$revision" =~ ^[a-z0-9]+$ ]]
root="/work/promo95-$version"
input="$root/inputs"
out="$root/render-$revision"
test -f "$root/PLAN.json"
test -f "$FONT"
test -f "$BOLD"
text() {
    convert "$out/$1.png" -font "$BOLD" -fill "$5" -pointsize "$4" \
      -gravity northwest -annotate "+$2+$3" "$6" "$out/$1.png"
}
regular() {
    convert "$out/$1.png" -font "$FONT" -fill "$5" -pointsize "$4" \
      -gravity northwest -annotate "+$2+$3" "$6" "$out/$1.png"
}
shot() {
    convert "$out/$1.png" \( "$input/$2.png" -filter point -resize "$5" \) \
      -gravity northwest -geometry "+$3+$4" -composite "$out/$1.png"
}
rect() {
    convert "$out/$1.png" -fill "$6" -stroke "$LINE" -strokewidth 2 \
      -draw "rectangle $2,$3 $4,$5" "$out/$1.png"
}
base() {
    convert "$out/background.png" "$out/$1.png"
    regular "$1" 90 42 24 "$DIM" '銀河英雄傳說 III SP  /  remake'
    regular "$1" 1630 42 24 "$ACCENT" "$2"
}
footer() {
    regular "$1" 90 1000 36 "$TEXT" "$2"
}
if [[ "$mode" == cards ]]; then
    test ! -e "$out"
    mkdir "$out"
    test "$(stat -c %u "$out")" = "$(id -u)"
    draw=''
    for i in $(seq 0 79); do
        x=$((80+(i*173+31)%1760))
        y=$((130+(i*97+11)%770))
        draw+=" circle $x,$y $((x+1)),$y"
    done
    convert -size "${W}x${H}" "xc:$BG" -fill '#415161' -stroke none -draw "$draw" \
      -stroke "$LINE" -strokewidth 1 -fill none \
      -draw 'line 70,92 1850,92 line 70,962 1850,962 circle 1440,570 1760,570 circle 1440,570 1640,570 line 1440,140 1440,922 line 1010,570 1850,570' \
      "$out/background.png"
    # 標題：大字、星圖母題與寬留白。
    base 00 '指揮席 / 01'
    regular 00 120 264 38 "$ACCENT" '1994 年 PC DOS 中文版重製'
    text 00 114 370 98 "$TEXT" '銀河英雄傳說 III SP'
    regular 00 120 532 46 "$DIM" '銀河的命運，回到你的指揮席'
    convert "$out/00.png" -stroke "$GOLD" -strokewidth 5 -draw 'line 120,638 820,638' "$out/00.png"
    regular 00 120 733 32 "$GOLD" '完整遊戲資料  ·  兩種 AI  ·  三平台試玩'
    footer 00 "$version  本機完整版測試版"
    # 世界觀：兩陣營檔案卡，不借用動畫或其他遊戲素材。
    base 01 '世界觀 / 02'
    text 01 90 142 64 "$TEXT" '兩個國家，一場漫長的戰爭'
    rect 01 90 300 916 870 "$PANEL"
    rect 01 1004 300 1830 870 "$PANEL"
    text 01 130 350 60 "$GOLD" '銀河帝國'
    regular 01 132 462 38 "$TEXT" '帝國的年輕將領'
    text 01 130 538 74 "$TEXT" '萊因哈特'
    regular 01 132 700 34 "$DIM" '從軍的起點，是救回姊姊'
    text 01 1044 350 60 "$ACCENT" '自由惑星同盟'
    regular 01 1046 462 38 "$TEXT" '同盟的艦隊指揮官'
    text 01 1044 538 74 "$TEXT" '楊威利'
    regular 01 1046 700 34 "$DIM" '兩位將領，改變銀河的命運'
    footer 01 '銀河帝國與自由惑星同盟，交戰一百五十年'
    # 情報側欄：完整原圖，畫面與文字不重疊。
    base 02 '星域部署 / 03'
    text 02 90 145 62 "$TEXT" '先看星圖，再下命令'
    shot 02 command 90 292 960x720
    rect 02 1120 292 1830 876 "$PANEL"
    text 02 1170 345 42 "$ACCENT" '作戰前的準備'
    regular 02 1170 465 38 "$TEXT" '星域情報'
    regular 02 1170 565 38 "$TEXT" '財政與生產'
    regular 02 1170 665 38 "$TEXT" '艦隊部署'
    regular 02 1170 783 30 "$DIM" '每一道命令都從這裡開始'
    # 此版面完整畫面到1012，字幕用側欄底部，避免擋住UI。
    regular 02 1130 1000 30 "$DIM" '實際遊戲畫面'
    # 原版對照：同一原生畫布，完整保留，明示不同回合。
    base 03 '畫面對照 / 04'
    text 03 90 145 62 "$TEXT" '從 DOS 畫面，走到新的平台'
    text 03 138 278 42 "$GOLD" 'DOS 原版'
    text 03 1052 278 42 "$ACCENT" '重製引擎'
    shot 03 original 138 376 640x480
    shot 03 battle 1052 376 640x480
    regular 03 842 562 54 "$DIM" '→'
    footer 03 '不同回合的版面展示；未宣稱戰場同態'
    # 艦隊檔案：大畫面與數字索引。
    base 04 '艦隊檔案 / 05'
    text 04 90 145 62 "$TEXT" '你的提督，你的艦隊'
    shot 04 admiral 780 280 960x720
    text 04 90 365 80 "$GOLD" '01'
    regular 04 222 386 40 "$TEXT" '任命將領'
    text 04 90 550 80 "$GOLD" '02'
    regular 04 222 571 40 "$TEXT" '配置副官'
    text 04 90 735 80 "$GOLD" '03'
    regular 04 222 756 40 "$TEXT" '編組部隊'
    # 全畫面：1280×960整數最近鄰，標題／字幕留在兩側。
    base 05 '戰術戰場 / 06'
    shot 05 battle 320 80 1280x960
    regular 05 42 398 42 "$ACCENT" '下令'
    regular 05 42 503 42 "$ACCENT" '移動'
    regular 05 42 608 42 "$ACCENT" '交戰'
    regular 05 40 760 26 "$DIM" '把部署帶進'
    regular 05 40 808 26 "$DIM" '戰術戰場'
    # AI模式：用既有正常兩模式收據，沒有新對拍實驗。
    base 06 '對手選擇 / 07'
    text 06 90 145 62 "$TEXT" '選擇你的對手'
    text 06 138 278 42 "$GOLD" '原版 AI'
    text 06 1052 278 42 "$ACCENT" '強化 AI'
    shot 06 ai-original 138 376 640x480
    shot 06 ai-enhanced 1052 376 640x480
    footer 06 '兩種戰略與戰術均已啟用；保留各自的決策方式'
    # 存檔：橫向資訊卡。
    base 07 '戰役紀錄 / 08'
    text 07 90 145 62 "$TEXT" '每一次決定，都能留下'
    shot 07 load 960 300 800x600
    text 07 100 378 84 "$ACCENT" '保存進度'
    regular 07 104 540 46 "$TEXT" '繼續自己的戰役'
    footer 07 '從主選單取進度，接回你的下一次部署'
    # 聲音：數字為片中實際包內容，不造假的播放器畫面。
    base 08 '遊戲音樂 / 09'
    text 08 90 145 62 "$TEXT" '熟悉的聲音，隨艦隊啟航'
    text 08 160 346 164 "$GOLD" '22'
    text 08 1010 346 164 "$ACCENT" '28'
    regular 08 164 576 54 "$TEXT" '首配樂'
    regular 08 1014 576 54 "$TEXT" '段音效'
    regular 08 164 774 34 "$DIM" '片中配樂取自遊戲的 BGM_0F'
    footer 08 '音樂、音效與必要資料，已放進完整版'
    # 平台：只有支援資訊，沒有冒充macOS實機截圖。
    base 09 '平台封包 / 10'
    text 09 90 145 62 "$TEXT" '三種平台，同一套遊戲機制'
    text 09 150 370 72 "$TEXT" 'Windows'
    text 09 805 370 72 "$TEXT" 'Linux'
    text 09 1390 370 72 "$TEXT" 'macOS'
    regular 09 150 522 34 "$DIM" 'amd64'
    regular 09 805 522 34 "$DIM" 'amd64'
    regular 09 1290 522 34 "$ACCENT" 'Intel / Apple Silicon'
    regular 09 150 744 42 "$TEXT" '完整版測試包，解壓後開始試玩'
    footer 09 'macOS 封包已建置，實機執行與聆聽待試玩'
    # 收尾：可讀呼籲與實際驗收限制，不宣稱公開發行。
    base 10 '開始試玩 / 11'
    text 10 120 332 96 "$TEXT" '下一場戰役，交給你'
    regular 10 124 505 46 "$ACCENT" '完整版測試包，現在開始試玩'
    regular 10 124 680 38 "$DIM" '原版 AI  /  強化 AI'
    regular 10 124 800 30 "$GOLD" '本機測試版；實機、聆聽及完整登陸分段仍待驗收'
    footer 10 "$version"
    cp /src/tools/scripts/retro_promo_theme.sh "$out/theme.sh"
    cp "$root/PLAN.json" "$out/PLAN.json"
    ffmpeg -y -loglevel error -filter_complex_threads 1 \
      -i "$out/00.png" -i "$out/01.png" -i "$out/02.png" -i "$out/03.png" -i "$out/04.png" -i "$out/05.png" \
      -filter_complex '[0:v]scale=480:270[s0];[1:v]scale=480:270[s1];[2:v]scale=480:270[s2];[3:v]scale=480:270[s3];[4:v]scale=480:270[s4];[5:v]scale=480:270[s5];[s0][s1][s2][s3][s4][s5]xstack=inputs=6:layout=0_0|480_0|960_0|0_270|480_270|960_270[v]' \
      -map '[v]' -frames:v 1 -threads 2 "$out/cards-contact.png"
    printf '完成11種版面素材：%s\n' "$out"
    exit 0
fi
[[ "$mode" == encode || "$mode" == audit ]]
test -d "$out"
if [[ "$mode" == encode ]]; then
test ! -e "$out/LOGH3-$version-promo.mp4"
: > "$out/concat.txt"
durations=(7 8 6 8 6 6 8 5 6 5 7)
cache_revision="${4:-}"
if [[ -n "$cache_revision" ]]; then
    [[ "$cache_revision" =~ ^[a-z0-9]+$ ]]
    test "$cache_revision" != "$revision"
fi
: > "$out/CACHE.txt"
for n in $(seq 0 10); do
    id=$(printf '%02d' "$n")
    seconds="${durations[$n]}"
    frames=$((FPS*seconds))
    test -f "$out/$id.png"
    if [[ -n "$cache_revision" ]]; then
        cache="$root/render-$cache_revision"
        if [[ -f "$cache/$id.png" && -f "$cache/$id.mp4" ]] && cmp -s "$out/$id.png" "$cache/$id.png"; then
            probe=$(ffprobe -v error -select_streams v:0 -show_entries stream=width,height,nb_frames,r_frame_rate -of csv=p=0 "$cache/$id.mp4")
            test "$probe" = "$W,$H,$FPS/1,$frames"
            cp "$cache/$id.mp4" "$out/$id.mp4"
            printf '%s %s %s\n' "$id" "$cache_revision" "$(sha256sum "$cache/$id.mp4" | cut -d ' ' -f 1)" >> "$out/CACHE.txt"
            printf "file '%s/%s.mp4'\n" "$out" "$id" >> "$out/concat.txt"
            continue
        fi
    fi
    end=$(awk "BEGIN {print $seconds-0.15}")
    filters="fade=t=in:st=0:d=0.15,fade=t=out:st=$end:d=0.15,format=yuv420p"
    if [[ "$n" == 0 || "$n" == 1 ]]; then
        filters="zoompan=z='min(1+on*0.00008,1.025)':x='iw/2-iw/zoom/2':y='ih/2-ih/zoom/2':d=$frames:s=${W}x${H}:fps=$FPS,$filters"
    fi
    # 不在zoompan之前製造多個輸入幀；每段精確限制輸出幀數。
    ffmpeg -y -loglevel error -loop 1 -i "$out/$id.png" -vf "$filters" \
      -frames:v "$frames" -r "$FPS" -threads 2 -c:v libx264 -preset veryfast -crf 20 "$out/$id.mp4"
    printf "file '%s/%s.mp4'\n" "$out" "$id" >> "$out/concat.txt"
done
ffmpeg -y -loglevel error -f concat -safe 0 -i "$out/concat.txt" -stream_loop -1 -i "$input/BGM_0F.ogg" \
  -t 72 -map 0:v -map 1:a -c:v copy -af 'volume=16dB,afade=t=in:st=0:d=2,afade=t=out:st=69:d=3' \
  -c:a aac -b:a 192k -threads 2 -movflags +faststart "$out/LOGH3-$version-promo.mp4"
ffprobe -v error -show_streams -show_format -of json "$out/LOGH3-$version-promo.mp4" > "$out/FFPROBE.json"
ffmpeg -hide_banner -i "$out/LOGH3-$version-promo.mp4" -af 'volumedetect,silencedetect=n=-70dB:d=0.8' -vn -f null - > /dev/null 2> "$out/AUDIO.txt"
ffmpeg -hide_banner -i "$out/LOGH3-$version-promo.mp4" -vf 'blackdetect=d=0.25:pix_th=0.02,freezedetect=n=-60dB:d=0.8' -an -f null - > /dev/null 2> "$out/VIDEO.txt"
fi
for t in 3 11 18 25 32 38 45 52 57 62 68; do
    ffmpeg -y -loglevel error -ss "$t" -i "$out/LOGH3-$version-promo.mp4" -frames:v 1 "$out/sample-$t.png"
done
ffmpeg -y -loglevel error -filter_complex_threads 1 \
  -threads 1 -i "$out/sample-3.png" -threads 1 -i "$out/sample-11.png" -threads 1 -i "$out/sample-18.png" -threads 1 -i "$out/sample-25.png" \
  -threads 1 -i "$out/sample-32.png" -threads 1 -i "$out/sample-38.png" -threads 1 -i "$out/sample-45.png" -threads 1 -i "$out/sample-52.png" \
  -threads 1 -i "$out/sample-57.png" -threads 1 -i "$out/sample-62.png" -threads 1 -i "$out/sample-68.png" \
  -filter_complex '[0:v]scale=480:270[s0];[1:v]scale=480:270[s1];[2:v]scale=480:270[s2];[3:v]scale=480:270[s3];[4:v]scale=480:270[s4];[5:v]scale=480:270[s5];[6:v]scale=480:270[s6];[7:v]scale=480:270[s7];[8:v]scale=480:270[s8];[9:v]scale=480:270[s9];[10:v]scale=480:270[s10];[s0][s1][s2][s3][s4][s5][s6][s7][s8][s9][s10]xstack=inputs=11:layout=0_0|480_0|960_0|0_270|480_270|960_270|0_540|480_540|960_540|0_810|480_810:fill=0x06151D[v]' \
  -map '[v]' -frames:v 1 -threads 2 "$out/samples-contact.png"
: > "$out/MOTION.txt"
for t in 1 5 9 13; do
    printf 'second=%s\n' "$t" >> "$out/MOTION.txt"
    ffmpeg -loglevel error -ss "$t" -i "$out/LOGH3-$version-promo.mp4" \
      -frames:v 1 -an -threads 2 -f framemd5 - >> "$out/MOTION.txt"
done
old="/src/dist-all/v.0.1.1-20261006/promo"
ffmpeg -y -loglevel error -filter_complex_threads 1 \
  -i "$old/sample-3.png" -i "$old/sample-15.png" -i "$old/sample-27.png" -i "$old/sample-39.png" \
  -i "$out/sample-3.png" -i "$out/sample-11.png" -i "$out/sample-25.png" -i "$out/sample-32.png" \
  -filter_complex '[0:v]scale=480:270[s0];[1:v]scale=480:270[s1];[2:v]scale=480:270[s2];[3:v]scale=480:270[s3];[4:v]scale=480:270[s4];[5:v]scale=480:270[s5];[6:v]scale=480:270[s6];[7:v]scale=480:270[s7];[s0][s1][s2][s3][s4][s5][s6][s7]xstack=inputs=8:layout=0_0|480_0|960_0|1440_0|0_270|480_270|960_270|1440_270[v]' \
  -map '[v]' -frames:v 1 -threads 2 "$out/previous-current.png"
printf '%s\n' '真實DOS原版／remake畫面與正式遊戲配樂；含原版衍生素材，只留本機。' \
  '對照為nearby-state不同回合版面展示，不代表戰場同態；macOS平台卡不冒充實機截圖。' \
  '標題與世界觀兩段為安全單幀zoompan；其餘靜態段為刻意展示，非遊戲凍結。' \
  '影片混音加16dB，遊戲封包音樂不改；無新造配樂。' > "$out/RIGHTS.txt"
sha256sum "$out/LOGH3-$version-promo.mp4" > "$out/SHA256SUMS.txt"
printf '完成72秒復古重製推廣影片：%s\n' "$out"
