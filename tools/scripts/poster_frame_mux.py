"""166：合併片段後校正60fps時間格，保留每個視訊及AAC封包原始內容。"""
import json
from pathlib import Path
import subprocess

import poster_frame_promo as movie

def packets(path):
    return json.loads(subprocess.check_output(['ffprobe','-v','error','-show_packets',
        '-show_data_hash','sha256','-show_entries','packet=stream_index,pts,dts,duration,data_hash',
        '-of','json',str(path)],text=True,timeout=60))['packets']

def main():
    movie.configure();work=movie.WORK;plan=movie.read(work/'PLAN.json');movie.movie.validate(plan)
    target=work/movie.NAME;before=work/'PRE-GRID.mp4';temporary=work/'GRID.mp4'
    assert not before.exists() and not temporary.exists()
    original=packets(target);video=[p for p in original if p['stream_index']==0]
    count=round(plan['seconds']*60)
    assert len(video)==count and sorted(round(p['pts']/1000)for p in video)==list(range(count))
    # concat demuxer rounds the individual MP4 container durations to ms.
    # The observed drift is below half a frame. Snap PTS and DTS together,
    # preserving the B-frame presentation order and every compressed packet.
    error=max(abs(p['pts']-round(p['pts']/1000)*1000)for p in video)
    assert error<500
    movie.run('ffmpeg','-nostdin','-v','error','-i',target,'-map','0:v:0','-map','0:a:0',
        '-c','copy','-bsf:v','setts=pts=round(PTS/1000)*1000:dts=round(DTS/1000)*1000:duration=1000',
        '-video_track_timescale','60000','-avoid_negative_ts','disabled','-movflags','+faststart',temporary)
    final=packets(temporary)
    for stream in [0,1]:
        assert [p['data_hash']for p in original if p['stream_index']==stream]==[p['data_hash']for p in final if p['stream_index']==stream]
    final_video=[p for p in final if p['stream_index']==0]
    assert sorted(p['pts']for p in final_video)==[n*1000 for n in range(count)]
    assert all(p['duration']==1000 for p in final_video)
    stream=movie.probe(temporary)['streams'][0]
    assert stream['avg_frame_rate']=='60/1'and int(stream['nb_frames'])==count
    assert abs(float(stream['duration'])-count/60)<.000001
    target.rename(before);temporary.rename(target)
    receipt=dict(passed=True,source=movie.movie.record(before),output=movie.movie.record(target),
        frames=count,time_base='1/60000',frame_ticks=1000,maximum_rounding_ticks=error,
        compressed_video_packets_identical=True,aac_packets_identical=True,
        scope='修正容器片段時長四捨五入；未重編碼、增刪影格、加速或改變音軌',
        generator=movie.movie.record(Path(__file__).resolve()))
    movie.write(work/'MUX-GRID.json',receipt)
    for p in [Path(__file__).resolve(),work/'MUX-GRID.json']:
        plan['input_sha256'][str(p.relative_to(movie.ROOT))]=movie.sha(p)
    plan['mux_grid']=receipt
    movie.write(work/'PLAN.json',plan)
    print('PASS exact 60fps',count,'frames; video/AAC packets unchanged; drift',error,'ticks',flush=True)

if __name__=='__main__':main()
