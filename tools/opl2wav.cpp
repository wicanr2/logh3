// opl2wav —— 把 OPL（AdLib）的暫存器寫入序列算成 WAV。
//
// 輸入是 AdPlug 認得的 OPL 曲檔；本專案直接使用 dosgolem 產生的 VGM
// （docs/re/275、359 已由 AdPlug 2.3.3 真實實跑確認，不經 DRO 轉換）。
// 輸出是 16 位元立體聲 WAV，交給 ffmpeg 編成 OGG。
//
// **為什麼要自己寫這三十行**：Debian bookworm 有 `libadplug-dev`，
// 但沒有 `adplay`（`adplug-utils` 只裝 `adplugdb`）。也沒有任何一個
// VGM 命令列播放器；函式庫的 factory 可以直接開啟 VGM。
//
//   用法: opl2wav <輸入> <輸出.wav> [取樣率] [上限秒]
//
// 上限秒是護欄：曲子會循環，`update()` 不一定回 false。

#include <adplug/adplug.h>
#include <adplug/emuopl.h>

#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <vector>

static void put32(std::vector<uint8_t> &h, uint32_t v) {
	for (int i = 0; i < 4; i++) h.push_back((v >> (8 * i)) & 0xFF);
}
static void put16(std::vector<uint8_t> &h, uint16_t v) {
	for (int i = 0; i < 2; i++) h.push_back((v >> (8 * i)) & 0xFF);
}

int main(int argc, char **argv) {
	if (argc < 3) {
		fprintf(stderr, "用法: opl2wav <輸入> <輸出.wav> [取樣率] [上限秒]\n");
		return 2;
	}
	const int rate = argc > 3 ? atoi(argv[3]) : 44100;
	const double maxsec = argc > 4 ? atof(argv[4]) : 600.0;

	CEmuopl opl(rate, true /*16 位元*/, true /*立體聲*/);
	CPlayer *p = CAdPlug::factory(argv[1], &opl);
	if (!p) {
		// **認不出來要吵。** 悄悄輸出一份無聲的 WAV 會一路通過後面每一步。
		fprintf(stderr, "AdPlug 認不出這個檔：%s\n", argv[1]);
		return 3;
	}

	std::vector<int16_t> pcm;
	double t = 0;
	while (p->update()) {
		double refresh = p->getrefresh();
		if (refresh <= 0) refresh = 50; // 防呆：0 會除爆
		long n = (long)(rate / refresh);
		if (n < 1) n = 1;
		size_t base = pcm.size();
		pcm.resize(base + (size_t)n * 2); // 立體聲
		opl.update(&pcm[base], (int)n);
		t += (double)n / rate;
		if (t > maxsec) break;
	}
	delete p;

	if (pcm.empty()) {
		fprintf(stderr, "算出來一個取樣都沒有：%s\n", argv[1]);
		return 4;
	}

	std::vector<uint8_t> h;
	const uint32_t bytes = (uint32_t)(pcm.size() * 2);
	h.insert(h.end(), {'R', 'I', 'F', 'F'});
	put32(h, 36 + bytes);
	h.insert(h.end(), {'W', 'A', 'V', 'E', 'f', 'm', 't', ' '});
	put32(h, 16);
	put16(h, 1);                    // PCM
	put16(h, 2);                    // 立體聲
	put32(h, (uint32_t)rate);
	put32(h, (uint32_t)rate * 4);   // 每秒位元組數
	put16(h, 4);                    // 區塊對齊
	put16(h, 16);                   // 位元深度
	h.insert(h.end(), {'d', 'a', 't', 'a'});
	put32(h, bytes);

	FILE *f = fopen(argv[2], "wb");
	if (!f) {
		perror(argv[2]);
		return 5;
	}
	fwrite(h.data(), 1, h.size(), f);
	fwrite(pcm.data(), 1, bytes, f);
	fclose(f);
	fprintf(stderr, "%s → %s（%.1f 秒，%d Hz 立體聲）\n", argv[1], argv[2], t, rate);
	return 0;
}
