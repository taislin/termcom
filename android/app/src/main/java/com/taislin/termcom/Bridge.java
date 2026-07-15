package com.taislin.termcom;

/**
 * Bridge wraps the gomobile-generated GameBridge class with a clean static API.
 * The Go package "android" produces termcom.termcom.android.Android (static factory)
 * and termcom.termcom.android.GameBridge (instance).
 */
public class Bridge {

    private static termcom.termcom.android.GameBridge bridge;

    public static void init(String dataDir, int cols, int rows) {
        if (bridge == null) {
            bridge = termcom.termcom.android.Android.newGame(dataDir, cols, rows);
        }
    }

    public static void start() {
        if (bridge != null) bridge.start();
    }

    public static void stop() {
        if (bridge != null) {
            bridge.stop();
            bridge = null;
        }
    }

    public static void resize(int cols, int rows) {
        if (bridge != null) bridge.resize(cols, rows);
    }

    public static void injectTouch(int x, int y, String action) {
        if (bridge != null) bridge.injectTouch(x, y, action);
    }

    public static void injectKey(String key) {
        if (bridge != null) bridge.injectKey(key);
    }

    public static byte[] frameData() {
        if (bridge != null) return bridge.frameData();
        return new byte[0];
    }

    public static int frameWidth() {
        if (bridge != null) return bridge.frameWidth();
        return 0;
    }

    public static int frameHeight() {
        if (bridge != null) return bridge.frameHeight();
        return 0;
    }

    public static void setLanguage(String lang) {
        if (bridge != null) bridge.setLanguage(lang);
    }
}
