"""150：v1.6.0三組海報、五語介面與駐地整軍的全新正常實錄影片。Docker內執行。"""
from pathlib import Path
import math

import garrison_promo_movie as movie

ROOT = Path('/src'); VERSION = 'v.1.6.0-20261024'
SETS = ROOT/'workplace/poster150/player-linux'
GARRISON = ROOT/'workplace/poster150/garrison-linux'
movie.VERSION = VERSION; movie.SOURCE = GARRISON
movie.WORK = ROOT/'workplace/poster150/movie'
movie.OUT = ROOT/'dist-all'/VERSION/'promo'
movie.MOVIE = movie.OUT/('LOGH3-'+VERSION+'-promo.mp4')
movie.WIDE = True
movie.EXTRA_INPUTS = {}
movie.PRESENTATION = '本輪v1.6.0實際Linux完整ZIP正常鍵鼠錄影，1280×720全框縮放為1920×1080。字幕只置中央上、下黑色留白，兩側人物與名言不遮擋；無重畫、裁切或替換遊戲UI。'
movie.CUTS = '三組與五語取同一新局設定切換；前線駐留、抵達有剪輯。改選提督至編成及出發保持原速連續，沒有沿用前版影片。'


def source(folder):
    value = movie.read(folder/'RESULT.json')
    assert value['passed'] and value['normal_entry']
    assert not any(value[k] for k in ['state_injection', 'seed_injection', 'script_flag'])
    assert movie.sha(folder/'RAW.mp4') == value['video_sha256']
    for f in [folder/'RESULT.json', folder/'RAW.mp4', Path(__file__).resolve()]: movie.EXTRA_INPUTS[str(f.relative_to(ROOT))] = movie.sha(f)
    marks = {e.get('name'): (e.get('wall_ns', e.get('wall_begin_ns', 0))-value['record_started_utc_ns'])/1e9 for e in value['events'] if e.get('name')}
    return value, marks, float(movie.probe(folder/'RAW.mp4')['format']['duration'])


def scenes():
    _, sets, sets_seconds = source(SETS); _, garrison, garrison_seconds = source(GARRISON)
    result = [dict(kind='card', duration=4, title='兩位英雄，三種海報', caption='銀河英雄傳說 III SP｜v1.6.0 全新實錄', lines=['三組海報', '五語介面', '前線出擊'])]
    def game(folder, start, duration, title, caption, total):
        duration = math.ceil(duration*60)/60
        assert start >= 0 and start+duration <= total
        result.append(dict(kind='game', source=str(folder/'RAW.mp4'), start=start, duration=duration, raw_duration=total, title=title, caption=caption, lines=[]))
    game(SETS, sets['zh-TW-settings-v1']-.2, 3.0, '從設定開啟海報', '預設關閉；開啟後左帝國、右同盟，中央維持完整4:3。', sets_seconds)
    for name, title, caption in [
        ('zh-TW-map-v1', 'v1｜萊因哈特與楊威利', '兩位主角分站左右，底部斜體名言跟隨遊戲語系。'),
        ('zh-TW-map-v2', 'v2｜兩軍將領', '帝國與同盟將領群像，可隨時從設定切換。'),
        ('zh-TW-map-v3', 'v3｜英雄身邊的人', '安妮羅潔、齊格飛、希爾德；菲列特利加、尤里安與夥伴。'),
    ]: game(SETS, sets[name]-.15, 3.0, title, caption, sets_seconds)
    for lang, caption in [('zh-CN', '简体中文'), ('ja', '日本語'), ('en', 'English'), ('ko', '한국어')]:
        game(SETS, sets[lang+'-settings-v1']-.1, 3.0, '五語介面，名言同步切換', caption+'｜同一局面、同一操作座標與原版存檔格式。', sets_seconds)
        game(SETS, sets[lang+'-map-v1']-.1, 3.0, '五語介面，名言同步切換', caption+'｜同一局面，左右海報名言隨語系更新。', sets_seconds)
    game(GARRISON, max(0, garrison['garrison-orders']-2), 4, '艦隊駐留前線', '先將船艦與提督留在前線星系，再就地整軍。', garrison_seconds)
    game(GARRISON, garrison['option-default-off']-.2, 4, '非戰艦旗艦，由你開啟', '新選項預設關閉；修護艦E與運補艦T始終排除。', garrison_seconds)
    # This source span stays continuous across the three editorial headings.
    first = garrison['first-commander']-1.6
    stock = garrison['local-ship-stock']-1.0
    ready = garrison['fleet-ready']-1.0
    departed = garrison['departure-order']+4.0
    for start, end, title, caption in [
        (first, stock, '從駐地改選提督', '先查看一位提督，再改選派特，組成第3艦隊。'),
        (stock, ready, '領取當地船艦，重新編成', '巡航艦可擔任旗艦；E／T拒絕後繼續編成，不消耗母星船艦。'),
        (ready, departed, '第3艦隊，從前線出發', '確認編成並選擇目的地，下達移動命令。'),
    ]: game(GARRISON, start, end-start, title, caption, garrison_seconds)
    game(GARRISON, garrison['departed']-.2, 1.4, '下一回合，艦隊抵達', '此處剪去回合等待，接續同一局的實際抵達結果。', garrison_seconds)
    result.append(dict(kind='card', duration=4, title='接續你的銀河遠征', caption='三組海報、五語介面｜海報與非戰艦旗艦均預設關閉', lines=['就地領船', '重新整軍', '再次出發']))
    return result


if __name__ == '__main__':
    movie.SCENE_FACTORY = scenes
    movie.main()
