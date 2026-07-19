# TERMCOM: 순수 ASCII X-COM 리메이크

[English](README.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt.md) | [Русский](README.ru.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md)

**X-COM: UFO Defense**(1994년, MicroProse)의 *로그라이크* 스타일 리메이크로, 터미널에서 컬러 ASCII로 전적으로 렌더링됩니다. [tcell](https://github.com/gdamore/tcell)을 사용하여 Go로 작성되었습니다. 클래식한 외계인 침입 전략 경험을 터미널로 가져옵니다. 모든 게임 루프를 완전히 구현했습니다: Geoscape(전 세계 전략), 기지 관리, 그리고 전술 Battlescape입니다.

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/) [![GitHub License](https://img.shields.io/github/license/taislin/termcom)](https://github.com/taislin/termcom/blob/master/LICENSE) [![Release](https://img.shields.io/github/v/release/taislin/termcom)](https://github.com/taislin/termcom/releases/latest)

[![Download Here](https://img.shields.io/badge/Download%20Here-8A2BE2)](https://github.com/taislin/termcom/releases/latest) [![Website](https://img.shields.io/badge/Website-8A2BE2)](https://taislin.github.io/termcom) [![Manual](https://img.shields.io/badge/Manual-8A2BE2)](https://taislin.github.io/termcom/manual.html) [![Dev Guide](https://img.shields.io/badge/Dev%20Guide-8A2BE2)](docs/dev.md)

<p align="center">
<a href="screenshots/battlescape.png"><img src="screenshots/battlescape.png" width="40%"/></a> <a href="screenshots/geoscape.png"><img src="screenshots/geoscape.png" width="40%"/></a>
</p>

## 기능

- **Geoscape** — 시간 압축, UFO 추적, 요격기 발진을 갖춘 실시간 세계 지도.
- **Battlescape** — 시간 유닛(TU), 엄호, 시선(LOS)을 갖춘 턴제 전술 전투.
- **기지 관리** — 시설 건조, 용병 고용, 분대 장비.
- **연구 및 제조** — 외계 기술 해금, 플라즈마 소총과 동력 슈트 제조.
- **외계인 AI** — 순찰, 탐색, 공격, 도주, 측면 포위, 후퇴 행동.
- **절차적 생성 외계인** — 각 캠페인 시작 시 고유한 능력, 강점, 약점, 무기를 갖춘 외계인 명단이 생성됨.
- **파괴 가능한 지형** — 수류탄이 벽, 나무, 바위를 파괴하여 잔해물을 남김.
- **동적 VFX** — 입자 폭발, 화면 흔들림, 잔해물 물리, 야간 조명.

## 요구 사항

> [!TIP]
> 아래는 소스 코드에서 빌드하는 방법입니다. 그냥 플레이하고 싶다면 [여기](https://github.com/taislin/termcom/releases/latest)에서 게임 바이너리를 다운로드할 수 있습니다.

- Go 1.25 이상
- 유니코드 지원 터미널 (상자 그리기 문자용)

### 터미널 글꼴 문제 해결

**termcom**은 외계인과 전술 지도를 렌더링하기 위해 확장 유니코드 문자(루나 문자, 기하학 기호, 에티오피아 기호)를 많이 사용합니다. 대부분의 기기는 우리가 사용하는 문자를 지원해야 하지만, 터미널에서 네이티브로 게임을 실행했을 때 문자가 겹치거나, 이상한 간격, 또는 외계인 대신 빈 사각형(□)이 보인다면 시스템에 필요한 폴백 글꼴가 없습니다.

#### Linux에서

**Ubuntu/Debian**에서 이를 해결하려면 Noto 글꼴과 Unifont 폴백 패키지를 설치하세요:

```bash
sudo apt update
sudo apt install fonts-noto fonts-unifont
```

**Arch Linux** 사용자의 경우:

```bash
sudo pacman -S noto-fonts unifont
```

**Fedora** 사용자의 경우:

```bash
sudo dnf install google-noto-sans-fonts unifont-fonts
```

#### macOS에서

macOS의 기본 `Terminal.app`은 그리드 정렬에 문제가 있거나 기호를 두 배 너비 이모지로 잘못 렌더링할 수 있습니다.

최상의 경험을 위해 **[iTerm2](https://iterm2.com/)** 사용을 강력히 권장합니다. 문자가 부족하다면 Homebrew를 통해 GNU Unifont를 설치할 수 있습니다:

```bash
brew install --cask font-gnu-unifont
```

* **iTerm2에서 정렬 수정:** **Settings > Profiles > Text**로 이동하여 *"Use a different font for non-ASCII text"* (비 ASCII 텍스트에 다른 글꼴 사용) 확인란을 체크하고, 해당 보조 글꼴을 `Unifont`로 설정하세요.

#### Windows에서

레거시 명령 프롬프트(`cmd.exe`)나 낡은 파란 PowerShell 창을 사용하지 마세요. 유니코드와 색상 지원이 극히 제한적입니다.

1. **[Windows Terminal](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701)**을 사용하세요 (Windows 11에는 기본 포함, Windows 10용 Microsoft Store에서 사용 가능).
2. 빈 사각형 `□`가 보인다면 시스템 기본 글꼴에 필요한 기호가 없습니다.
3. **[GNU Unifont](http://unifoundry.com/unifont/)** 또는 **[Noto Sans Mono](https://fonts.google.com/noto/specimen/Noto+Sans+Mono)**와 같은 견고하고 호환성 높은 글꼴을 다운로드하여 설치하세요.
4. Windows Terminal 설정(`Ctrl + ,`)을 열고 **Profiles > Defaults > Appearance**로 이동한 다음 **Font face**를 새로 설치한 글꼴으로 변경하세요.

## 빌드 및 실행

### 터미널 버전

```bash
go run ./cmd/termcom
```

또는 바이너리 빌드:

```bash
go build -o termcom ./cmd/termcom
./termcom
```

### 브라우저 버전 (실험적)

> [!CAUTION]
> 브라우저 버전은 실험적이며 터미널 버전에 비해 기능이 제한적일 수 있습니다.
> 
브라우저 버전은 xterm.js를 사용하여 웹 브라우저에서 termcom을 플레이할 수 있게 해줍니다.

1. 웹 서버 시작:

```bash
go run ./cmd/webserver
```

2. 브라우저를 열고 다음으로 이동:

```
http://localhost:8080
```

브라우저 버전은 다음을 지원합니다:

- xterm.js를 통한 완전한 키보드 입력
- WebSocket 기반 실시간 통신
- 반응형 터미널 크기 조절
- 모든 게임 기능 (Geoscape, Battlescape, 기지 관리)
- **모바일 터치 플레이** — 탭하여 클릭, 길게 누르면 우클릭, 드래그하여 스크롤, 상황별 버튼이 있는 화면 제어 메뉴

### Android 네이티브 (실험적)

> [!CAUTION]
> Android 버전은 실험적이며 터미널 버전에 비해 기능이 제한적일 수 있습니다.

Android 이식판은 [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile)을 통해 Go 게임 코어를 네이티브 `.aar` 라이브러리로 컴파일하며, 완전한 터치 입력과 오디오를 갖춘 `SurfaceView`에 문자 그리드로 렌더링됩니다.

**사전 요구 사항:**

- Go 1.25 이상
- Android SDK + NDK (API 21 이상)
- Gradle 8.2 (로컬 APK 빌드용)
- `gomobile`:
  ```bash
  go install golang.org/x/mobile/cmd/gomobile@latest
  gomobile init
  ```

**게임 라이브러리 빌드:**

```bash
make android-aar
```

이렇게 하면 `android/app/libs/termcom.aar`을 작성합니다.

**APK 빌드 (CI / GitHub Actions):**

`.github/workflows/android-release.yml` 워크플로는 `mobile`/`main`/`master`로 푸시하거나 `v*` 태그 시 자동으로 서명된 릴리스 APK(또는 디버그 APK)를 빌드합니다. 디버그 APK는 모든 푸시에서 생성되며, 릴리스에 태그(`v*`)를 지정하면 서명된 APK를 GitHub Release로 게시합니다. 서명된 릴리스를 위해 다음 저장소 비밀을 설정하세요: `ANDROID_KEYSTORE_BASE64`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS`, `ANDROID_KEY_PASSWORD`.

**로컬에서 APK 빌드:**

```bash
make android-aar                                  # 단계 1: Go .aar
cd android && gradle assembleDebug               # 단계 2: APK → app/build/outputs/apk/debug/
# 또는 android/를 Android Studio에서 열고 실행
```

`adb install android/app/build/outputs/apk/debug/app-debug.apk`로 설치하세요.

**조종:**

- 탭하여 클릭 / 선택 / 이동
- 길게 누르기(500ms)로 우클릭 / 취소; 길게 누를 때 진동
- 드래그하여 스크롤
- 하드웨어 키보드 (DPAD, Enter, Escape, F키) 지원

## 프로젝트 구조

아키텍처 세부 사항은 [AGENTS 파일](AGENTS.md)을 참조하세요.

## 라이선스

MIT, [LICENSE](LICENSE) 파일을 참조하세요.

> [!NOTE]
> ***AI 사용 면책 조항***: 이 프로젝트에서는 프랑스어, 스페인어, 러시아어, 한국어, 중국어, 일본어 번역을 생성하고 업데이트하는 데 AI가 사용되었습니다. 오디오나 이미지는 AI를 통해 생성되지 않았습니다 (어차피 게임은 어떤 것도 사용하지 않으며, 전적으로 터미널 기반입니다!).
