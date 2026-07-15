package com.taislin.termcom;

import android.content.Context;
import android.graphics.Bitmap;
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
import org.json.JSONArray;
import org.json.JSONObject;

public class TermComView extends SurfaceView implements SurfaceHolder.Callback {

    private static final int CELL_BYTES = 9;

    private final String dataDir;
    private final int initCols;
    private final int initRows;

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
    private Bitmap offscreenBitmap;
    private Canvas offscreenCanvas;

    private static class ViewButton {
        String label;
        boolean enabled;
        int index;
        float left, top, right, bottom;
    }

    private final java.util.List<ViewButton> activeButtons = new java.util.ArrayList<>();
    private int pressedButtonIndex = -1;

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

        recalcGrid();

        getHolder().addCallback(this);
    }

    @Override
    public void surfaceCreated(SurfaceHolder holder) {}

    @Override
    public void surfaceChanged(SurfaceHolder holder, int format, int width, int height) {
        recalcGrid();
        Bridge.resize(gridCols, gridRows);
        forceRedraw = true;
        triggerRedraw();
    }

    @Override
    public void surfaceDestroyed(SurfaceHolder holder) {
        running = false;
    }

    public void start() {
        Bridge.init(dataDir, initCols, initRows);
        Bridge.setFrameListener(new android.FrameListener() {
            @Override
            public void onFrameReady() {
                triggerRedraw();
            }
        });
        Bridge.start();
        running = true;
    }

    public void pause() {
        paused = true;
    }

    public void resume() {
        paused = false;
        forceRedraw = true;
        triggerRedraw();
    }

    public void stop() {
        running = false;
        Bridge.setFrameListener(null);
        Bridge.stop();
        if (offscreenBitmap != null) {
            offscreenBitmap.recycle();
            offscreenBitmap = null;
        }
        offscreenCanvas = null;
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
        
        // Check if the touch event is in the button area
        int hitButtonIndex = -1;
        for (int i = 0; i < activeButtons.size(); i++) {
            ViewButton btn = activeButtons.get(i);
            if (x >= btn.left && x <= btn.right && y >= btn.top && y <= btn.bottom) {
                hitButtonIndex = i;
                break;
            }
        }
        
        if (hitButtonIndex != -1) {
            ViewButton btn = activeButtons.get(hitButtonIndex);
            switch (action) {
                case MotionEvent.ACTION_DOWN:
                    if (btn.enabled) {
                        pressedButtonIndex = hitButtonIndex;
                        triggerRedraw();
                    }
                    break;
                case MotionEvent.ACTION_MOVE:
                    if (pressedButtonIndex != -1 && pressedButtonIndex != hitButtonIndex) {
                        pressedButtonIndex = -1;
                        triggerRedraw();
                    }
                    break;
                case MotionEvent.ACTION_UP:
                    if (pressedButtonIndex == hitButtonIndex && btn.enabled) {
                        if (vibrator != null && vibrator.hasVibrator()) {
                            if (android.os.Build.VERSION.SDK_INT >= 26) {
                                vibrator.vibrate(VibrationEffect.createOneShot(30, VibrationEffect.DEFAULT_AMPLITUDE));
                            } else {
                                vibrator.vibrate(30);
                            }
                        }
                        Bridge.clickButton(btn.index);
                    }
                    pressedButtonIndex = -1;
                    triggerRedraw();
                    break;
                case MotionEvent.ACTION_CANCEL:
                    pressedButtonIndex = -1;
                    triggerRedraw();
                    break;
            }
            return true;
        } else {
            // If dragging outside any button, clear pressed state
            if (pressedButtonIndex != -1) {
                pressedButtonIndex = -1;
                triggerRedraw();
            }
        }
        
        // If we hit outside the buttons, check if it's within the terminal grid
        int gridX = (int) (x / cellW);
        int gridY = (int) (y / cellH);

        if (gridX < 0 || gridX >= gridCols || gridY < 0 || gridY >= gridRows) {
            // Cancel any active button press if finger dragged outside
            if (pressedButtonIndex != -1) {
                pressedButtonIndex = -1;
                triggerRedraw();
            }
            return true;
        }

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

    private synchronized void triggerRedraw() {
        if (paused || !running) return;
        Canvas canvas = null;
        try {
            SurfaceHolder holder = getHolder();
            canvas = holder.lockCanvas();
            if (canvas != null) {
                render(canvas);
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

    private void render(Canvas canvas) {
        int w = getWidth();
        int h = getHeight();
        if (w <= 0 || h <= 0) return;

        int cols = Bridge.frameWidth();
        int rows = Bridge.frameHeight();
        if (cols <= 0 || rows <= 0) return;

        byte[] data = Bridge.frameData();
        if (data == null || data.length < cols * rows * CELL_BYTES) return;

        if (offscreenBitmap == null || offscreenBitmap.getWidth() != w || offscreenBitmap.getHeight() != h) {
            if (offscreenBitmap != null) {
                offscreenBitmap.recycle();
            }
            offscreenBitmap = Bitmap.createBitmap(w, h, Bitmap.Config.ARGB_8888);
            offscreenCanvas = new Canvas(offscreenBitmap);
            forceRedraw = true;
        }

        recalcGrid();

        if (forceRedraw) {
            offscreenCanvas.drawColor(Color.BLACK);
        }

        // Precompute character positioning constants for perfect monospace alignment
        float monospaceW = textPaint.measureText("W");
        Paint.FontMetrics fm = textPaint.getFontMetrics();
        float fontH = fm.bottom - fm.top;
        float textXOffset = (cellW - monospaceW) / 2f;
        float textYOffset = (cellH - fontH) / 2f - fm.top;

        int cellsToDraw = cols * rows;
        for (int i = 0; i < cellsToDraw; i++) {
            int idx = i * CELL_BYTES;
            if (idx + CELL_BYTES > data.length) break;

            int col = i % cols;
            int row = i / cols;

            if (col >= gridCols || row >= gridRows) continue;

            // Check if cell changed since we last updated offscreenCanvas
            boolean cellChanged = forceRedraw || prevFrame == null || prevFrame.length != data.length;
            if (!cellChanged) {
                for (int j = 0; j < CELL_BYTES; j++) {
                    if (data[idx + j] != prevFrame[idx + j]) {
                        cellChanged = true;
                        break;
                    }
                }
            }

            if (!cellChanged) continue;

            int runeLo = data[idx] & 0xFF;
            int runeHi = data[idx + 1] & 0xFF;
            int rune = (runeHi << 8) | runeLo;

            int fgR = data[idx + 2] & 0xFF;
            int fgG = data[idx + 3] & 0xFF;
            int fgB = data[idx + 4] & 0xFF;

            int bgR = data[idx + 5] & 0xFF;
            int bgG = data[idx + 6] & 0xFF;
            int bgB = data[idx + 7] & 0xFF;
            int attr = data[idx + 8] & 0xFF;

            float cx = col * cellW;
            float cy = row * cellH;

            // Draw background
            bgPaint.setColor(Color.rgb(bgR, bgG, bgB));
            offscreenCanvas.drawRect(cx, cy, cx + cellW, cy + cellH, bgPaint);

            if (rune != 0 && rune != ' ') {
                textPaint.setColor(Color.rgb(fgR, fgG, fgB));
                textPaint.setFakeBoldText((attr & 1) != 0);

                String ch = String.valueOf((char) rune);
                offscreenCanvas.drawText(ch, cx + textXOffset, cy + textYOffset, textPaint);
            }
        }

        // Parse active buttons and set their layout boundaries
        activeButtons.clear();
        try {
            JSONArray arr = new JSONArray(Bridge.getButtonsJSON());
            int len = arr.length();

            int buttonCols = 3;
            int buttonRows = (len + buttonCols - 1) / buttonCols;

            float startY = gridRows * cellH + 20;
            float availH = h - startY - 20;
            float availW = w - 40;

            float btnW = availW / buttonCols;
            float btnH = Math.min(120, availH / Math.max(1, buttonRows));

            for (int i = 0; i < len; i++) {
                JSONObject obj = arr.getJSONObject(i);
                ViewButton btn = new ViewButton();
                btn.label = obj.getString("label");
                btn.enabled = obj.getBoolean("enabled");
                btn.index = obj.getInt("index");

                int row = i / buttonCols;
                int col = i % buttonCols;

                btn.left = 20 + col * btnW + 5;
                btn.top = startY + row * btnH + 5;
                btn.right = 20 + (col + 1) * btnW - 5;
                btn.bottom = startY + (row + 1) * btnH - 5;

                activeButtons.add(btn);
            }
        } catch (Exception ignored) {
        }

        // Clear the bottom button area on the offscreen canvas to black
        Paint clearPaint = new Paint();
        clearPaint.setColor(Color.BLACK);
        offscreenCanvas.drawRect(0, gridRows * cellH, w, h, clearPaint);

        // Draw the native buttons
        Paint buttonPaint = new Paint();
        Paint buttonTextPaint = new Paint(Paint.ANTI_ALIAS_FLAG);
        buttonTextPaint.setTypeface(Typeface.create(Typeface.DEFAULT, Typeface.BOLD));

        for (int i = 0; i < activeButtons.size(); i++) {
            ViewButton btn = activeButtons.get(i);

            int bgColor;
            int textColor;
            int borderColor;

            if (!btn.enabled) {
                bgColor = Color.rgb(30, 30, 30);
                textColor = Color.rgb(100, 100, 100);
                borderColor = Color.rgb(50, 50, 50);
            } else if (i == pressedButtonIndex) {
                bgColor = Color.rgb(0, 120, 255);
                textColor = Color.WHITE;
                borderColor = Color.WHITE;
            } else {
                bgColor = Color.rgb(50, 50, 70);
                textColor = Color.WHITE;
                borderColor = Color.rgb(120, 120, 150);
            }

            // Draw background
            buttonPaint.setColor(bgColor);
            buttonPaint.setStyle(Paint.Style.FILL);
            offscreenCanvas.drawRect(btn.left, btn.top, btn.right, btn.bottom, buttonPaint);

            // Draw border
            buttonPaint.setColor(borderColor);
            buttonPaint.setStyle(Paint.Style.STROKE);
            buttonPaint.setStrokeWidth(4);
            offscreenCanvas.drawRect(btn.left, btn.top, btn.right, btn.bottom, buttonPaint);

            // Scale down button font size dynamically if label is too long
            buttonTextPaint.setTextSize(36);
            float textW = buttonTextPaint.measureText(btn.label);
            float maxTextW = (btn.right - btn.left) - 20;
            if (textW > maxTextW) {
                buttonTextPaint.setTextSize(36 * (maxTextW / textW));
            }

            // Draw label centered
            buttonTextPaint.setColor(textColor);
            Rect bounds = new Rect();
            buttonTextPaint.getTextBounds(btn.label, 0, btn.label.length(), bounds);
            float tx = btn.left + (btn.right - btn.left - buttonTextPaint.measureText(btn.label)) / 2f;
            float ty = btn.top + (btn.bottom - btn.top - bounds.height()) / 2f - bounds.top;
            offscreenCanvas.drawText(btn.label, tx, ty, buttonTextPaint);
        }

        // Draw the persistent offscreen bitmap directly to the double/triple-buffered SurfaceView canvas
        canvas.drawBitmap(offscreenBitmap, 0, 0, null);

        // Cache frame for next iteration
        if (prevFrame == null || prevFrame.length != data.length) {
            prevFrame = data.clone();
        } else {
            System.arraycopy(data, 0, prevFrame, 0, data.length);
        }
        forceRedraw = false;
    }

    private void recalcGrid() {
        int viewW = getWidth();
        int viewH = getHeight();
        if (viewW <= 0 || viewH <= 0) return;

        // In portrait mode, lock grid width to a responsive 80 columns
        gridCols = 80;
        
        // Calculate cell width to span the full width of the view
        cellW = (float) viewW / gridCols;
        
        // Use a standard monospace character aspect ratio (height-to-width) of ~1.65
        cellH = cellW * 1.65f;
        
        // Let the grid occupy the top 60% of the screen height, leaving the rest for touch controls later
        gridRows = Math.max(25, (int) ((viewH * 0.6f) / cellH));
        
        // Set the text size to fit the computed cell dimensions without overlapping
        textPaint.setTextSize(cellH * 0.85f);
    }
}
