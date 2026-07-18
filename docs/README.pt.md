# TERMCOM: Um remake de X-COM em ASCII puro

[English](README.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt.md) | [Русский](README.ru.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md)

Um remake estilo *roguelike* de **X-COM: UFO Defense** *(1994, MicroProse)*, renderizado inteiramente em ASCII colorido em um terminal. Escrito em Go com [tcell](https://github.com/gdamore/tcell). Traz a experiência clássica de estratégia de invasão alienígena para o seu terminal. Inclui uma implementação completa de todos os lopos de jogo: o Geoscape (estratégia global), o gerenciamento de base e o Battlescape tático.

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/) [![GitHub License](https://img.shields.io/github/license/taislin/termcom)](https://github.com/taislin/termcom/blob/master/LICENSE) [![Release](https://img.shields.io/github/v/release/taislin/termcom)](https://github.com/taislin/termcom/releases/latest)

[![Download Here](https://img.shields.io/badge/Download%20Here-8A2BE2)](https://github.com/taislin/termcom/releases/latest) [![Website](https://img.shields.io/badge/Website-8A2BE2)](https://taislin.github.io/termcom) [![Manual](https://img.shields.io/badge/Manual-8A2BE2)](https://taislin.github.io/termcom/manual.html) [![Dev Guide](https://img.shields.io/badge/Dev%20Guide-8A2BE2)](docs/dev.md)

## Recursos

- **Geoscape** — Mapa mundial em tempo real com compressão de tempo, rastreamento de OVNIs e lançamento de interceptadores.
- **Battlescape** — Combate tático por turnos com Unidades de Tempo (TU), cobertura e linha de visão.
- **Gerenciamento de Base** — Constroi instalações, contrata soldados, equipe seu esquadrão.
- **Pesquisa e Fabricação** — Desbloqueia tecnologia alienígena, fabrica rifles de plasma e trajes de energia.
- **IA Alienígena** — Comportamentos de patrulha, busca, ataque, fuga, flanqueio e retirada.
- **Alienígenas Gerados Proceduralmente** — Um rol de alienígenas é gerado no início de cada campanha, cada um com habilidades, forças, fraquezas e armas únicas.
- **Terreno Destrutível** — Granadas destroem paredes, árvores e rochas, deixando escombros.
- **VFX Dinâmico** — Explosões de partículas, tremida de tela, física de escombros, iluminação noturna.

## Requisitos

> [!TIP]
> O que segue é para construir a partir do código-fonte. Se você apenas quer jogar, pode baixar os binários do jogo a partir daqui.

- Go 1.25 ou superior
- Um terminal com suporte a Unicode (para caracteres de desenho de caixas)

### Solução de problemas de fontes no terminal

**termcom** faz uso intensivo de caracteres Unicode estendidos (Runas, Geométricos e símbolos Etíopes) para renderizar alienígenas e mapas táticos. A maioria dos dispositivos deve suprir os caracteres que usamos, mas se você executar o jogo nativamente no seu terminal e vir caracteres sobrepostos, espaçamento estranho ou caixas vazias (□) em vez de alienígenas, seu sistema operacional está sem as fontes de recuo necessárias.

#### No Linux

Para corrigir isso no **Ubuntu/Debian**, instale os pacotes de fontes Noto e o recuo Unifont:

```bash
sudo apt update
sudo apt install fonts-noto fonts-unifont
```

Para usuários do **Arch Linux**:

```bash
sudo pacman -S noto-font-s unifont
```

Para usuários do **Fedora**:

```bash
sudo dnf install google-noto-sans-fonts unifont-fonts
```

#### No macOS

O `Terminal.app` padrão no macOS às vezes pode ter problemas com o alinhamento da grade ou renderizar símbolos incorretamente como emojis de dupla largura.

Para a melhor experiencia, recomendamos fortemente usar o **[iTerm2](https://iterm2.com/)**. Se estiverem faltando caracteres, você pode instalar o GNU Unifont via Homebrew:

```bash
brew install --cask font-gnu-unifont
```

* **Para corrigir o alinhamento no iTerm2:** Vá em **Settings > Profiles > Text**, marque a caixa *"Use a different font for non-ASCII text"*, e defina essa fonte secundária como `Unifont`.

#### No Windows

Não use o prompt de comando legado (`cmd.exe`) nem a antiga janela azul do PowerShell, pois seu suporte a Unicode e cores é extremamente limitado.

1. Use o **[Windows Terminal](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701)** (incluído por padrão no Windows 11, disponível na Microsoft Store para o Windows 10).
2. Se você vir caixas vazias `□`, a fonte padrão do seu sistema não tem os símbolos requeridos.
3. Baixe e instale uma fonte robusta e de alta compatibilidade como **[GNU Unifont](http://unifoundry.com/unifont/)** ou **[Noto Sans Mono](https://fonts.google.com/noto/specimen/Noto+Sans+Mono)**.
4. Abra as configurações do Windows Terminal (`Ctrl + ,`), vá em **Profiles > Defaults > Appearance**, e mude **Font face** para a fonte recém-instalada.

## Construir e Executar

### Versão de Terminal

```bash
go run ./cmd/termcom
```

Ou construa um binário:

```bash
go build -o termcom ./cmd/termcom
./termcom
```

### Versão para Navegador (Experimental)

> [!CAUTION]
> A versão para navegador é experimental e pode ter funcionalidade limitada comparada à versão de terminal.

A versão para navegador permite que você jogue termcom em um navegador web usando xterm.js.

1. Inicie o servidor web:

```bash
go run ./cmd/webserver
```

2. Abra seu navegador e navegue para:

```
http://localhost:8080
```

A versão para navegador suporta:

- Entrada de teclado completa via xterm.js
- Comunicação em tempo real baseada em WebSocket
- Redimensionamento responsivo do terminal
- Todos os recursos do jogo (Geoscape, Battlescape, Gerenciamento de Base)
- **Jogo por toque em mobile** — toque para clicar, pressione longamente para clicar com o botão direito, arraste para rolar, menu de controle na tela com botões sensíveis ao contexto

### Android Nativo (Experimental)

> [!CAUTION]
> A versão Android é experimental e pode ter funcionalidade limitada comparada à versão de terminal.

O porte Android compila o núcleo do jogo em Go em uma biblioteca nativa `.aar` via [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile), renderizada como uma grade de caracteres em uma `SurfaceView` com entrada por toque completa e áudio.

**Pré-requisitos:**

- Go 1.25 ou superior
- Android SDK + NDK (API 21+)
- Gradle 8.2 (para builds APK locais)
- `gomobile`:
  ```bash
  go install golang.org/x/mobile/cmd/gomobile@latest
  gomobile init
  ```

**Construa a biblioteca do jogo:**

```bash
make android-aar
```

Isso escreve `android/app/libs/termcom.aar`.

**Construir um APK (CI / GitHub Actions):**

Um fluxo de trabalho `.github/workflows/android-release.yml` constrói automaticamente um APK assinado (ou APK de depuração) em push para `mobile`/`main`/`master` e em tags `v*`. APKs de depuração são produzidos a partir de qualquer push; marque uma versão (`v*`) para publicar um APK assinado como um GitHub Release. Defina estes segredos do repositório para lançamentos assinados: `ANDROID_KEYSTORE_BASE64`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS`, `ANDROID_KEY_PASSWORD`.

**Construir um APK localmente:**

```bash
make android-aar                                  # passo 1: Go .aar
cd android && gradle assembleDebug               # passo 2: APK → app/build/outputs/apk/debug/
# ou abra android/ no Android Studio e Execute
```

Instale com `adb install android/app/build/outputs/apk/debug/app-debug.apk`.

**Controles:**

- Toque para clicar / selecionar / mover
- Pressione longamente (500ms) para clicar com botão direito / cancelar; vibração ao pressionar longamente
- Arraste para rolar
- Teclado físico suportado (DPAD, Enter, Escape, teclas F)

## Estrutura do Projeto

Veja o [arquivo AGENTS](AGENTS.md) para detalhes de arquitetura.

## Licença

MIT, veja o arquivo [LICENSE](LICENSE).

> [!NOTE]
> ***Aviso de Uso de IA***: IA foi usada neste projeto para gerar e atualizar as traduções para francês, espanhol, russo, coreano, chinês e japonês. Nenhum áudio ou imagem foi gerado via IA (o jogo não usa nenhum de qualquer forma — é totalmente baseado em terminal!).
