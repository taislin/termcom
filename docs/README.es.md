# TERMCOM：Un remake de X-COM en ASCII puro

[English](README.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt.md) | [Русский](README.ru.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md)

Un remake estilo *roguelike* de **X-COM: UFO Defense** *(1994, MicroProse)*, renderizado completamente en ASCII en color en una terminal. Escrito en Go con [tcell](https://github.com/gdamore/tcell). Trae la experiencia clásica de estrategia de invasión alienígena a tu terminal. Incluye una implementación completa de todos los bucles de juego: el Geoscape (estrategia global), la gestión de bases y el Battlescape táctico.

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/) [![GitHub License](https://img.shields.io/github/license/taislin/termcom)](https://github.com/taislin/termcom/blob/master/LICENSE) [![Release](https://img.shields.io/github/v/release/taislin/termcom)](https://github.com/taislin/termcom/releases/latest)

[![Download Here](https://img.shields.io/badge/Download%20Here-8A2BE2)](https://github.com/taislin/termcom/releases/latest) [![Website](https://img.shields.io/badge/Website-8A2BE2)](https://taislin.github.io/termcom) [![Manual](https://img.shields.io/badge/Manual-8A2BE2)](https://taislin.github.io/termcom/manual.html) [![Dev Guide](https://img.shields.io/badge/Dev%20Guide-8A2BE2)](docs/dev.md)

## Características

- **Geoscape** — Mapa mundial en tiempo real con compresión de tiempo, rastreo de OVNI e intercepción de cazas.
- **Battlescape** — Combate táctico por turnos con Unidades de Tiempo (TU), cobertura y línea de visión.
- **Gestión de Bases** — Construye instalaciones, contrata soldados, equipa tu escuadrón.
- **Investigación y Fabricación** — Desbloquea tecnología alienígena, fabrica rifles de plasma y trajes de energía.
- **IA Alienígena** — Comportamientos de patrulla, búsqueda, ataque, huida, flanqueo y retirada.
- **Alienígenas Generados Proceduralmente** — Una lista de alienígenas se genera al inicio de cada campaña, cada uno con habilidades, fortalezas, debilidades y armas únicas.
- **Terreno Destructible** — Las granadas destruyen paredes, árboles y rocas, dejando escombros.
- **VFX Dinámico** — Explosiones de partículas, temblor de pantalla, física de escombros, iluminación nocturna.

## Requisitos

> [!TIP]
> Lo siguiente es para construir desde el código fuente. Si solo quieres jugar, puedes descargar los binarios del juego desde [aquí](https://github.com/taislin/termcom/releases/latest).

- Go 1.25 o superior
- Una terminal con soporte de Unicode (para caracteres de dibujo de cajas)

### Solución de problemas de fuentes en la terminal

**termcom** hace un uso intensivo de caracteres Unicode extendidos (Runas, Geometricos y símbulos Etíopes) para renderizar alienígenas y mapas tácticos. La mayoría de los dispositivos deberían soportar los caracteres que usamos, pero si ejecutas el juego de forma nativa en tu terminal y ves caracteres superpuestos, espaciado extraño o cajas vacías (□) en lugar de alienígenas, tu sistema operativo no tiene las fuentes de respaldo necesarias.

#### En Linux

Para solucionarlo en **Ubuntu/Debian**, instala los paquetes de fuentes Noto y el respaldo Unifont:

```bash
sudo apt update
sudo apt install fonts-noto fonts-unifont
```

Para usuarios de **Arch Linux**:

```bash
sudo pacman -S noto-fonts unifont
```

Para usuarios de **Fedora**:

```bash
sudo dnf install google-noto-sans-fonts unifont-fonts
```

#### En macOS

La `Terminal.app` por defecto en macOS a veces puede tener problemas con la alineación de la cuadrícula o renderizar símbolos incorrectamente como emoji de doble ancho.

Para la mejor experiencia, recomendamos altamente usar **[iTerm2](https://iterm2.com/)**. Si te faltan caracteres, puedes instalar GNU Unifont vía Homebrew:

```bash
brew install --cask font-gnu-unifont
```

* **Para corregir la alineación en iTerm2:** Ve a **Settings > Profiles > Text**, marca la casilla *"Use a different font for non-ASCII text"*, y configura esa fuente secundaria a `Unifont`.

#### En Windows

No uses el símbolo del sistema heredado (`cmd.exe`) ni la vieja ventana azul de PowerShell, ya que su soporte de Unicode y color es extremadamente limitado.

1. Usa **[Windows Terminal](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701)** (incluido por defecto en Windows 11, disponible en la Microsoft Store para Windows 10).
2. Si ves cajas vacías `□`, la fuente por defecto de tu sistema no tiene los símbolos requeridos.
3. Descarga e instala una fuente robusta y de alta compatibilidad como **[GNU Unifont](http://unifoundry.com/unifont/)** o **[Noto Sans Mono](https://fonts.google.com/noto/specimen/Noto+Sans+Mono)**.
4. Abre la configuración de Windows Terminal (`Ctrl + ,`), ve a **Profiles > Defaults > Appearance**, y cambia **Font face** a la fuente que acabas de instalar.

## Construir y Ejecutar

### Versión de Terminal

```bash
go run ./cmd/termcom
```

O construye un binario:

```bash
go build -o termcom ./cmd/termcom
./termcom
```

### Versión para Navegador (Experimental)

> [!CAUTION]
> La versión para navegador es experimental y puede tener funcionalidad limitada en comparación con la versión de terminal.

La versión para navegador te permite jugar a termcom en un navegador web usando xterm.js.

1. Inicia el servidor web:

```bash
go run ./cmd/webserver
```

2. Abre tu navegador y navega a:

```
http://localhost:8080
```

La versión para navegador soporta:

- Entrada de teclado completa vía xterm.js
- Comunicación en tiempo real basada en WebSocket
- Redimensión de terminal responsiva
- Todas las funciones del juego (Geoscape, Battlescape, Gestión de Bases)
- **Juego táctil en móvil** — toca para hacer clic, mantén presionado para clic derecho, arrastra para desplazar, menú de control en pantalla con botones según el contexto

### Android Nativo (Experimental)

> [!CAUTION]
> La versión de Android es experimental y puede tener funcionalidad limitada en comparación con la versión de terminal.

El puerto de Android compila el núcleo del juego en Go en una biblioteca nativa `.aar` vía [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile), renderizada como una cuadrícula de caracteres en un `SurfaceView` con entrada táctil completa y audio.

**Prerrequisitos:**

- Go 1.25 o superior
- Android SDK + NDK (API 21 o superior)
- Gradle 8.2 (para construcciones APK locales)
- `gomobile`:
  ```bash
  go install golang.org/x/mobile/cmd/gomobile@latest
  gomobile init
  ```

**Construye la biblioteca del juego:**

```bash
make android-aar
```

Esto escribe `android/app/libs/termcom.aar`.

**Construye un APK (CI / GitHub Actions):**

Un flujo de trabajo `.github/workflows/android-release.yml` construye automáticamente un APK firmado (o APK de depuración) al hacer push a `mobile`/`main`/`master` y en las etiquetas `v*`. Los APK de depuración se producen con cualquier push; etiqueta una versión (`v*`) para publicar un APK firmado como un GitHub Release. Configura estos secretos del repositorio para lanzamientos firmados: `ANDROID_KEYSTORE_BASE64`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS`, `ANDROID_KEY_PASSWORD`.

**Construye un APK localmente:**

```bash
make android-aar                                  # paso 1: Go .aar
cd android && gradle assembleDebug               # paso 2: APK → app/build/outputs/apk/debug/
# o abre android/ en Android Studio y Ejecuta
```

Instala con `adb install android/app/build/outputs/apk/debug/app-debug.apk`.

**Controles:**

- Toca para hacer clic / seleccionar / mover
- Mantén presionado (500ms) para clic derecho / cancelar; vibración al mantener presionado
- Arrastra para desplazar
- Teclado físico soportado (DPAD, Enter, Escape, teclas F)

## Estructura del Proyecto

Consulta el [archivo AGENTS](AGENTS.md) para detalles de la arquitectura.

## Licencia

MIT, ver archivo [LICENSE](LICENSE).

> [!NOTE]
> ***Aviso de uso de IA***: Se utilizó IA en este proyecto para generar y actualizar las traducciones al francés, español, ruso, coreano, chino y japonés. No se generaron audio ni imágenes vía IA (el juego no usa ninguno de todos modos, ¡es totalmente basado en terminal!).
