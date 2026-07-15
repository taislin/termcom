package com.taislin.termcom;

import android.media.AudioAttributes;
import android.media.AudioFormat;
import android.media.AudioTrack;

/**
 * AudioPlayer wraps Android AudioTrack for low-latency PCM audio output.
 * Used by the Go game engine via JNI for sound effect playback.
 *
 * The Go side (oto-based PCM synthesis in audio_other.go) generates 16-bit
 * mono PCM samples at 44100 Hz and writes them here. This class also serves
 * as a standalone utility for Java-side UI sounds.
 */
public class AudioPlayer {

    private static AudioTrack track;
    private static int bufferSize;

    /**
     * Initialises the AudioTrack. Safe to call multiple times (re-init on
     * audio routing changes, e.g. plugging headphones).
     */
    public static void init(int sampleRate, int channels) {
        release();

        int channelConfig = channels == 2
                ? AudioFormat.CHANNEL_OUT_STEREO
                : AudioFormat.CHANNEL_OUT_MONO;

        bufferSize = Math.max(
                AudioTrack.getMinBufferSize(sampleRate, channelConfig, AudioFormat.ENCODING_PCM_16BIT),
                sampleRate * 2 * channels / 25 // 40ms buffer
        );

        AudioAttributes attrs = new AudioAttributes.Builder()
                .setUsage(AudioAttributes.USAGE_GAME)
                .setContentType(AudioAttributes.CONTENT_TYPE_SONIFICATION)
                .build();

        AudioFormat fmt = new AudioFormat.Builder()
                .setSampleRate(sampleRate)
                .setChannelMask(channelConfig)
                .setEncoding(AudioFormat.ENCODING_PCM_16BIT)
                .build();

        track = new AudioTrack.Builder()
                .setAudioAttributes(attrs)
                .setAudioFormat(fmt)
                .setBufferSizeInBytes(bufferSize)
                .setTransferMode(AudioTrack.MODE_STREAM)
                .build();

        if (track.getState() == AudioTrack.STATE_INITIALIZED) {
            track.play();
        } else {
            track = null;
        }
    }

    /**
     * Writes PCM data (16-bit signed little-endian mono) to the audio track.
     * Non-blocking; queues data for playback.
     */
    public static void write(byte[] data) {
        if (track != null && track.getState() == AudioTrack.STATE_INITIALIZED) {
            try {
                track.write(data, 0, data.length);
            } catch (Exception ignored) {
                // Swallow transient write failures
            }
        }
    }

    /**
     * Stops playback and releases the AudioTrack.
     */
    public static void stop() {
        release();
    }

    /**
     * Returns the size of the playback buffer in bytes.
     */
    public static int getBufferSize() {
        return bufferSize;
    }

    /**
     * Returns true if the AudioTrack is actively playing.
     */
    public static boolean isPlaying() {
        return track != null && track.getPlayState() == AudioTrack.PLAYSTATE_PLAYING;
    }

    private static void release() {
        if (track != null) {
            try {
                track.stop();
                track.release();
            } catch (Exception ignored) {}
            track = null;
        }
    }
}
