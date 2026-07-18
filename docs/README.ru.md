# TERMCOM: ASCII-ремейк X-COM

[English](README.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt.md) | [Русский](README.ru.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md)

Рогалик-подобный ремейк **X-COM: UFO Defense** *(1994, MicroProse)*, полностью отрисованный цветным ASCII в терминале. Написан на Go с использованием [tcell](https://github.com/gdamore/tcell). Он переносит классический опыт стратегии вторжения пришельцев в ваш терминал. Включает полную реализацию всех игровых циклов: Geoscape (глобальная стратегия), управление базой и тактический Battlescape.

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/) [![GitHub License](https://img.shields.io/github/license/taislin/termcom)](https://github.com/taislin/termcom/blob/master/LICENSE) [![Release](https://img.shields.io/github/v/release/taislin/termcom)](https://github.com/taislin/termcom/releases/latest)

[![Download Here](https://img.shields.io/badge/Download%20Here-8A2BE2)](https://github.com/taislin/termcom/releases/latest) [![Website](https://img.shields.io/badge/Website-8A2BE2)](https://taislin.github.io/termcom) [![Manual](https://img.shields.io/badge/Manual-8A2BE2)](https://taislin.github.io/termcom/manual.html) [![Dev Guide](https://img.shields.io/badge/Dev%20Guide-8A2BE2)](docs/dev.md)

## Особенности

- **Geoscape** — Карта мира в реальном времени со сжатием времени, отслеживанием НЛО и запуском перехватчиков.
- **Battlescape** — Пошаговый тактический бой с единицами времени (ТВ), укрытием и линией видимости.
- **Управление базой** — Стройте сооружения, нанимайте солдат, экипируйте отряд.
- **Исследования и производство** — Открывайте инопланетные технологии, создавайте плазменные винтовки и силовые костюмы.
- **ИИ пришельцев** — Поведение патрулирования, поиска, атаки, бегства, флангового обхода и отступления.
- **Процедурно сгенерированные пришельцы** — Список пришельцев генерируется в начале каждой кампании, каждый со уникальными способностями, сильными и слабыми сторонами, а также оружием.
- **Разрушаемый ландшафт** — Гранаты разрушают стены, деревья и камни, оставляя обломки.
- **Динамические VFX** — Частицы взрывов, тряска экрана, физика обломков, ночное освещение.

## Требования

> [!TIP]
> Ниже описано, как собрать игру из исходного кода. Если вы просто хотите играть, вы можете загрузить исполняемые файлы игры [отсюда](https://github.com/taislin/termcom/releases/latest).

- Go 1.25 или новее
- Терминал с поддержкой Unicode (для отображения символов рисования рамок)

### Устранение неполадок со шрифтами в терминале

**termcom** активно использует расширенные символы Unicode (руны, геометрические и эфиопские символы) для отрисовки пришельцев и тактических карт. Большинство устройств должны поддерживать используемые нами символы, но если вы запустите игру нативно в своём терминале и увидите перекрывающиеся символы, странные интервалы или пустые квадраты (□) вместо пришельцев, значит в вашей системе отсутствуют необходимые резервные шрифты.

#### В Linux

Чтобы исправить это в **Ubuntu/Debian**, установите пакеты шрифтов Noto и резервный Unifont:

```bash
sudo apt update
sudo apt install fonts-noto fonts-unifont
```

Для пользователей **Arch Linux**:

```bash
sudo pacman -S noto-fonts unifont
```

Для пользователей **Fedora**:

```bash
sudo dnf install google-noto-sans-fonts unifont-fonts
```

#### В macOS

Стандартное приложение `Terminal.app` в macOS иногда может испытывать проблемы с выравниванием сетки или некорректно отображать символы как эмодзи двойной ширины.

Для наилучшего опыта мы настоятельно рекомендуем использовать **[iTerm2](https://iterm2.com/)**. Если вам не хватает символов, вы можете установить GNU Unifont через Homebrew:

```bash
brew install --cask font-gnu-unifont
```

* **Чтобы исправить выравнивание в iTerm2:** Перейдите в **Settings > Profiles > Text**, отметьте галочку *"Use a different font for non-ASCII text"*, и установите этот вторичный шрифт на `Unifont`.

#### В Windows

Не используйте устаревшую командную строку (`cmd.exe`) или старое синее окно PowerShell, так как их поддержка Unicode и цветов крайне ограничена.

1. Используйте **[Windows Terminal](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701)** (включён по умолчанию в Windows 11, доступен в Microsoft Store для Windows 10).
2. Если вы видите пустые квадраты `□`, значит в шрифте по умолчанию вашей системы отсутствуют необходимые символы.
3. Загрузите и установите надёжный, высокосовместимый шрифт, например **[GNU Unifont](http://unifoundry.com/unifont/)** или **[Noto Sans Mono](https://fonts.google.com/noto/specimen/Noto+Sans+Mono)**.
4. Откройте настройки Windows Terminal (`Ctrl + ,`), перейдите в **Profiles > Defaults > Appearance**, и измените **Font face** на недавно установленный шрифт.

## Сборка и запуск

### Терминальная версия

```bash
go run ./cmd/termcom
```

Или соберите исполняемый файл:

```bash
go build -o termcom ./cmd/termcom
./termcom
```

### Браузерная версия (экспериментальная)

> [!CAUTION]
> Браузерная версия находится на экспериментальной стадии и может иметь ограниченный функционал по сравнению с терминальной версией.

Браузерная версия позволяет вам играть в termcom в веб-браузере с помощью xterm.js.

1. Запустите веб-сервер:

```bash
go run ./cmd/webserver
```

2. Откройте ваш браузер и перейдите по адресу:

```
http://localhost:8080
```

Браузерная версия поддерживает:

- Полный ввод с клавиатуры через xterm.js
- Связь в реальном времени на базе WebSocket
- Адаптивное изменение размера терминала
- Все игровые функции (Geoscape, Battlescape, Управление базой)
- **Сенсорное управление на мобильных** — нажмите, чтобы кликнуть, долгое нажатие для правой кнопки, перетаскивание для прокрутки, контекстное меню управления на экране

### Нативный Android (экспериментальный)

> [!CAUTION]
> Версия для Android находится на экспериментальной стадии и может иметь ограниченный функционал по сравнению с терминальной версией.

Порт Android компилирует игровое ядро на Go в нативную библиотеку `.aar` через [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile), отрисовываемую в виде сетки символов на `SurfaceView` с полным сенсорным вводом и звуком.

**Предварительные требования:**

- Go 1.25 или новее
- Android SDK + NDK (API 21+)
- Gradle 8.2 (для локальной сборки APK)
- `gomobile`:
  ```bash
  go install golang.org/x/mobile/cmd/gomobile@latest
  gomobile init
  ```

**Сборка игровой библиотеки:**

```bash
make android-aar
```

Это запишет `android/app/libs/termcom.aar`.

**Сборка APK (CI / GitHub Actions):**

Рабочий процесс `.github/workflows/android-release.yml` автоматически собирает подписанный APK (или отладочный APK) при push в `mobile`/`main`/`master` и по тегам `v*`. Отладочные APK создаются при любом push; пометьте релиз (`v*`), чтобы опубликовать подписанный APK как GitHub Release. Установите эти секреты репозитория для подписанных релизов: `ANDROID_KEYSTORE_BASE64`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS`, `ANDROID_KEY_PASSWORD`.

**Сборка APK локально:**

```bash
make android-aar                                  # шаг 1: Go .aar
cd android && gradle assembleDebug               # шаг 2: APK → app/build/outputs/apk/debug/
# или откройте android/ в Android Studio и Запустите
```

Установите с помощью `adb install android/app/build/outputs/apk/debug/app-debug.apk`.

**Управление:**

- Нажмите, чтобы кликнуть / выбрать / переместиться
- Долгое нажатие (500мс) для правой кнопки / отмены; вибрация при долгом нажатии
- Перетаскивание для прокрутки
- Поддерживается аппаратная клавиатура (DPAD, Enter, Escape, F-клавиши)

## Структура проекта

См. файл [AGENTS](AGENTS.md) для подробностей архитектуры.

## Лицензия

MIT, см. файл [LICENSE](LICENSE).

> [!NOTE]
> ***Замечание об использовании ИИ***: ИИ использовалась в этом проекте для генерации и обновления переводов на французский, испанский, русский, корейский, китайский и японский языки. Никакое аудио или изображения не генерировались через ИИ (игра всё равно не использует их — она полностью основана на терминале!).
