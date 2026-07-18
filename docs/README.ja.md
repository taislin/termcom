# TERMCOM：純 ASCII の X-COM リメイク

[English](README.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt.md) | [Русский](README.ru.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md)

**X-COM: UFO Defense**（1994年、MicroProse）の *ローグライク* リメイクで、ターミナル上で色付き ASCII だけで完全に描画されます。[tcell](https://github.com/gdamore/tcell) を用いて Go で記述されています。クラシックなエイリアン侵略ストラテジー体験をターミナルでもたらします。全てのゲームループを完全に実装しています：Geoscape（地球全体の戦略）、基地管理、そして戦術 Battlescape です。

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/) [![GitHub License](https://img.shields.io/github/license/taislin/termcom)](https://github.com/taislin/termcom/blob/master/LICENSE) [![Release](https://img.shields.io/github/v/release/taislin/termcom)](https://github.com/taislin/termcom/releases/latest)

[![Download Here](https://img.shields.io/badge/Download%20Here-8A2BE2)](https://github.com/taislin/termcom/releases/latest) [![Website](https://img.shields.io/badge/Website-8A2BE2)](https://taislin.github.io/termcom) [![Manual](https://img.shields.io/badge/Manual-8A2BE2)](https://taislin.github.io/termcom/manual.html) [![Dev Guide](https://img.shields.io/badge/Dev%20Guide-8A2BE2)](docs/dev.md)

## 機能

- **Geoscape** — 時間圧縮、UFO 追跡、要撃機発進を備えたリアルタイム世界地図。
- **Battlescape** — タイムユニット（TU）、カバー、視線（LOS）を備えたターン制戦術戦闘。
- **基地管理** — 施設の建設、兵士の雇用、部隊の装備。
- **研究と製造** — エイリアン技術のアンロック、プラズマライフルとパワードスーツの製造。
- **エイリアン AI** — 哨戒、探索、攻撃、退却、側面包囲、撤退の挙動。
- **手続き生成されたエイリアン** — 各キャンペェーンの開始時にエイリアンの名簿が生成され、それぞれがユニークな能力、長所、短所、武器を持ちます。
- **破壊可能な地形** — 手榴弾が壁、木、岩を破壊し、瓦礫を残します。
- **動的 VFX** — 粒子爆発、画面揺れ、瓦礫の物理、夜間照明。

## 要件

> [!TIP]
> 以下はソースコードからのビルドについてです。単に遊びたいだけなら、[こちら](https://github.com/taislin/termcom/releases/latest) からゲームのバイナリをダウンロードできます。

- Go 1.25 以上
- Unicode 対応ターミナル（罫引き文字のため）

### ターミナルのフォント・トラブルシューティング

**termcom** はエイリアンや戦術マップの描画に、拡張 Unicode 文字（ルーン文字、幾何学記号、エチオピア記号）を多用しています。私たちが使用する文字は大部分のデバイスでサポートされるはずですが、ターミナル上でネイティブにゲームを起動した際に、文字の重なり、不自然な間隔、あるいはエイリアンの代わりに空の四角（□）が見える場合は、必要なフォールバックフォントがお使いの OS に欠けています。

#### Linux の場合

**Ubuntu/Debian** でこれを修正するには、Noto フォントと Unifont フォールバックパッケージをインストールしてください：

```bash
sudo apt update
sudo apt install fonts-noto fonts-unifont
```

**Arch Linux** ユーザー向け：

```bash
sudo pacman -S noto-fonts unifont
```

**Fedora** ユーザー向け：

```bash
sudo dnf install google-noto-sans-fonts unifont-fonts
```

#### macOS の場合

macOS の標準 `Terminal.app` では、グリッドの整列に問題が生じたり、記号を二重幅の絵文字として誤って描画したりすることがあります。

最良の体験のため、** [iTerm2](https://iterm2.com/)** の使用を強く推奨します。文字が欠けている場合は、Homebrew 経由で GNU Unifont をインストールできます：

```bash
brew install --cask font-gnu-unifont
```

* **iTerm2 で整列を修正するには：** **Settings > Profiles > Text** に進み、*"Use a different font for non-ASCII text"*（非 ASCII テキストに別フォントを使用）のチェックボックスをオンにし、その第 2 フォントを `Unifont` に設定してください。

#### Windows の場合

レガシーコマンドプロンプト（`cmd.exe`）や古い青い PowerShell ウィンドウは、Unicode と色のサポートが極めて限定的であるため使用しないでください。

1. **[Windows Terminal](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701)** を使用してください（Windows 11 では標準搭載、Windows 10 向けには Microsoft Store で入手可能）。
2. 空の四角 `□` が見える場合は、システムの標準フォントに必要な記号が欠けています。
3. **[GNU Unifont](http://unifoundry.com/unifont/)** や **[Noto Sans Mono](https://fonts.google.com/noto/specimen/Noto+Sans+Mono)** のような、堅牢で高互換性を持つフォントをダウンロードしてインストールしてください。
4. Windows Terminal の設定（`Ctrl + ,`）を開き、**Profiles > Defaults > Appearance** に進み、**Font face**（フォント）を新たにインストールしたフォントに変更してください。

## ビルドと実行

### ターミナル版

```bash
go run ./cmd/termcom
```

あるいはバイナリをビルドします：

```bash
go build -o termcom ./cmd/termcom
./termcom
```

### ブラウザ版（実験的）

> [!CAUTION]
> ブラウザ版は実験的であり、ターミナル版と比べて機能が限定的な場合があります。

ブラウザ版では、xterm.js を用いてウェブブラウザ内で termcom を遊ぶことができます。

1. Web サーバーを起動します：

```bash
go run ./cmd/webserver
```

2. ブラウザを開き、以下に移動します：

```
http://localhost:8080
```

ブラウザ版は以下をサポートします：

- xterm.js 経由の完全なキーボード入力
- WebSocket ベースのリアルタイム通信
- レスポンシブなターミナル・リサイズ
- 全ゲーム機能（Geoscape、Battlescape、基地管理）
- **モバイルのタッチ操作** — タップでクリック、長押しで右クリック、ドラッグでスクロール、コンテキストに応じたボタンを持つ画面上制御メニュー

### Android ネイティブ版（実験的）

> [!CAUTION]
> Android 版は実験的であり、ターミナル版と比べて機能が限定的な場合があります。

Android ポートは、Go のゲームコアを [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile) 経由でネイティブな `.aar` ライブラリにコンパイルし、`SurfaceView` 上に文字グリッドとして描画し、完全なタッチ入力と音声を備えます。

**前提条件：**

- Go 1.25 以上
- Android SDK + NDK（API 21 以上）
- Gradle 8.2（ローカル APK ビルド向け）
- `gomobile`：
  ```bash
  go install golang.org/x/mobile/cmd/gomobile@latest
  gomobile init
  ```

**ゲームライブラリのビルド：**

```bash
make android-aar
```

これは `android/app/libs/termcom.aar` を書き出します。

**APK のビルド（CI / GitHub Actions）：**

`.github/workflows/android-release.yml` ワークフローは、`mobile`/`main`/`master` へのプッシュ時および `v*` タグ時に、署名付きリリース APK（またはデバッグ APK）を自動ビルドします。デバッグ APK は任意のプッシュから生成されます；リリースにタグ（`v*`）を打つと、署名付き APK を GitHub Release として公開します。署名付きリリースには以下のリポジトリ・シークレットを設定してください：`ANDROID_KEYSTORE_BASE64`、`ANDROID_KEYSTORE_PASSWORD`、`ANDROID_KEY_ALIAS`、`ANDROID_KEY_PASSWORD`。

**ローカルでの APK ビルド：**

```bash
make android-aar                                  # ステップ 1: Go .aar
cd android && gradle assembleDebug               # ステップ 2: APK → app/build/outputs/apk/debug/
# または android/ を Android Studio で開いて Run
```

`adb install android/app/build/outputs/apk/debug/app-debug.apk` でインストールします。

**操作：**

- タップでクリック / 選択 / 移動
- 長押し（500ms）で右クリック / キャンセル；長押し時にバイブレーション
- ドラッグでスクロール
- ハードウェア・キーボード（DPAD、Enter、Escape、Fキー）対応

## プロジェクト構造

アーキテクチャの詳細については [AGENTS ファイル](AGENTS.md) を参照してください。

## ライセンス

MIT、[LICENSE](LICENSE) ファイルを参照。

> [!NOTE]
> ***AI 使用に関する免責事項***：本プロジェクトでは、フランス語、スペイン語、ロシア語、韓国語、中国語、日本語への翻訳の生成と更新に AI を使用しました。音声や画像が AI 経由で生成されることはありません（そもそも本ゲームはいずれも使用せず、完全にターミナルベースです！）。
