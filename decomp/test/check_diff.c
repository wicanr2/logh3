/*
 * 通用差異檢查：讀 diffprobe 產生的向量檔，在記憶體模型上重跑 C 參考實作，
 * 比對回傳值與全部記憶體寫入（堆疊以外）。
 *
 * 向量檔（文字）：
 *   H func=<id> nargs=<n> ret=<ax|dxax|none> base=<檔名> base_sha256=<hex> stack_lo=<hex> stack_hi=<hex>
 *   # ...                          其餘 # 開頭是 metadata
 *   C <編號> <args,逗號分隔hex|-> <突變數> <addr:val,...|-> <ret hex8> <寫入數> <addr:val,...|-> <步數>
 * addr 是線性位址（hex），val 是 byte（hex）。base 檔是 1 MB 記憶體傾印，與向量檔同目錄。
 * 退出碼：全部相符才是 0。
 */
#define _POSIX_C_SOURCE 200809L
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "gin3_reg.h"

typedef struct {
    const char *id;
    int nargs;
    Gin3Run run;
} Entry;

static Entry g_entries[4096];
static int g_nentries;

void gin3_register(const char *id, int nargs, Gin3Run run)
{
    if (g_nentries < (int)(sizeof g_entries / sizeof g_entries[0])) {
        g_entries[g_nentries].id = id;
        g_entries[g_nentries].nargs = nargs;
        g_entries[g_nentries].run = run;
        g_nentries++;
    }
}

static Mem g_base, g_work, g_pre;

static int cmp_u32(const void *a, const void *b)
{
    uint32_t x = *(const uint32_t *)a, y = *(const uint32_t *)b;
    return x < y ? -1 : x > y;
}

typedef struct {
    uint32_t addr;
    uint8_t val;
} Pair;

static int cmp_pair(const void *a, const void *b)
{
    uint32_t x = ((const Pair *)a)->addr, y = ((const Pair *)b)->addr;
    return x < y ? -1 : x > y;
}

/* 解析 "addr:val,addr:val"；"-" 是空。回傳筆數，失敗回 -1。 */
static long parse_pairs(char *s, Pair **out)
{
    long n = 0, cap = 0;
    Pair *v = NULL;
    if (strcmp(s, "-") == 0) {
        *out = NULL;
        return 0;
    }
    for (char *tok = strtok(s, ","); tok; tok = strtok(NULL, ",")) {
        char *colon = strchr(tok, ':');
        if (!colon)
            return -1;
        *colon++ = '\0';
        if (n == cap) {
            cap = cap ? cap * 2 : 64;
            v = realloc(v, (size_t)cap * sizeof *v);
        }
        v[n].addr = (uint32_t)strtoul(tok, NULL, 16);
        v[n].val = (uint8_t)strtoul(colon, NULL, 16);
        n++;
    }
    *out = v;
    return n;
}

static const char *field(const char *line, const char *key, char *buf, size_t n)
{
    const char *p = strstr(line, key);
    if (!p)
        return NULL;
    p += strlen(key);
    size_t i = 0;
    while (p[i] && p[i] != ' ' && p[i] != '\n' && i + 1 < n) {
        buf[i] = p[i];
        i++;
    }
    buf[i] = '\0';
    return buf;
}

/* ---- 硬體替身的呼叫軌跡 ------------------------------------------------ */
typedef struct {
    char id[32];
    int nargs;
    uint16_t args[16];
    uint32_t ret;
} TraceEnt;

static TraceEnt g_trace[512];
static int g_ntrace, g_tcur;
static int g_trace_flag; /* 本案軌跡不符的旗標：1 多呼叫或參數／順序不同 */
static char g_trace_why[160];

uint32_t gin3_stub(Mem *m, const char *id, int nargs, const uint16_t *args)
{
    TraceEnt *e;
    (void)m;
    if (g_tcur >= g_ntrace) {
        if (!g_trace_flag)
            snprintf(g_trace_why, sizeof g_trace_why, "C 版多呼叫了 %s（原版只有 %d 次呼叫）", id, g_ntrace);
        g_trace_flag = 1;
        return 0;
    }
    e = &g_trace[g_tcur++];
    if (strcmp(e->id, id) != 0 || e->nargs != nargs) {
        if (!g_trace_flag)
            snprintf(g_trace_why, sizeof g_trace_why, "第 %d 次呼叫：C=%s/%d 原版=%s/%d", g_tcur, id, nargs, e->id, e->nargs);
        g_trace_flag = 1;
    } else {
        for (int i = 0; i < nargs; i++) {
            if (e->args[i] != args[i]) {
                if (!g_trace_flag)
                    snprintf(g_trace_why, sizeof g_trace_why, "第 %d 次呼叫 %s 的參數 %d：C=%04X 原版=%04X", g_tcur, id, i, args[i], e->args[i]);
                g_trace_flag = 1;
            }
        }
    }
    return e->ret;
}

/* 解析 "ID:a,b=ret;ID:-=ret"；"-" 是空。成功回 0。 */
static int parse_trace(char *s)
{
    char *save = NULL;
    g_ntrace = g_tcur = 0;
    g_trace_flag = 0;
    g_trace_why[0] = '\0';
    if (!s || strcmp(s, "-") == 0)
        return 0;
    for (char *ent = strtok_r(s, ";", &save); ent && g_ntrace < 512; ent = strtok_r(NULL, ";", &save)) {
        char *colon = strchr(ent, ':'), *eq = strrchr(ent, '=');
        TraceEnt *e;
        if (!colon || !eq || eq < colon)
            return -1;
        *colon = '\0';
        *eq = '\0';
        e = &g_trace[g_ntrace++];
        snprintf(e->id, sizeof e->id, "%s", ent);
        e->nargs = 0;
        e->ret = (uint32_t)strtoul(eq + 1, NULL, 16);
        if (strcmp(colon + 1, "-") != 0) {
            char *p = colon + 1;
            while (*p && e->nargs < 16) {
                e->args[e->nargs++] = (uint16_t)strtoul(p, &p, 16);
                if (*p == ',')
                    p++;
            }
        }
    }
    return 0;
}

int main(int argc, char **argv)
{
    char *line = NULL;
    size_t cap = 0;
    Entry *fn = NULL;
    char ret[16] = "ax", func[64] = "", basefile[256] = "", tmp[64];
    uint32_t stack_lo = 0, stack_hi = 0;
    long total = 0, bad_ret = 0, bad_writes = 0, bad_fmt = 0, bad_trace = 0;
    FILE *f;

    if (argc != 2) {
        fprintf(stderr, "用法: %s 向量檔\n", argv[0]);
        return 2;
    }
    f = fopen(argv[1], "r");
    if (!f) {
        perror(argv[1]);
        return 2;
    }
    while (getline(&line, &cap, f) > 0) {
        if (line[0] == 'H') {
            field(line, "func=", func, sizeof func);
            field(line, "ret=", ret, sizeof ret);
            field(line, "base=", basefile, sizeof basefile);
            if (field(line, "stack_lo=", tmp, sizeof tmp))
                stack_lo = (uint32_t)strtoul(tmp, NULL, 16);
            if (field(line, "stack_hi=", tmp, sizeof tmp))
                stack_hi = (uint32_t)strtoul(tmp, NULL, 16);
            for (int i = 0; i < g_nentries; i++)
                if (strcmp(g_entries[i].id, func) == 0)
                    fn = &g_entries[i];
            if (!fn) {
                fprintf(stderr, "沒有登記的 C 函式：%s\n", func);
                return 2;
            }
            char path[1024];
            const char *slash = strrchr(argv[1], '/');
            snprintf(path, sizeof path, "%.*s%s", slash ? (int)(slash - argv[1] + 1) : 0, argv[1], basefile);
            FILE *bf = fopen(path, "rb");
            if (!bf || fread(g_base.b, 1, GIN3_MEM_SIZE, bf) != GIN3_MEM_SIZE) {
                fprintf(stderr, "讀不到基底記憶體 %s\n", path);
                return 2;
            }
            fclose(bf);
            continue;
        }
        if (line[0] != 'C' || !fn)
            continue;

        char *save = NULL;
        char *tok = strtok_r(line, " \n", &save); /* C */
        long id = atol(strtok_r(NULL, " \n", &save));
        char *argstr = strtok_r(NULL, " \n", &save);
        long nmut = atol(strtok_r(NULL, " \n", &save));
        char *mutstr = strtok_r(NULL, " \n", &save);
        char *retstr = strtok_r(NULL, " \n", &save);
        long nwr = atol(strtok_r(NULL, " \n", &save));
        char *wrstr = strtok_r(NULL, " \n", &save);
        char *stepstr = strtok_r(NULL, " \n", &save);
        char *ntstr = strtok_r(NULL, " \n", &save);
        char *trstr = strtok_r(NULL, " \n", &save);
        (void)stepstr;
        (void)ntstr;
        (void)tok;
        if (!argstr || !mutstr || !retstr || !wrstr) {
            bad_fmt++;
            continue;
        }

        uint16_t args[16] = {0};
        int na = 0;
        if (strcmp(argstr, "-") != 0)
            for (char *a = strtok(argstr, ","); a && na < 16; a = strtok(NULL, ","))
                args[na++] = (uint16_t)strtoul(a, NULL, 16);
        Pair *muts = NULL, *want = NULL;
        long nm = parse_pairs(mutstr, &muts);
        long nw = parse_pairs(wrstr, &want);
        if (nm != nmut || nw != nwr || na != fn->nargs) {
            bad_fmt++;
            free(muts);
            free(want);
            continue;
        }

        memcpy(g_work.b, g_base.b, GIN3_MEM_SIZE);
        for (long i = 0; i < nm; i++)
            g_work.b[muts[i].addr & 0xFFFFF] = muts[i].val;
        memcpy(g_pre.b, g_work.b, GIN3_MEM_SIZE);

        if (parse_trace(trstr) != 0) {
            bad_fmt++;
            free(muts);
            free(want);
            continue;
        }
        uint64_t got = fn->run(&g_work, args);
        if (g_tcur != g_ntrace && !g_trace_flag) {
            snprintf(g_trace_why, sizeof g_trace_why, "C 版少呼叫：只呼叫 %d 次，原版 %d 次", g_tcur, g_ntrace);
            g_trace_flag = 1;
        }
        if (g_trace_flag) {
            if (bad_trace < 10)
                printf("呼叫軌跡不符 case %ld: %s\n", id, g_trace_why);
            bad_trace++;
        }
        uint64_t wantret = strtoull(retstr, NULL, 16);
        uint64_t mask = 0;
        if (strcmp(ret, "dxax") == 0)
            mask = 0xFFFFFFFFull;
        else if (strcmp(ret, "ax") == 0)
            mask = 0xFFFFull;
        else if (strncmp(ret, "pack:", 5) == 0) {
            int nreg = 1;
            for (const char *q = ret + 5; *q; q++)
                nreg += *q == ',';
            mask = nreg >= 4 ? ~0ull : ((1ull << (16 * nreg)) - 1);
        }
        total++;
        if ((got & mask) != (wantret & mask)) {
            if (bad_ret < 10)
                printf("回傳值不符 case %ld: C=%016llX 原版=%016llX (mask %016llX)\n", id, (unsigned long long)got, (unsigned long long)wantret, (unsigned long long)mask);
            bad_ret++;
        }

        /* C 版的寫入集合（堆疊以外；C 版不碰堆疊，保險起見仍排除） */
        Pair *mine = NULL;
        long nmine = 0, mcap = 0;
        for (uint32_t a = 0; a < GIN3_MEM_SIZE; a++) {
            if (g_work.b[a] != g_pre.b[a] && !(a >= stack_lo && a < stack_hi)) {
                if (nmine == mcap) {
                    mcap = mcap ? mcap * 2 : 64;
                    mine = realloc(mine, (size_t)mcap * sizeof *mine);
                }
                mine[nmine].addr = a;
                mine[nmine].val = g_work.b[a];
                nmine++;
            }
        }
        qsort(want, (size_t)nw, sizeof *want, cmp_pair);
        int same = nmine == nw;
        for (long i = 0; same && i < nw; i++)
            same = mine[i].addr == want[i].addr && mine[i].val == want[i].val;
        if (!same) {
            if (bad_writes < 10)
                printf("寫入集合不符 case %ld: C 寫了 %ld 個 byte，原版 %ld 個\n", id, nmine, nw);
            bad_writes++;
        }
        free(mine);
        free(muts);
        free(want);
    }
    (void)cmp_u32;
    free(line);
    fclose(f);

    printf("函式 %s：案數 %ld，回傳值不符 %ld，寫入集合不符 %ld，軌跡不符 %ld，格式錯誤 %ld\n", func, total, bad_ret, bad_writes, bad_trace, bad_fmt);
    if (total == 0) {
        printf("沒有任何案，不能當通過\n");
        return 1;
    }
    return (bad_ret == 0 && bad_writes == 0 && bad_trace == 0 && bad_fmt == 0) ? 0 : 1;
}
