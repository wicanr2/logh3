/* 研究 363：隔離檢查 Unicode argv、檔案存在及 UTF-8 → UTF-16 開檔。 */
#include <windows.h>
#include <glib.h>
#include <stdio.h>
#include <stdlib.h>
#include <errno.h>

int main(void) {
    char **args = g_win32_get_command_line();
    int i;
    printf("wchar=%zu\n", sizeof(wchar_t));
    for (i = 1; args[i]; ++i) {
        int n = MultiByteToWideChar(CP_UTF8, MB_ERR_INVALID_CHARS, args[i], -1, NULL, 0);
        wchar_t *wide = malloc(n * sizeof(wchar_t));
        FILE *f;
        HANDLE h;
        MultiByteToWideChar(CP_UTF8, MB_ERR_INVALID_CHARS, args[i], -1, wide, n);
        gunichar2 *gwide = g_utf8_to_utf16(args[i], -1, NULL, NULL, NULL);
        printf("glib_exists=%d regular=%d gwide_attr=%lu\n",
               g_file_test(args[i], G_FILE_TEST_EXISTS), g_file_test(args[i], G_FILE_TEST_IS_REGULAR),
               GetFileAttributesW((wchar_t *)gwide));
        printf("same_wide=%d\n", wcscmp(wide, (wchar_t *)gwide) == 0);
        g_free(gwide);
        printf("arg=%s bytes=", args[i]);
        for (unsigned char *p = (unsigned char *)args[i]; *p; ++p) printf("%02x", *p);
        printf("\nattr=%lu\n", GetFileAttributesW(wide));
        h = CreateFileW(wide, GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING, 0, NULL);
        printf("handle=%p error=%lu\n", h, GetLastError());
        if (h != INVALID_HANDLE_VALUE) CloseHandle(h);
        f = _wfopen(wide, L"rb");
        printf("file=%p errno=%d\n", f, errno);
        if (f) fclose(f);
        free(wide);
    }
    g_strfreev(args);
    return 0;
}
