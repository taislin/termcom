# Démarrage Rapide termcom

Un démake ASCII de X-COM pour votre terminal. Commandez la défense de l'humanité
contre l'invasion alien.

## Lancer

### Version Terminal

```bash
go run ./cmd/termcom      # ou : make run
```

### Version Navigateur (WASM)

Pas de serveur backend nécessaire — s'exécute directement dans le navigateur.

```bash
# Compiler et servir
cd cmd/termcom_wasm
GOOS=js GOARCH=wasm go build -o ../../web_wasm/termcom.wasm .
cd ../../web_wasm
python -m http.server 8080
# Ouvrir http://localhost:8080
```

Ou utilisez le script de compilation : `./scripts/build_wasm.sh`

## Boucle de Jeu

1. **Géoscape** -- Les OVNIs volent vers les villes. Détectez-les et interceptez-les.
2. **Intercepter** -- Lancez des chasseurs (L) ou résolvez auto (A) pour abattre les OVNIs.
3. **Bataille** -- Déployez sur les sites de crash (R). Entrez dans le combat tactique.
4. **Base** -- Recherchez la techno alien, fabriquez l'équipement, engagez/équipez des soldats.
5. **Répéter** -- Gagnez 10 batailles, puis assaillez Cydonia pour sauver la Terre.

Perdez si l'Activité Extraterrestre atteint 100 %.

## Touches Essentielles (Géoscape)

| Touche | Action |
|--------|--------|
| Space | Pause |
| 1-4 | Vitesse du temps |
| L | Lancer l'intercepteur |
| A | Résolution auto OVNI |
| M | Répondre à la mission |
| R | Envoyer le transport au crash |
| B | Ouvrir la base |
| F5/F9 | Sauver/Charger |
| Q | Quitter |

## Touches Essentielles (Champ de Bataille)

| Touche | Action |
|--------|--------|
| Flèche/ZQSD | Déplacer le curseur |
| Space/Entrée | Sélectionner/Confirmer |
| F | Tirer |
| R | Recharger |
| Q | Cycler soldat |
| E | Fin du tour |
| C | S'accroupir |
| Esc | Annuler |

## Stratégie Rapide

- **Début :** Engagez des soldats, recherchez les Alliages Extraterrestres, construisez Labo + Atelier
- **Milieu :** Armes laser sur mesure (Concepteur d'Armes) → Armure Personnelle, étendez les bases
- **Fin :** Armes à plasma sur mesure, Tenues de Puissance/Vol, entraînement psi
- Équipez toujours les soldats avant la bataille. Les blessés guérissent 2 PV/jour.
- Concevez des intercepteurs sur mesure dans le Concepteur d'Avions — missiles Stingray + blindage Alliage Léger est un bon départ.
- Vendez l'excès d'artefacts alien pour du cash. Les installations Radar boostent le financement.

## Victoire

Gagnez 10 batailles terrestres pour débloquer la mission finale de Cydonia. Détruisez
Cydonia pour gagner.

Pour le manuel complet voir [manual.fr.md](manual.fr.md).
