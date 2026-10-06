#!/usr/bin/env bash
# #92：實際遊戲畫面的60秒靜態剪輯，含原版衍生配樂，只留本機。
# 執行：LOGH3_AUDIO_IMAGE=psychicwar-video:latest bash tools/audio.sh bash /src/tools/scripts/local_test_promo.sh <版本>
set -euo pipefail
export MAGICK_THREAD_LIMIT=2
version="${1:?需版本}"
[[ "$version" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
out="/work/promo92-$version"
test ! -e "$out"
mkdir "$out"
test "$(stat -c %u "$out")" = "$(id -u)"
shots="/src/workplace/package92-smoke-$version"
ai=/src/workplace/active-ai93-r118-b
font=/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc
bold=/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc
test -f "$font"
test -f "$bold"
music="/src/dist-all/$version/full-local/LOGH3-$version-linux-amd64/music/BGM_0F.ogg"
test -f "$music"
card() {
    convert -size 1920x1080 xc:'#081018' -fill '#547bb5' -draw 'rectangle 180,200 1740,204' \
      -font "$bold" -fill '#f3f3ef' -pointsize 94 -gravity center -annotate +0-90 "$2" \
      -font "$font" -pointsize 44 -fill '#85b59b' -annotate +0+64 "$3" \
      -pointsize 28 -fill '#bac5d1' -annotate +0+170 "$4" "$out/$1.png"
}
slide() {
    test -f "$2"
    convert -size 1920x1080 xc:'#081018' \( "$2" -filter point -resize 1280x960 \) \
      -gravity north -geometry +0+24 -composite -font "$font" -fill '#f3f3ef' \
      -pointsize 36 -gravity south -annotate +0+40 "$3" "$out/$1.png"
    printf '%s\t%s\n' "$1" "$2" >> "$out/SOURCES.tsv"
}
card 00 '銀河英雄傳說 III SP' 'remake 本機測試版' "$version"
slide 01 "$shots/main.png" '重現熟悉的選單與繁體中文介面'
slide 02 "$shots/input.png" '星域情報、財政與艦隊指揮'
slide 03 "$shots/ship.png" '從造艦到編組，建立你的艦隊'
slide 04 "$shots/fleet.png" '選擇提督，配置作戰部隊'
slide 05 "$shots/battle.png" '進入戰場，下令、移動與交戰'
slide 06 "$ai/original/turn3/frame.png" '原版 AI 模式：戰略與戰術都已啟用'
slide 07 "$ai/enhanced/turn3/frame.png" '強化 AI 模式：集中兵力與包圍目標'
slide 08 "$shots/load.png" '保存進度，繼續下一場戰役'
card 09 '現在開始試玩' '22 首配樂 · 28 段音效 · 兩種 AI' '本機測試版；人耳與實機驗收待試玩，不宣稱全遊戲原版一致'
: > "$out/concat.txt"
for i in 00 01 02 03 04 05 06 07 08 09; do
    ffmpeg -y -loglevel error -loop 1 -i "$out/$i.png" -t 6 -r 30 \
      -vf 'fade=t=in:st=0:d=0.2,fade=t=out:st=5.8:d=0.2,format=yuv420p' \
      -threads 2 -c:v libx264 -preset veryfast -crf 20 "$out/$i.mp4"
    printf "file '%s/%s.mp4'\n" "$out" "$i" >> "$out/concat.txt"
done
ffmpeg -y -loglevel error -f concat -safe 0 -i "$out/concat.txt" -stream_loop -1 -i "$music" \
  -t 60 -map 0:v -map 1:a -c:v copy -af 'volume=16dB,afade=t=in:st=0:d=1,afade=t=out:st=57:d=3' \
  -c:a aac -b:a 192k -threads 2 -movflags +faststart "$out/LOGH3-$version-promo.mp4"
ffprobe -v error -show_streams -show_format -of json "$out/LOGH3-$version-promo.mp4" > "$out/FFPROBE.json"
ffmpeg -hide_banner -i "$out/LOGH3-$version-promo.mp4" -af volumedetect -vn -f null - > /dev/null 2> "$out/VOLUME.txt"
ffmpeg -hide_banner -i "$out/LOGH3-$version-promo.mp4" -vf 'blackdetect=d=0.25:pix_th=0.02' -an -f null - > /dev/null 2> "$out/BLACK.txt"
for t in 3 15 27 39 45 57; do
    ffmpeg -y -loglevel error -ss "$t" -i "$out/LOGH3-$version-promo.mp4" -frames:v 1 "$out/sample-$t.png"
done
ffmpeg -y -loglevel error -filter_complex_threads 1 \
  -i "$out/sample-3.png" -i "$out/sample-15.png" -i "$out/sample-27.png" \
  -i "$out/sample-39.png" -i "$out/sample-45.png" -i "$out/sample-57.png" \
  -filter_complex '[0:v]scale=640:360[s0];[1:v]scale=640:360[s1];[2:v]scale=640:360[s2];[3:v]scale=640:360[s3];[4:v]scale=640:360[s4];[5:v]scale=640:360[s5];[s0][s1][s2][s3][s4][s5]xstack=inputs=6:layout=0_0|640_0|0_360|640_360|0_720|640_720[v]' \
  -map '[v]' -frames:v 1 -threads 2 "$out/contact-sheet.png"
printf '%s\n' '實際遊戲畫面展示，靜態剪輯；畫面每段6秒為刻意停留，非遊戲凍結。' \
  '配樂BGM_0F.ogg取自本機已驗真實遊戲曲庫；含原版衍生音畫，僅本機測試，不公開散布。' \
  '影片混音增益16dB並淡入淡出；遊戲封包內的OGG保持原檔。' \
  '星圖藍與陣營綠來自本作介面；採軍事情報風格、6秒節奏與系統巡覽，未沿用其他遊戲的紫金或羊皮紙模板。' \
  > "$out/RIGHTS.txt"
sha256sum "$out/LOGH3-$version-promo.mp4" > "$out/SHA256SUMS.txt"
printf '完成60秒本機推廣片：%s\n' "$out"
