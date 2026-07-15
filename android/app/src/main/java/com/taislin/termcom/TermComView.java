package com.taislin.termcom;

import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Color;
import android.graphics.Paint;
import android.graphics.Rect;
import android.graphics.Typeface;
import android.os.VibrationEffect;
import android.os.Vibrator;
import android.view.KeyEvent;
import android.view.MotionEvent;
import android.view.SurfaceHolder;
import android.view.SurfaceView;

public class TermComView extends SurfaceView implements SurfaceHolder.Callback, Runnable {

    private static final int CELL_BYTES = 8;
    private static final long FRAME_TIME_MS = 16;

    private final String dataDir;
    private final int initCols;
    private final int initRows;

    private Thread renderThread;
    private volatile boolean running;
    private volatile boolean paused;

    private int gridCols;
    private int gridRows;
    private float cellW;
    private float cellH;
    private final Paint textPaint;
    private final Paint bgPaint;
    private final Rect textBounds;
    private final Vibrator vibrator;

    private byte[] prevFrame;
    private boolean forceRedraw;

    public TermComView(Context context, String dataDir, int cols, int rows) {
        super(context);
        this.dataDir = dataDir;
        this.initCols = cols;
        this.initRows = rows;
        this.gridCols = cols;
        this.gridRows = rows;

        textPaint = new Paint(Paint.ANTI_ALIAS_FLAG | Paint.SUBPIXEL_TEXT_FLAG);
        textPaint.setTypeface(Typeface.MONOSPACE);
        textPaint.setTextSize(24);

        bgPaint = new Paint();
        textBounds = new Rect();
        forceRedraw = true;

        vibrator = (Vibrator) context.getSystemService(Context.VIBRATOR_SERVICE);

        measureCell();
        recalcGrid();

        getHolder().addCallback(this);
    }

    @Override
    public void surfaceCreated(SurfaceHolder holder) {}

    @Override
    public void surfaceChanged(SurfaceHolder holder, int format, int width, int height) {
        recalcGrid();
        forceRedraw = true;
    }

    @Override
    public void surfaceDestroyed(SurfaceHolder holder) {
        running = false;
    }

    public void start() {
        if (renderThread != null) return;
        Bridge.init(dataDir, initCols, initRows);
        Bridge.start();
        running = true;
        renderThread = new Thread(this, "TermComRender");
        renderThread.start();
    }

    public void pause() {
        paused = true;
    }

    public void resume() {
        paused = false;
    }

    public void stop() {
        running = false;
        Bridge.stop();
        try {
            if (renderThread != null) {
                renderThread.join(1000);
            }
        } catch (InterruptedException ignored) {}
        renderThread = null;
    }

    public void onOrientationChanged(int orientation) {
        recalcGrid();
        forceRedraw = true;
    }

    public void injectKey(String key) {
        Bridge.injectKey(key);
    }

    public String mapKey(int keyCode, KeyEvent event) {
        switch (keyCode) {
            case KeyEvent.KEYCODE_ESCAPE:       return "escape";
            case KeyEvent.KEYCODE_ENTER:        return "enter";
            case KeyEvent.KEYCODE_SPACE:        return "space";
            case KeyEvent.KEYCODE_DEL:          return "backspace";
            case KeyEvent.KEYCODE_TAB:          return "tab";
            case KeyEvent.KEYCODE_DPAD_UP:
            case KeyEvent.KEYCODE_W:            return "up";
            case KeyEvent.KEYCODE_DPAD_DOWN:
            case KeyEvent.KEYCODE_S:            return "down";
            case KeyEvent.KEYCODE_DPAD_LEFT:
            case KeyEvent.KEYCODE_A:            return "left";
            case KeyEvent.KEYCODE_DPAD_RIGHT:
            case KeyEvent.KEYCODE_D:            return "right";
            case KeyEvent.KEYCODE_F1:           return "f1";
            case KeyEvent.KEYCODE_F2:           return "f2";
            case KeyEvent.KEYCODE_F3:           return "f3";
            case KeyEvent.KEYCODE_F4:           return "f4";
            case KeyEvent.KEYCODE_F5:           return "f5";
            case KeyEvent.KEYCODE_F6:           return "f6";
            case KeyEvent.KEYCODE_F7:           return "f7";
            case KeyEvent.KEYCODE_F8:           return "f8";
            case KeyEvent.KEYCODE_F9:           return "f9";
            case KeyEvent.KEYCODE_F10:          return "f10";
            case KeyEvent.KEYCODE_F11:          return "f11";
            case KeyEvent.KEYCODE_F12:          return "f12";
            default:
                if (event.getAction() == KeyEvent.ACTION_DOWN) {
                    int unicode = event.getUnicodeChar();
                    if (unicode > 0 && unicode < 128) {
                        return String.valueOf((char) unicode);
                    }
                }
                return null;
        }
    }

    @Override
    public boolean onTouchEvent(MotionEvent event) {
        int action = event.getActionMasked();
        float x = event.getX();
        float y = event.getY();
        int gridX = (int) (x / cellW);
        int gridY = (int) (y / cellH);

        switch (action) {
            case MotionEvent.ACTION_DOWN:
                downX = gridX;
                downY = gridY;
                downTime = System.currentTimeMillis();
                pressed = true;
                longPressed = false;
                Bridge.injectTouch(gridX, gridY, "tap");
                postDelayed(longPressRunnable, 500);
                break;
            case MotionEvent.ACTION_UP:
                pressed = false;
                removeCallbacks(longPressRunnable);
                if (!longPressed) {
                    Bridge.injectTouch(gridX, gridY, "tap");
                }
                break;
            case MotionEvent.ACTION_MOVE:
                Bridge.injectTouch(gridX, gridY, "move");
                break;
            case MotionEvent.ACTION_CANCEL:
                pressed = false;
                removeCallbacks(longPressRunnable);
                break;
        }
        return true;
    }

    private boolean pressed;
    private boolean longPressed;
    private int downX, downY;
    private long downTime;

    private final Runnable longPressRunnable = new Runnable() {
        @Override
        public void run() {
            if (pressed) {
                longPressed = true;
                Bridge.injectTouch(downX, downY, "long_press");
                if (vibrator != null && vibrator.hasVibrator()) {
                    if (android.os.Build.VERSION.SDK_INT >= 26) {
                        vibrator.vibrate(VibrationEffect.createOneShot(30, VibrationEffect.DEFAULT_AMPLITUDE));
                    } else {
                        vibrator.vibrate(30);
                    }
                }
            }
        }
    };

    @Override
    public void run() {
        while (running) {
            if (!paused) {
                Canvas canvas = null;
                try {
                    SurfaceHolder holder = getHolder();
                    synchronized (holder) {
                        canvas = holder.lockCanvas();
                        if (canvas != null) {
                            render(canvas);
                        }
                    }
                } catch (Exception ignored) {
                } finally {
                    if (canvas != null) {
                        try {
                            getHolder().unlockCanvasAndPost(canvas);
                        } catch (Exception ignored) {}
                    }
                }
            }
            try {
                Thread.sleep(FRAME_TIME_MS);
            } catch (InterruptedException ignored) {
                break;
            }
        }
    }

    private void render(Canvas canvas) {
        int w = getWidth();
        int h = getHeight();
        if (w <= 0 || h <= 0) return;

        int cols = Bridge.frameWidth();
        int rows = Bridge.frameHeight();
        if (cols <= 0 || rows <= 0) return;

        byte[] data = Bridge.frameData();
        if (data == null || data.length < cols * rows * CELL_BYTES) return;

        boolean diff = forceRedraw || prevFrame == null || prevFrame.length != data.length;
        if (!diff) {
            for (int i = 0; i < data.length; i++) {
                if (data[i] != prevFrame[i]) {
                    diff = true;
                    break;
                }
            }
        }

        if (!diff) return;

        recalcGrid();

        if (forceRedraw) {
            canvas.drawColor(Color.BLACK);
        }

        int cellsToDraw = cols * rows;
        for (int i = 0; i < cellsToDraw; i++) {
            int idx = i * CELL_BYTES;
            if (idx + CELL_BYTES > data.length) break;

            int col = i % cols;
            int row = i / cols;

            if (col >= gridCols || row >= gridRows) continue;

            // Skip unchanged cells unless force redraw
            if (!forceRedraw && prevFrame != null) {
                boolean same = true;
                for (int j = 0; j < CELL_BYTES; j++) {
                    if (data[idx + j] != prevFrame[idx + j]) {
                        same = false;
                        break;
                    }
                }
                if (same) continue;
            }

            int runeLo = data[idx] & 0xFF;
            int runeHi = data[idx + 1] & 0xFF;
            int rune = (runeHi << 8) | runeLo;

            int fgR = data[idx + 2] & 0xFF;
            int fgG = data[idx + 3] & 0xFF;
            int fgB = data[idx + 4] & 0xFF;

            int bgR = data[idx + 5] & 0xFF;
            int bgG = data[idx + 6] & 0xFF;
            int attr = data[idx + 7] & 0xFF;

            float cx = col * cellW;
            float cy = row * cellH;

            // Draw background
            if (bgR != 0 || bgG != 0) {
                bgPaint.setColor(Color.rgb(bgR, bgG, 0));
                canvas.drawRect(cx, cy, cx + cellW, cy + cellH, bgPaint);
            } else if (forceRedraw) {
                // Ensure black background on full redraw
                canvas.drawRect(cx, cy, cx + cellW, cy + cellH, bgPaint);
            }

            if (rune != 0 && rune != ' ') {
                textPaint.setColor(Color.rgb(fgR, fgG, fgB));
                textPaint.setFakeBoldText((attr & 1) != 0);

                String ch = String.valueOf((char) rune);
                textPaint.getTextBounds(ch, 0, 1, textBounds);
                float textX = cx + (cellW - textPaint.measureText(ch)) / 2f;
                float textY = cy + (cellH - textBounds.height()) / 2f - textBounds.top;
                canvas.drawText(ch, textX, textY, textPaint);
            }
        }

        // Cache frame for next iteration
        if (prevFrame == null || prevFrame.length != data.length) {
            prevFrame = data.clone();
        } else {
            System.arraycopy(data, 0, prevFrame, 0, data.length);
        }
        forceRedraw = false;
    }

    private void measureCell() {
        textPaint.setTextSize(24);
        cellW = textPaint.measureText("W");
        Paint.FontMetrics fm = textPaint.getFontMetrics();
        cellH = fm.bottom - fm.top + 4;
    }

    private void recalcGrid() {
        int viewW = getWidth();
        int viewH = getHeight();
        if (viewW <= 0 || viewH <= 0) return;

        measureCell();

        gridCols = Math.max(20, (int) (viewW / cellW));
        gridRows = Math.max(10, (int) (viewH / cellH));

        cellW = viewW / (float) gridCols;
        cellH = viewH / (float) gridRows;
        textPaint.setTextSize(cellH * 0.85f);
    }
}
