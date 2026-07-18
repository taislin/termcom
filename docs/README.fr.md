# TERMCOM : un remake de X-COM en ASCII pur

[English](README.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt.md) | [Русский](README.ru.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md)

Un remake de style *roguelike* de **X-COM : UFO Defense** *(1994, MicroProse)*, entièrement rendu en ASCII coloré dans un terminal. Écrit en Go avec [tcell](https://github.com/gdamore/tcell). Il ramène l'expérience classique de stratégie d'invasion alien dans ton terminal. Il propose une implémentation complète de toutes les boucles de jeu : le Geoscape (stratégie globale), la gestion de base et le Battlescape tactique.

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/) [![GitHub License](https://img.shields.io/github/license/taislin/termcom)](https://github.com/taislin/termcom/blob/master/LICENSE) [![Release](https://img.shields.io/github/v/release/taislin/termcom)](https://github.com/taislin/termcom/releases/latest)

[![Download Here](https://img.shields.io/badge/Download%20Here-8A2BE2)](https://github.com/taislin/termcom/releases/latest) [![Website](https://img.shields.io/badge/Website-8A2BE2)](https://taislin.github.io/termcom) [![Manual](https://img.shields.io/badge/Manual-8A2BE2)](https://taislin.github.io/termcom/manual.html) [![Dev Guide](https://img.shields.io/badge/Dev%20Guide-8A2BE2)](docs/dev.md)

## Fonctionnalités

- **Geoscape** — Carte mondiale en temps réel avec compression du temps, suivi des OVNI et lancement d'intercepteurs.
- **Battlescape** — Combat tactique au tour par tour avec Unités de Temps (TU), couverture et ligne de vue.
- **Gestion de base** — Construis des installations, recrute des soldats, équipe ton escouade.
- **Recherche et fabrication** — Débloque la technologie alien, fabrique des fusils à plasma et des combinaisons énergétiques.
- **IA alien** — Comportements de patrouille, recherche, attaque, fuite, flanc et retraite.
- **Aliens générés procéduralement** — Un roster d'aliens est généré au débute de chaque campagne, chacun avec des capacités, forces, faiblesses et armes uniques.
- **Terrain destructible** — Les grenades détruisent les murs, les arbres et les rochers, laissant des décombes.
- **VFX dynamique** — Explosions de particules, tremblement d'écran, physique des décombes, éclairage nocturne.

## Prérequis

> [!TIP]
> Ce qui suit concerne la compilation à partir du code source. Si tu veux simplement jouer, tu peux télécharger les binaires du jeu depuis [ici](https://github.com/taislin/termcom/releases/latest).

- Go 1.25 ou plus
- Un terminal avec support Unicode (pour les caractères de dessin de boîte)

### Dépannage des polices du terminal

**termcom** utilise abondamment des caractères Unicode étendus (Runes, Géométriques et symboles Éthiopiens) pour rendre les aliems et les cartes tactiques. La plupart des appareils devraient supporter les caractères que nous utilisons, mais si tu lances le jeu nativement dans ton terminal et vois des caractères superposés, un espacement bizarre ou des cases vides (□) au lieu d'aliens, ton système d'exploitation manque les polices de repli nécessaires.

#### Sous Linux

Pour corriger cela sous **Ubuntu/Debian**, installe les polices Noto et le repli Unifont :

```bash
sudo apt update
sudo apt install fonts-noto fonts-unifont
```

Pour les utilisateurs **Arch Linux** :

```bash
sudo pacman -S noto-fonts unifont
```

Pour les utilisateurs **Fedora** :

```bash
sudo dnf install google-noto-sans-fonts unifont-fonts
```

#### Sous macOS

Le `Terminal.app` par défaut sur macOS peut parfois avoir des difficultés avec l'alignement de la grille ou renderer les symboles incorrectement comme des emojis à double largeur.

Pour la meilleure expérience, nous recommandons vivement d'utiliser **[iTerm2](https://iterm2.com/)**. Si tu manques de caractères, tu peux installer GNU Unifont via Homebrew :

```bash
brew install --cask font-gnu-unifont
```

* **Pour corriger l'alignement dans iTerm2 :** Va dans **Settings > Profiles > Text**, coche la case *"Use a different font for non-ASCII text"*, et définit cette police secondaire sur `Unifont`.

#### Sous Windows

N'utilise pas l'invite de commandes hérité (`cmd.exe`) ni la vieille fenêtre PowerShell bleue, car leur support d'Unicode et des couleurs est extrêmement limité.

1. Utilise **[Windows Terminal](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701)** (inclus par défaut dans Windows 11, disponible dans le Microsoft Store pour Windows 10).
2. Si tu vois des cases vides `□`, la police par défaut de ton système n'a pas les symboles requis.
3. Télécharge et installe une police robuste et très compatible comme **[GNU Unifont](http://unifoundry.com/unifont/)** ou **[Noto Sans Mono](https://fonts.google.com/noto/specimen/Noto+Sans+Mono)**.
4. Ouvre les paramètres de Windows Terminal (`Ctrl + ,`), va dans **Profiles > Defaults > Appearance**, et change **Font face** pour la police que tu viens d'installer.

## Compilation et exécution

### Version terminal

```bash
go run ./cmd/termcom
```

Ou compile un binaire :

```bash
go build -o termcom ./cmd/termcom
./termcom
```

### Version navigateur (Expérimentale)

> [!CAUTION]
> La version navigateur est expérimentale et peut avoir des fonctionnalités limitées par rapport à la version terminal.

La version navigateur te permet de jouer à termcom dans un navigateur web en utilisant xterm.js.

1. Démarre le serveur web :

```bash
go run ./cmd/webserver
```

2. Ouvre ton navigateur et va à :

```
http://localhost:8080
```

La version navigateur prend en charge :

- Entrée clavier complète via xterm.js
- Communication temps réel basée sur WebSocket
- Redimensionnement de terminal réactif
- Toutes les fonctionnalités du jeu (Geoscape, Battlescape, Gestion de base)
- **Jeu tactile mobile** — tape pour cliquer, appuie longuement pour le clic droit, glisse pour défiler, menu de contrôle à l'écran avec des boutons contextuels

### Android natif (Expérimental)

> [!CAUTION]
> La version Android est expérimentale et peut avoir des fonctionnalités limitées par rapport à la version terminal.

Le portage Android compile le cœur de jeu Go en une bibliothèque native `.aar` via [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile), rendue sous forme de grille de caractères sur une `SurfaceView` avec entrée tactile complète et audio.

**Prérequis :**

- Go 1.25 ou plus
- Android SDK + NDK (API 21+)
- Gradle 8.2 (pour les builds APK locaux)
- `gomobile` :
  ```bash
  go install golang.org/x/mobile/cmd/gomobile@latest
  gomobile init
  ```

**Compile la bibliothèque de jeu :**

```bash
make android-aar
```

Cela écrit `android/app/libs/termcom.aar`.

**Compile une APK (CI / GitHub Actions) :**

Un flux de travail `.github/workflows/android-release.yml` compile automatiquement une APK signée (ou APK de débogage) lors d'un push vers `mobile`/`main`/`master` et sur les tags `v*`. Les APK de débogage sont produites à chaque push ; taggue une version (`v*`) pour publier une APK signée en tant que GitHub Release. Définis ces secrets de dépôt pour les versions signées : `ANDROID_KEYSTORE_BASE64`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS`, `ANDROID_KEY_PASSWORD`.

**Compile une APK localement :**

```bash
make android-aar                                  # étape 1 : Go .aar
cd android && gradle assembleDebug               # étape 2 : APK → app/build/outputs/apk/debug/
# ou ouvre android/ dans Android Studio et Execute
```

Installe avec `adb install android/app/build/outputs/apk/debug/app-debug.apk`.

**Contrôles :**

- Tape pour cliquer / sélectionner / déplacer
- Appuie longuement (500 ms) pour clic droit / annuler ; vibration lors de l'appui long
- Glisse pour défiler
- Clavier matériel pris en charge (DPAD, Enter, Escape, touches F)

## Structure du projet

Vois le [fichier AGENTS](AGENTS.md) pour les détails d'architecture.

## Licence

MIT, vois le fichier [LICENSE](LICENSE).

> [!NOTE]
> ***Avertissement d'utilisation de l'IA*** : L'IA a été utilisée dans ce projet pour générer et mettre à jour les traductions en français, espagnol, russe, coréen, chinois et japonais. Aucun audio ni image n'a été généré par l'IA (le jeu n'en utilise de toute façon aucun — il est entièrement basé sur le terminal !).
