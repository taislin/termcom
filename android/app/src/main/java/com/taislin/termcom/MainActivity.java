package com.taislin.termcom;

import android.app.Activity;
import android.content.res.Configuration;
import android.os.Bundle;
import android.view.KeyEvent;
import android.view.MotionEvent;
import android.view.View;

public class MainActivity extends Activity {

    private TermComView termComView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        String dataDir = getFilesDir().getAbsolutePath();
        int cols = 80;
        int rows = 25;
        termComView = new TermComView(this, dataDir, cols, rows);
        termComView.setSystemUiVisibility(
            View.SYSTEM_UI_FLAG_FULLSCREEN |
            View.SYSTEM_UI_FLAG_HIDE_NAVIGATION |
            View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY |
            View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN |
            View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION |
            View.SYSTEM_UI_FLAG_LAYOUT_STABLE
        );
        setContentView(termComView);

        termComView.start();
    }

    @Override
    protected void onPause() {
        super.onPause();
        if (termComView != null) {
            termComView.pause();
        }
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (termComView != null) {
            termComView.resume();
        }
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        if (termComView != null) {
            termComView.stop();
        }
    }

    @Override
    public void onConfigurationChanged(Configuration newConfig) {
        super.onConfigurationChanged(newConfig);
        if (termComView != null) {
            termComView.onOrientationChanged(newConfig.orientation);
        }
    }

    @Override
    public boolean onKeyDown(int keyCode, KeyEvent event) {
        if (termComView != null) {
            String key = termComView.mapKey(keyCode, event);
            if (key != null) {
                termComView.injectKey(key);
                return true;
            }
        }
        return super.onKeyDown(keyCode, event);
    }
}
