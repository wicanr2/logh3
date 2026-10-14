"""固定Munt 2.8.2，將保存的原版MIDI事件轉為本機MT-32 OGG曲庫。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import subprocess

def digest(path):
    with Path(path).open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()

def map_channels(raw):
    """Only rewrite channel nibbles; retain delta, note, program and tempo bytes."""
    assert raw[:4] == b'MThd' and struct.unpack_from('>I', raw, 4)[0] == 6
    at, locations, channels = 14, [], set()
    def vlq():
        nonlocal at
        value = 0
        for _ in range(4):
            n = raw[at]; at += 1
            value = value * 128 + (n & 127)
            if n < 128:
                return value
        raise ValueError('invalid MIDI VLQ')
    for _ in range(struct.unpack_from('>H', raw, 10)[0]):
        assert raw[at:at+4] == b'MTrk'
        end = at + 8 + struct.unpack_from('>I', raw, at+4)[0]
        at += 8
        while at < end:
            vlq()
            pos, status = at, raw[at]; at += 1
            if status == 255:
                at += 1; size = vlq(); at += size
            elif status in (240, 247):
                size = vlq(); at += size
            else:
                assert 128 <= status < 240, 'source MIDI uses unexpected running status'
                channel = status & 15
                channels.add(channel); locations.append((pos, channel))
                at += 1 if status >> 4 in (12, 13) else 2
        assert at == end
    assert at == len(raw) and len(channels) <= 8
    mapping = {channel: i+1 for i, channel in enumerate(sorted(channels))}
    result = bytearray(raw)
    for pos, channel in locations:
        result[pos] = (result[pos] & 240) | mapping[channel]
    changed = set(i for i, (x, y) in enumerate(zip(raw, result)) if x != y)
    assert changed <= {pos for pos, _ in locations}
    return bytes(result), mapping

def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--midi', required=True)
    p.add_argument('--rom-dir', default='/fonts')
    p.add_argument('--out', required=True)
    a = p.parse_args()
    source, roms, out = Path(a.midi), Path(a.rom_dir), Path(a.out)
    assert source.is_dir() and roms.is_dir() and not out.exists()
    assert str(out).startswith('/work/') and out.parent.stat().st_uid == os.getuid()
    expected = {'MT32_CONTROL.1987-10-07.v1.07.ROM': 'a73a06c23ed38370e58a11fb1b86f7ea4c547061a60d7aa62bae446235fd2dff',
                'MT32_PCM.ROM': 'd9164063f293410cf33f2f64cdcea6893b44723fe41938e07ec9aef58b406238'}
    for name, sha in expected.items():
        assert digest(roms/name) == sha, 'MT-32 1.07 source identity changed'
    names = [f'BGM_{i}M' for i in range(8)]+['END_1M', 'END_2M', 'OPEN_M']
    assert all((source/(name+'.mid')).is_file() for name in names)
    out.mkdir()
    receipt = {'schema': 'logh3.mt32-music/1', 'munt_commit': '3b05ec276f9e605af86b0eaef7f5eda43477a31f',
               'machine': 'mt32_1_07', 'roms': expected, 'approximation': 'Munt synthesis; remapped melodic channels; original event timing preserved; no real-hardware waveform claim', 'tracks': []}
    for name in names:
        raw = (source/(name+'.mid')).read_bytes()
        adapted, mapping = map_channels(raw)
        midi, wave, ogg = out/(name+'.mid'), out/(name+'.wav'), out/(name+'.ogg')
        midi.write_bytes(adapted)
        command = ['mt32emu-smf2wav', '--rom-dir', str(roms), '--machine-id', 'mt32_1_07',
                   '--sample-rate', '44100', '--src-quality', '3', '--analog-output-mode', '2',
                   '--record-max-start-silence', '-1', '--output', str(wave), str(midi)]
        with (out/(name+'.munt.log')).open('x') as log:
            subprocess.run(command, check=True, stdout=log, stderr=subprocess.STDOUT, timeout=300)
        encode = ['ffmpeg', '-nostdin', '-v', 'error', '-i', str(wave), '-af', 'alimiter=limit=0.84:level=false',
                  '-ar', '44100', '-ac', '2', '-c:a', 'libvorbis', '-q:a', '7', str(ogg)]
        subprocess.run(encode, check=True, timeout=90)
        probe = json.loads(subprocess.check_output(['ffprobe', '-v', 'error', '-show_streams', '-show_format', '-of', 'json', str(ogg)]))
        stream = probe['streams'][0]
        assert int(stream['sample_rate']) == 44100 and stream['channels'] == 2 and stream['codec_name'] == 'vorbis'
        stats = subprocess.run(['ffmpeg', '-nostdin', '-v', 'info', '-i', str(ogg), '-af', 'astats=reset=0', '-f', 'null', '-'], check=True, capture_output=True, text=True, timeout=90).stderr
        peak = float(re.findall(r'Peak level dB: ([-\w.]+)', stats)[-1])
        rms = float(re.findall(r'RMS level dB: ([-\w.]+)', stats)[-1])
        assert peak < -0.2 and rms > -90, (name, peak, rms)
        receipt['tracks'].append({'name': name, 'source_midi_sha256': hashlib.sha256(raw).hexdigest(),
            'adapted_midi_sha256': digest(midi), 'channel_map': mapping, 'commands': [command, encode],
            'ogg_sha256': digest(ogg), 'wave_sha256': digest(wave), 'duration': float(probe['format']['duration']),
            'peak_db': peak, 'rms_db': rms})
        (out/'SOURCES.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
        print(name, 'PASS', receipt['tracks'][-1]['duration'], 'peak', peak, flush=True)
    (out/'NOTICE.txt').write_text('本機原曲MT-32音色。Munt 2.8.2，使用者自備MT-32 1.07 ROM；不公開散布ROM或原曲衍生音訊。\n')

if __name__ == '__main__':
    main()
