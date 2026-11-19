package tw.logh3.app;

import android.app.Activity;
import android.os.Bundle;
import android.view.MotionEvent;
import android.view.View;
import android.view.WindowManager;
import android.widget.TextView;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import go.Seq;
import tw.logh3.mobile.EbitenView;
import tw.logh3.mobile.Mobile;

/**
 * 銀河英雄傳說III SP Android 外殼（規格161）。
 *
 * 完整版的 APK assets 有 full/original 與 full/data；首次啟動解到 App 私有目錄，
 * Go 端用一般檔案路徑讀。release 版沒有 full/，匯入流程在後續階段（A3）。
 */
public final class MainActivity extends Activity {
    // 解開的內容改版時提高這個數字，舊的就會重解。
    private static final String STAGE_STAMP = "full-3"; // 2：加入 music；3：加入 HD 圖

    private EbitenView view;
    private final ExecutorService worker = Executors.newSingleThreadExecutor();

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // gomobile 綁定的程式庫要先拿到 Context：Ebitengine 靠它取螢幕密度。
        // 少了這行，縮放係數是 0，外部尺寸變 +Inf，第一格就 panic、畫面全黑。
        Seq.setContext(getApplicationContext());
        getWindow().addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
        TextView status = new TextView(this);
        status.setText("正在準備遊戲資料，請稍候");
        status.setTextSize(20);
        setContentView(status);
        worker.execute(() -> {
            try {
                File root = new File(getFilesDir(), "full");
                File stamp = new File(root, ".stamp");
                if (!stampMatches(stamp)) {
                    String[] names = getAssets().list("full");
                    if (names == null || names.length == 0) {
                        runOnUiThread(() -> status.setText("這是 release 版：需要匯入原版遊戲檔（尚未實作）。"));
                        return;
                    }
                    deleteTree(root);
                    copyAssetTree("full", root);
                    try (OutputStream out = new FileOutputStream(stamp)) {
                        out.write(stageStamp().getBytes("UTF-8"));
                    }
                }
                File save = new File(getFilesDir(), "save");
                if (!save.isDirectory() && !save.mkdirs()) {
                    throw new IOException("無法建立存檔目錄");
                }
                File music = new File(root, "music");
                File art = new File(root, "art");
                Mobile.start(new File(root, "original").getAbsolutePath(),
                        new File(root, "data").getAbsolutePath(), save.getAbsolutePath(),
                        music.isDirectory() ? music.getAbsolutePath() : "",
                        art.isDirectory() ? art.getAbsolutePath() : "");
                runOnUiThread(this::showGame);
            } catch (IOException e) {
                runOnUiThread(() -> status.setText("遊戲資料準備失敗：" + e.getMessage()));
            }
        });
    }

    private void showGame() {
        view = new EbitenView(this) {
            @Override
            public boolean onTouchEvent(MotionEvent e) {
                // 手指放開時把位置排進 Go 的佇列（規格161）：慢的機器上，輕觸可能整個
                // 落在兩格之間，Ebitengine 的逐格取樣看不到。
                int action = e.getActionMasked();
                if ((action == MotionEvent.ACTION_UP || action == MotionEvent.ACTION_POINTER_UP)
                        && getWidth() > 0 && getHeight() > 0) {
                    int i = e.getActionIndex();
                    Mobile.tap(e.getX(i) / getWidth(), e.getY(i) / getHeight());
                }
                return super.onTouchEvent(e);
            }
        };
        setContentView(view);
        immersive();
    }

    private void immersive() {
        getWindow().getDecorView().setSystemUiVisibility(View.SYSTEM_UI_FLAG_LAYOUT_STABLE
                | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
                | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION | View.SYSTEM_UI_FLAG_FULLSCREEN
                | View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY);
    }

    // 戳記帶 App 的更新時間：覆蓋安裝新版後重新展開素材（存檔在另一個目錄，不受影響）。
    private String stageStamp() {
        long t = 0;
        try {
            t = getPackageManager().getPackageInfo(getPackageName(), 0).lastUpdateTime;
        } catch (android.content.pm.PackageManager.NameNotFoundException e) {
            // 讀不到就只看素材版本
        }
        return STAGE_STAMP + "@" + t;
    }

    private boolean stampMatches(File stamp) {
        if (!stamp.isFile()) {
            return false;
        }
        try (InputStream in = new java.io.FileInputStream(stamp)) {
            byte[] b = new byte[64];
            int n = in.read(b);
            return n > 0 && new String(b, 0, n, "UTF-8").equals(stageStamp());
        } catch (IOException e) {
            return false;
        }
    }

    private void copyAssetTree(String asset, File dest) throws IOException {
        String[] names = getAssets().list(asset);
        if (names != null && names.length > 0) {
            if (!dest.isDirectory() && !dest.mkdirs()) {
                throw new IOException("無法建立目錄 " + dest);
            }
            for (String name : names) {
                copyAssetTree(asset + "/" + name, new File(dest, name));
            }
            return;
        }
        try (InputStream in = getAssets().open(asset); OutputStream out = new FileOutputStream(dest)) {
            byte[] buf = new byte[1 << 16];
            int n;
            while ((n = in.read(buf)) > 0) {
                out.write(buf, 0, n);
            }
        }
    }

    private static void deleteTree(File f) {
        File[] kids = f.listFiles();
        if (kids != null) {
            for (File k : kids) {
                deleteTree(k);
            }
        }
        f.delete();
    }

    @Override
    public void onBackPressed() {
        // 返回手勢＝右鍵（Ebitengine 的 Android 鍵盤表沒有返回鍵，規格161）。
        if (view != null) {
            Mobile.back();
        } else {
            super.onBackPressed();
        }
    }

    @Override
    protected void onPause() {
        super.onPause();
        if (view != null) {
            view.suspendGame();
        }
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (view != null) {
            view.resumeGame();
            immersive();
        }
    }

    @Override
    public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        if (hasFocus && view != null) {
            immersive();
        }
    }
}
