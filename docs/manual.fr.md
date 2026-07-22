# termcom — Manuel du Démake Roguelike ASCII de X-COM (v0.51.18)

## Table des Matières

1. [Aperçu](#aperçu)
2. [Premiers Pas](#premiers-pas)
3. [Tutoriel / Onboarding](#tutoriel--onboarding)
4. [Géoscape](#géoscape)
5. [Gestion de la Base](#gestion-de-la-base)
6. [Recherche & Fabrication](#recherche--fabrication)
7. [Équiper les Soldats](#équiper-les-soldats)
8. [Champ de Bataille](#champ-de-bataille)
9. [Armes & Équipement](#armes--équipement)
10. [Armure](#armure)
11. [Extraterrestres](#extraterrestres)
12. [Rangs & Progression des Soldats](#rangs--progression-des-soldats)
13. [Sauvegarder/Charger](#sauvegardercharger)
14. [Référence des Touches](#référence-des-touches)
15. [Conseils & Stratégie](#conseils--stratégie)

---

## Aperçu

**termcom** est un démake ASCII roguelike de X-COM: UFO Defense (1994), rendu
entièrement dans un terminal. Vous commandez X-COM — une force d'intervention
internationale chargée de défendre la Terre contre une invasion extraterrestre.

**Votre objectif :** Rechercher la technologie alien, fabriquer des armes et des
armures, et mener des escouades au combat tactique pour éliminer la menace alien.

**Victoire :** Remportez assez de batailles pour déclencher la mission finale de
Cydonia, puis gagnez-la.

**Défaite :** L'Activité Extraterrestre atteint 100 % — l'invasion submerge la Terre.

**Difficulté :** Choisissez un niveau avant de commencer. Les difficultés plus
élevées rendent les aliens plus coriaces, les OVNIs plus fréquents, et les fonds
de départ plus serrés.

- **Débutant** — Aliens plus faibles, OVNIs plus lents, plus de fonds de départ
- **Expérimenté** — Standard
- **Vétéran** — Aliens plus coriaces, OVNIs plus rapides, moins de cash
- **Génie** — Beaucoup plus dur dans tous les domaines
- **Surhumain** — Menace alien maximale

**Langue :** 8 langues disponibles — changez dans l'écran des Options.
Anglais, Chinois, Espagnol, Français, Russe, Portugais, Japonais, Coréen.

**Options :** Appuyez sur `?` sur n'importe quel écran pour ouvrir l'aide, ou
accédez à l'écran des Options pour ajuster la luminescence, l'éclairage, le son,
la sauvegarde automatique, la secousse d'écran, le support souris, les lignes de
grille, les boîtes de confirmation, le thème, la vitesse de résolution, le volume
et la langue.

---

## Tutoriel / Onboarding

Lors de votre première partie (aucun fichier de sauvegarde détecté), un
**Briefing du Commandant** pas à pas apparaît automatiquement après que vous avez
choisi votre difficulté. Il couvre :

1. **Bienvenue** — Introduction à X-COM
2. **Géoscape & Temps** — Pause (`Space`), vitesse (`1`–`4`)
3. **Détection OVNI** — Radar et marqueurs OVNI
4. **Lancement d'Intercepteur** — Appuyez sur `L` pour engager
5. **Réponse à Mission** — Appuyez sur `M` pour déployer
6. **Gestion de Base** — Appuyez sur `B` pour gérer votre base
7. **Champ de Bataille** — Unités Temps, mouvement et combat
8. **Terminé** — Vous êtes prêt

**Contrôles :** Entrée avance, S passe, Esc rejette.

**Relecture :** Ouvrez l'écran des Options et sélectionnez « Revoir le Tutoriel »
à tout moment.

---

## Premiers Pas

Vous commencez sur le **Géoscape** — la carte mondiale. Le temps avance
automatiquement. Les OVNIs apparaissent sur le radar à leur entrée dans la portée.

**Ressources de départ :**
- 500 000 $ (modifié selon la difficulté)
- 10 scientifiques, 10 ingénieurs
- Une base avec Quartiers de Vie, Laboratoire, Atelier, Entrepôt et Radar
- Plusieurs fusils et pistolets

**Lorsqu'un OVNI est détecté :**

| Action | Touche | Ce qui se passe |
|--------|--------|-----------------|
| Lancer un intercepteur | `L` | Envoyer un chasseur pour l'abattre |
| Déployer un transport | `R` | Envoyer des troupes pour investiguer un site de crash |
| Résolution auto | `A` | Résultat automatique rapide d'interception |
| Répondre à la mission | `M` | Déployer sur les missions terror/ravitaillement alien |

---

## Géoscape

Le Géoscape affiche un **tableau de bord régional** avec les niveaux de menace
pour chaque région :

- **Panneau gauche :** liste des régions avec barres de menace et statut radar
- **Panneau droit :** minimap ASCII montrant les bases, OVNIs, intercepteurs et routes

### Symboles de la Minimap

| Symbole | Signification |
|---------|---------------|
| ◆ | Votre base |
| ◉ | Nœud actuellement sélectionné |
| ○ | Hub régional (vert=sûr, jaune=menace, rouge=danger) |
| · | Anneau de couverture radar |
| ! | OVNI (rouge, gras) |
| > | Intercepteur en patrouille |
| ► | Intercepteur engagé sur un OVNI |
| ✕ | Intercepteur ou OVNI détruit |
| * | Site de crash (jaune=non pillé, gris=pillé) |
| ≈ | Transport en route (vert) |

### Contrôles du Temps

| Touche | Vitesse | Usage |
|--------|---------|-------|
| Space | Pause | Stopper le temps pour planifier |
| 1 | 1x | Avance lente |
| 2 | 5x | Vitesse de patrouille normale |
| 3 | 20x | Avance rapide |
| 4 | 60x | Vitesse maximale |

### Budget Mensuel

- **Revenus :** 200 000 $ de base + 50 000 $ par installation Radar
- **Dépenses :** 2 000 $ par soldat, scientifique et ingénieur

### Bases Multiples

Appuyez sur `N` sur un nœud vide pour construire une nouvelle base (500 K$). Chaque
base a ses propres installations, soldats et entrepôts. Appuyez sur `C` pour
cycler la base active. Appuyez sur `T` pour ouvrir l'écran de Transfert et déplacer
des soldats ou objets entre les bases.

### Réponse à Mission

Lorsqu'une mission apparaît (terreur, raid d'approvisionnement, enlèvement, etc.),
appuyez sur `M` pour répondre :

| Option | Résultat |
|--------|----------|
| **Déployer l'escouade** | Combat tactique complet — meilleures récompenses, risque maximal |
| **Résolution auto** | Issue rapide — XP réduit, aucun cadavre, mais sûr |
| **Ignorer** | Passer — l'activité alien augmente |

La résolution auto donne environ la moitié de l'XP d'un vrai combat, aucun cadavre
alien, et une petite chance de pertes en cas de défaite.

### Défense de Base

Si une mission vise un nœud avec votre base, y répondre lance une bataille de
**Défense de Base**. Perdre une défense de base détruit la base et son personnel.
Perdre votre dernière base met fin à la partie.

### Interception OVNI

Appuyez sur `L` pour lancer un intercepteur sur l'OVNI le plus proche. L'intercepteur
poursuit et engage dans un court combat aérien résolu automatiquement. La minimap
montre l'engagement avec barres de PV et retour touché/raté.

### Missions Alien

Les missions apparaissent toutes les ~30 minutes de jeu avec un minuteur de 12 à
36 heures :

| Mission | Minuteur | À quoi s'attendre |
|---------|----------|-------------------|
| Terreur | 24h | Carte urbaine, nombreux civils en danger |
| Raid d'Approvisionnement | 24h | Intérieur d'OVNI, bonus alliages/élérium |
| Enlèvement | 24h | Carte rurale, secourir des civils |
| Recherche Alien | 24h | Intérieur d'OVNI, bonus technologie alien |
| Conseil | 36h | Carte urbaine, bonus de financement +100 K$ |
| Assaut de Base Alien | 12h | Base alien rocheuse, butin technologique majeur |

Laisser expirer une mission augmente l'Activité Extraterrestre de 10 %.

---

## Gestion de la Base

Appuyez sur `B` depuis le Géoscape pour ouvrir votre base.

### Onglets

| Touche | Onglet |
|--------|--------|
| 1 | Installations |
| 2 | Soldats |
| 3 | Recherche |
| 4 | Fabrication |
| 5 | Transfert |
| 6 | Hangars |

### Installations

| Installation | Coût | Temps de Construction | Effet |
|--------------|------|-----------------------|-------|
| Quartiers de Vie | 50 K$ | 5 jours | +8 capacité de soldats |
| Laboratoire | 75 K$ | 7 jours | Active la recherche |
| Atelier | 60 K$ | 7 jours | Active la fabrication |
| Entrepôt | 40 K$ | 3 jours | +50 stockage d'objets |
| Radar | 80 K$ | 5 jours | +50 K$ de financement mensuel |
| Conteneur à Extraterrestres | 100 K$ | 10 jours | Contient jusqu'à 10 aliens vivants |
| Labo Psi | 150 K$ | 14 jours | Entraîne la capacité psi |
| Hangar | 120 K$ | 8 jours | Abrite un intercepteur |

**Bonus d'adjacence :** Placer des installations de même type côte à côte aide :
- Laboratoires adjacents : recherche plus rapide (jusqu'à +30 %)
- Ateliers adjacents : fabrication plus rapide (jusqu'à +30 %)
- Quartiers de Vie adjacents : guérison des soldats plus rapide (jusqu'à +3 PV/jour)

**Contrôles :** `B` pour construire, `S` pour vendre (remboursement 50 %).

### Onglet Soldats

Engagez des soldats à 50 K$ chacun.

| Touche | Action |
|--------|--------|
| H | Engager un soldat |
| E | Ouvrir l'écran d'équipement |
| G | Ouvrir le concepteur d'armes |
| D | Démétre un soldat |

### Onglet Hangars

Chaque Hangar abrite un intercepteur. Gérez votre force aérienne ici.

| Touche | Action |
|--------|--------|
| B | Acheter un intercepteur |
| W | Équiper l'arme |
| G | Ouvrir le concepteur d'armes |
| D | Ouvrir le concepteur d'avions |

---

## Recherche & Fabrication

### Recherche

Depuis l'onglet Recherche, affectez des scientifiques à des sujets. La recherche
progresse automatiquement au fil du temps de jeu.

L'arbre technologique est **procédural** — chaque partie génère un arbre unique
à partir d'un algorithme à graine. Les technologies de base (Armes Laser, Armure
Personnelle, Armes à Plasma) sont toujours présentes, mais les prérequis et coûts
varient.

**Priorités :**
- **Alliages Extraterrestres** et **Élérium-115** doivent être recherchés en premier
- Les autopsies d'espèces alien débloquent le lore et peuvent bloquer des technos d'armes
- Interrogez les aliens capturés (touche `I`) pour terminer la recherche plus vite

### Fabrication

Depuis l'onglet Fabrication, affectez des ingénieurs pour produire des objets.
Cela produit **des armes et armures standard** — pour des équipements sur mesure
plus puissants, utilisez le Concepteur d'Armes et le Concepteur d'Avions.

**Objets fabriquables** (construire plusieurs à la fois) :

| Objet | Temps | Matériaux |
|-------|-------|-----------|
| Pistolet, Fusil, Canon Lourd, Canon Automatique | 3–8 jours | Alliages |
| Lance-Roquettes, Bâton Électrique | 2–8 jours | Alliages + Élérium |
| Armure Personnelle, Tenue Légère/Intermédiaire/Lourde/De Puissance/De Vol | 6–18 jours | Alliages + Élérium |
| Trousse Médicale | 3 jours | Alliages |

Plus d'ingénieurs = production plus rapide. Les armes à énergie (Laser, Plasma) ne
peuvent pas être fabriquées — elles doivent être recherchées, conçues et construites
via le Concepteur d'Armes, ou récupérées sur les aliens.

---

## Équiper les Soldats

Depuis l'onglet Soldats, appuyez sur `E` pour ouvrir l'écran d'équipement.

### Contrôles

| Touche | Action |
|--------|--------|
| ↑/↓ | Sélectionner un soldat |
| Tab | Cycler les objets disponibles |
| 1 | Emplacement Arme |
| 2 | Emplacement Armure |
| 3 | Emplacement Inventaire |
| Space | Équiper l'objet sélectionné |
| G | Ouvrir le concepteur d'armes |
| A | Auto-équiper tous les soldats |
| Esc | Retour |

### Emplacements

- **Emplacement 1 (Arme) :** Arme principale — conçue sur mesure ou un fusil/pistolet standard
- **Emplacement 2 (Armure) :** Armure corporelle — personnelle, tenue légère, etc.
- **Emplacement 3 (Inventaire) :** Objets supplémentaires — grenades, trousses médicales,
  scanners, mines de proximité, amplificateurs psi, armes de mêlée

### Encombrement

Chaque objet a un poids. Le poids total arme + armure + inventaire est votre
**encombrement**. Un encombrement plus élevé réduit vos Unités Temps au combat
(environ 1 UT de pénalité par 5 unités de poids). Gardez vos soldats légèrement
chargés pour une mobilité maximale.

### Auto-Équipement

Appuyez sur `A` pour équiper automatiquement chaque soldat avec la meilleure arme
et armure disponibles dans les entrepôts. L'équipement existant est retourné aux
entrepôts — un moyen rapide de rééquiper votre escouade après avoir recherché une
nouvelle techno.

---

## Champ de Bataille

Le Champ de Bataille est un combat tactique au tour par tour. Vous contrôlez une
escouade de soldats contre des forces alien sur une carte de grille 50×50.

### Structure du Tour

1. **Tour du Joueur** — Déplacez et agissez avec chaque soldat en utilisant les Unités Temps (UT)
2. **Tour des Extraterrestres** — Les aliens agissent en utilisant leurs propres réserves d'UT
3. Répétez jusqu'à l'élimination d'un camp

### Unités Temps (UT)

Chaque action coûte des UT. Les UT se restaurent entièrement au début de chaque tour du joueur.

| Action | Coût UT approx. |
|--------|-----------------|
| Déplacer (par case) | 4 |
| S'accroupir | 4 |
| Tirer | Varie selon l'arme (visé=base, rafale=1,5×, auto=2×) |
| Recharger | 8 |
| Lancer une grenade | 20 |
| Utiliser le médikit | 25 |
| Attaque psi | 20 |

La réserve d'UT d'un soldat commence à 45–55 et peut augmenter par l'expérience (max ~80).

### Modes de Tir

Les armes peuvent avoir plusieurs modes de tir. Appuyez sur **Tab** pour cycler, et
vérifiez le mode affiché dans la barre latérale.

| Mode | Coût | Précision | Munitions | Quand l'utiliser |
|------|------|-----------|-----------|------------------|
| **Visé** | UT de base | Meilleure | 1 tir | Longue portée, cibles de haute valeur |
| **Rafale** | 1,5× UT | -10 % | 3 tirs | Moyenne portée, suppression |
| **Auto** | 2× UT | -20 % | Toutes les munitions restantes | Courte portée, urgences |

Toutes les armes ne prennent pas en charge tous les modes. Les fusils et fusils
laser prennent en charge la rafale ; seules quelques armes prennent en charge le tir auto.

### Combat

Facteurs de combat :
- **La précision** dépend de la compétence du soldat, distance à la cible, couverture
  et mode de tir utilisé
- **L'accroupi** donne un bonus de précision et réduit les dégâts entrants
- **La couverture** — les murs bloquent 80 % des dégâts, les arbres 60 %, les buissons 40 %,
  les clôtures 30 %. Placez vos soldats derrière une couverture solide
- **Contournez la couverture** avec des grenades — elles explosent en zone et ignorent
  la réduction des dégâts de couverture

### Ligne de Vue

Les soldats ne peuvent voir qu'en ligne droite. Les murs, arbres et roches bloquent
la ligne de vue. Les sols, portes et herbes ne la bloquent pas.

### Objets & Couverture

Les tirs passant à travers des objets ont leurs dégâts réduits par la valeur de
couverture de l'objet. La plus haute couverture le long de la ligne de feu est appliquée.

| Objet | Couverture % |
|-------|--------------|
| Mur / Mur d'OVNI | 80 % |
| Roche | 70 % |
| Arbre | 60 % |
| Mobilier d'OVNI | 50 % |
| Buisson | 40 % |
| Fumée Épaisse | 40 % |
| Clôture | 30 % |
| Décombres | 20 % |

### Grenades

- Portée : ~6 cases
- Dégâts : basés sur la force, avec éclaboussure en zone
- Détruit les murs, arbres, roches et clôtures dans le rayon d'explosion
- Crée des nuages de fumée qui bloquent la ligne de vue à haute densité

### Trousse Médicale

Soigne 10 PV par usage (15 PV avec le perk Médecin de Terrain), coûte 25 UT.

### Missions de Nuit

Nuit (avant 6h00 ou après 18h00) :
- Précision réduite (environ 75 % de jour)
- Portée de vue réduite (de 20 à ~10 cases)
- Les soldats brillent chaud, les aliens brillent faiblement bleu

### Modes de Vision

Appuyez sur `V` pour cycler : **Normale → Vision Nocturne → Vision Thermique → Normale**
- Vision Nocturne : superposition phosphorique verte avec statique
- Vision Thermique : les entités vivantes brillent chaud, le terrain est bleu froid

### Combat Psi

Nécessite une arme Amplificateur Psi et une installation Labo Psi. Le succès dépend
de la compétence psi de votre soldat contre la force psi de la cible. Une attaque psi
réussie panique la cible — elle perd son tour.

Les soldats dans une base avec un Labo Psi peuvent gagner de la compétence psi avec le
temps (jusqu'à ~80). La recherche Contrôle Mental accorde un boost psi significatif à
tous les soldats.

### Modificateurs de Mission

Modificateurs aléatoires qui changent à chaque bataille :

| Modificateur | Ce qui se passe |
|--------------|-----------------|
| Op. Nocturnes | Bataille de nuit forcée, butin bonus |
| Renforts | Des aliens supplémentaires arrivent au tour 4 |
| Limite de Temps | 15 tours pour éliminer tous les aliens |
| Sauvetage VIP | Protéger un VIP, cash bonus s'il survit |
| Piégé | Plus de grenades et de mines sur la carte |
| Brouillard Épais | Portée de vue réduite de 40 % |
| Embuscade Alien | Les aliens démarrent en positions de surveillance |
| Faible Visibilité | Précision réduite pour toutes les unités |
| Hauteurs | Les positions surélevées donnent un bonus de précision |

### Météo

La météo affecte le combat selon l'emplacement de la mission :

| Météo | Effet |
|-------|-------|
| Pluie | Précision réduite, feu se propage plus lentement |
| Vent | Feu se propage plus vite, grenades peuvent dériver |
| Neige | Déplacement coûte plus cher dans la neige profonde |
| Brouillard | Précision réduite, vue réduite |
| Tempête | Pluie + vent combinés |
| Froid | Légère pénalité de précision |

### Rapport Après Action

Après chaque bataille, vous voyez :
- Issue (Victoire / Défaite)
- Aliens tués et soldats perdus
- Butin récupéré et prisonniers capturés
- Fonds gagnés
- Gains de stats par soldat ou marqueur « TUE »

Appuyez sur **Entrée**, **Space** ou **Esc** pour rejeter.

---

## Armes & Équipement

### Concepteur d'Armes sur Mesure

Appuyez sur `G` depuis la Base, l'onglet Soldats ou l'écran d'Équipement pour ouvrir
le **Concepteur d'Armes**. C'est le moyen principal de créer des armes pour vos
soldats. Choisissez un modèle de base et personnalisez chaque composant :

| Composant | Options | Ce qu'il affecte |
|-----------|---------|------------------|
| **Base** | Pistolet / Fusil | Dégâts de départ, portée, précision, coût UT |
| **Canon** | Court / Standard / Long / Allongé | Portée, précision, coût UT, poids |
| **Optique** | Aucun / Viseur Fer / Lunette / Optique Avancée | Précision, coût UT, poids |
| **Mode de Tir** | Semi / Full-Auto | Mode full-auto (tire plus vite, moins précis) |
| **Munitions** | Standard / Perforant / Incendiaire / Explosif | Mod dégâts, coût UT, poids |
| **Crosse** | Aucune / Légère / Lourde | Précision, coût UT, poids |

Chaque composant affecte les dégâts, la précision, le coût UT, la portée et le poids
de l'arme. Le panneau d'aperçu montre l'arme assemblée en ASCII coloré et affiche ses
stats finales. Les conceptions sont sauvegardées comme objets personnalisés disponibles
dans l'écran d'Équipement.

**Astuce :** Commencez avec une base Fusil pour la plupart des usages. Les canons longs
et les lunettes améliorent la précision à distance. Les munitions explosives frappent
fort mais coûtent des UT supplémentaires.

### Armes Standard

Ces objets de base sont disponibles dès le début et peuvent être fabriqués :

| Type | Dégâts | Munitions | Notes |
|------|--------|-----------|-------|
| Pistolet | Léger | Balistique | Nécessite rechargement, faible poids |
| Fusil | Moyen | Balistique | Équipement standard, supporte la rafale |
| Canon Lourd | Élevé | Balistique | Lent, lourd, frappe fort |
| Canon Automatique | Moyen | Balistique | Option full-auto |
| Lance-Roquettes | Très Élevé | Explosif | Dégâts en zone |

### Armes à Énergie

Recherchées plus tard — ne nécessitent jamais de rechargement :

| Type | Dégâts | Notes |
|------|--------|-------|
| Pistolet Laser | Léger | Arme à énergie précoce |
| Fusil Laser | Moyen | Supporte le tir en rafale, ne recharge jamais |
| Pistolet à Plasma | Moyen | Arme alien, ne recharge jamais |
| Fusil à Plasma | Élevé | Arme alien, ne recharge jamais |
| Plasma Lourd | Très Élevé | Arme alien de haut niveau |

### Munitions & Rechargement

- **Armes balistiques** nécessitent un rechargement — appuyez sur `R` au combat
- **Armes à énergie** (Laser, Plasma) ne nécessitent jamais de rechargement
- **Consommables** (grenades, trousses médicales) sont utilisés depuis votre inventaire

### Modes de Tir

Voir [Champ de Bataille → Modes de Tir](#modes-de-tir) pour les détails.

### Objets d'Inventaire

Les soldats peuvent transporter des objets supplémentaires dans leur emplacement inventaire :
- **Grenades** — explosif lancé avec dégâts en zone
- **Trousses Médicales** — soignez-vous ou un allié adjacent
- **Scanners de Mouvement** — détectent les ennemis proches
- **Mines de Proximité** — posées au sol, détonnent quand un ennemi passe dessus
- **Amplificateurs Psi** — activent les attaques psi (nécessite compétence psi)
- **Armes de mêlée** — Bâton Électrique pour les neutralisations non létales

Chaque objet d'inventaire ajoute du poids et augmente l'encombrement, réduisant vos
UT disponibles au combat. Chargez judicieusement.

### Objets Procéduraux

Chaque partie génère des armes et armures uniques selon l'espèce alien rencontrée.
Elles ont des noms et stats aléatoires — chaque jeu est différent.

**Armes procédurales :** 2–3 armes avec types de dégâts correspondant à l'espèce alien.
**Armures procédurales :** 1–2 pièces d'armure avec protection correspondant aux types
de dégâts alien.

Ces objets sont automatiquement ajoutés à vos entrepôts au démarrage du jeu.

---

## Armure

| Armure | Défense | Pénalité UT | Notes |
|--------|---------|-------------|-------|
| Aucune | 0 | Aucune | Par défaut |
| Armure Personnelle | 10 | Aucune | Standard en début de partie |
| Tenue Légère | 20 | -5 % UT | Bonne option milieu de partie |
| Tenue Intermédiaire | 30 | -10 % UT | Protection solide |
| Tenue Lourde | 40 | -15 % UT | Défense max, pénalité lourde |
| Tenue de Puissance | 50 | -10 % UT | Armure de fin de partie |
| Tenue de Vol | 45 | -5 % UT | Presque fin de partie, plus légère que Puissance |

Une défense plus élevée réduit les dégâts subis, mais les tenues plus lourdes coûtent
des Unités Temps.

---

## Extraterrestres

### Espèces Procédurales

Chaque jeu génère 5–7 espèces alien uniques à partir d'une graine. Chaque espèce a
2–5 variantes de rang (Soldat → Navigateur → Commandant → Élite → Suprême).

Les espèces diffèrent par :
- **Type de dégâts** — la sorte de dégâts qu'elles infligent
- **Résistances & faiblesses** — certaines sont faibles au plasma, d'autres aux explosifs
- **Préférence d'arme** — les rangs bas utilisent des pistolets, les rangs hauts des armes lourdes
- **Morphologie** — plan corporel physique affectant stats et résistances

Cela signifie que **chaque partie présente des menaces alien différentes**. Une partie
peut avoir une espèce à dominante psionique, une autre des prédateurs de mêlée faibles
aux explosifs.

### Morphologie

La morphologie détermine la forme physique d'un alien. Facteurs clés :

**Membres :**
- Bras (0–6) : Moins de bras = moins de précision, plus de bras = meilleure stabilité ou double emploi
- Jambes (0–8) : Plus de jambes = plus rapide mais cible plus grande ; zéro jambe = flottant, plus dur à toucher

**Types de corps et leurs résistances :**
- **Chair de Carbone :** +Résistance Cinétique, -Faiblesse Explosif
- **À Base de Silicium :** +Laser/+Plasma, -Faiblesse Explosif, réfléchissant
- **Gazeux :** Immunisé au cinétique, faible au plasma, peut traverser les murs
- **Cristallin :** Bonne résistance générale, très faible aux explosifs, se brise à la mort
- **Amorphe :** +Résistance Psi, régénère les PV chaque tour
- **Mécanique :** Immunisé au psi, +Résistance Plasma, -Faiblesse Laser, auto-destruction
- **Bio-Synthétique :** Résistances équilibrées, soigne les aliens adjacents
- **Nanotechnologique :** +Résistance Cinétique, peut ressusciter à la mort

**Sens :**
- **Vue :** Affecte la précision — le multi-spectre ignore fumée/obscurité
- **Ouïe :** L'écholocation détecte les unités à travers la fumée de près
- **Sens Thermique :** Détecte les unités vivantes sans égard à la couverture de près
- **Sens Psionique :** Boost psi, détecte les humains contrôlés mentalement
- **Sens Chimique :** Bonus de précision contre les cibles blessées

### Niveaux de Connaissance

Au fur et à mesure que vous rencontrez des aliens, le renseignement s'améliore :

| Niveau | Ce que vous apprenez |
|--------|----------------------|
| Inconnu | Le nom apparaît comme « ??? » |
| Aperçu | Nom et icône révélés |
| Tué | Stats et résistances révélées |
| Autopsié | Lore complet et faiblesses détaillées |

### IA Alien

Les aliens patrouillent jusqu'à repérer un humain, puis attaquent. Comportements :
- **Recherche** — se déplacer vers la dernière position connue pendant quelques tours
- **Fuite** — s'enfuir quand gravement blessé et faible en bravoure
- **Adaptation** — les aliens étudient vos tactiques à travers les missions.
  Tirez de loin ? Ils vous chargeront. Utilisez des grenades ? Ils s'étaleront.
  Flanquez souvent ? Ils posteront des suppresseurs.

### Escalade d'Équipement

Les aliens obtiennent un meilleur équipement à mesure que la campagne progresse :
- **Premiers mois :** Pistolets à plasma, armure basique
- **Milieu de campagne :** Fusils à plasma, plasma lourd, canons alien
- **Fin de campagne :** Armes et armures alien de haut niveau

### Capture Alien

Utilisez un **Bâton Électrique** (mêlée, 2 K$ à fabriquer) pour assommer les aliens.
Si les dégâts d'assommement dépassent leurs PV, ils tombent inconscients et peuvent
être récupérés après la mission — à condition d'avoir un Conteneur à Extraterrestres
avec capacité libre.

Les aliens capturés peuvent être interrogés depuis l'écran Recherche (touche `I`) :
- L'interrogatoire peut terminer une autopsie active instantanément
- Ou accorder un bonus de progression à la recherche en cours
- Nécessite au moins un Laboratoire

---

## Rangs & Progression des Soldats

### Rangs

Les rangs se débloquent à mesure que votre effectif total grandit :

| Rang | Débloqué quand l'effectif atteint |
|------|-----------------------------------|
| Recrue | Toujours disponible |
| Caporal | Toujours disponible |
| Caporal-Chef | 4 soldats |
| Sergent | 8 soldats |
| Lieutenant | 14 soldats |
| Capitaine | 22 soldats |
| Commandant | 30 soldats |
| Colonel | 40 soldats |

### Croissance des Stats

Les soldats s'améliorent par **expérience par action** pendant la bataille :
- **Tir** → améliore la Précision
- **Réactions** → améliore les Réactions
- **Mêlée** → améliore la Force
- **Bravoure** → améliore la Bravoure (en résistant à la panique)
- **Compétence psi** → améliore la Compétence Psi et la Force Psi

Après chaque mission, l'XP accumulée est convertie en gains de stats. Les soldats
ayant gagné de l'XP obtiennent aussi une croissance générale de « halo » vers leurs
plafonds de PV, UT et Force. Les plafonds sont environ : UT 80, PV 60, Précision 120,
Réactions 100, Bravoure 100, Force 70, Psi 100.

### Fatigue & Blessures

- **Soldats blessés** ne peuvent déployer jusqu'à guérison (2 PV/jour de récupération)
- **Fatigue :** Les batailles causent 1–5 jours de fatigue
- Les installations de soin et les Quartiers de Vie accélèrent la récupération

### Blessures Fatales

Au combat, les coups peuvent causer des blessures fatales et des saignements. Le
saignement draine les PV à chaque tour — appliquez-y un médikit vite. Les blessures
survivantes deviennent des jours de récupération après la mission.

### Moral

Les soldats récupèrent du moral à chaque tour. Un moral bas peut déclencher la panique
(sauter le tour). Résister à la panique construit l'XP de bravoure.

### Perks

Chaque montée de rang accorde un perk aléatoire :

| Perk | Effet |
|------|-------|
| Réflexes Éclair | +10 Réactions |
| Tireur d'Élite | +Précision à longue portée |
| Grenadier | Plus grande éclaboussure de grenade |
| Médecin de Terrain | Médikit soigne plus |
| Volonté de Fer | +Compétence Psi et +Force Psi |
| Visée Stable | +Précision à l'arrêt |
| Spécialiste de Combat Rapproché | +Précision à courte portée |
| Spécialiste de Surveillance | +Précision de tir de réaction |
| Démolitions | +Dégâts de grenade |
| Écorcheur | +Butin des batailles |
| Robuste | +5 PV Max |
| Apprenant Rapide | +Gain XP |

### Mémorial

Les soldats tués au combat sont inscrits dans le Mémorial en jeu.
Vous pouvez le consulter pour honorer les tombés.

---

## Sauvegarder/Charger

| Touche | Action |
|--------|--------|
| F5 | Ouvrir le sélecteur d'emplacement de sauvegarde |
| F9 | Ouvrir le sélecteur d'emplacement de chargement |

Les sauvegardes incluent : temps de jeu, fonds, état de pause, activité alien, état
de la base, OVNIs, missions actives, graine d'espèces procédurales et niveaux de
connaissance alien. La graine assure la régénération des mêmes espèces alien au rechargement.

**Sauvegarde auto :** Si activée dans les Options, le jeu sauvegarde automatiquement
périodiquement.

---

## Référence des Touches

### Géoscape

| Touche | Action |
|--------|--------|
| Touches fléchées | Déplacer la caméra |
| j/k | Naviguer dans la liste des régions |
| Space | Pause/reprise |
| 1–4 | Vitesse du temps |
| B | Ouvrir la base |
| L | Lancer l'intercepteur |
| A | Résolution auto de l'OVNI le plus proche |
| M | Répondre à la mission |
| R | Déployer le transport vers le site de crash |
| C | Cycler vers la base suivante |
| N | Construire une nouvelle base (500 K$) |
| T | Ouvrir l'écran de transfert |
| E | Ouvrir l'encyclopédie |
| V | Basculer la superposition radar |
| F5 | Sauver |
| F9 | Charger |
| Q | Quitter |
| ? | Aide |

### Gestion de la Base

| Touche | Action |
|--------|--------|
| 1–6 | Changer d'onglet |
| j/k | Naviguer dans les objets |
| B | Construire une installation |
| S | Vendre une installation |
| H | Engager un soldat |
| E | Ouvrir l'écran d'équipement |
| G | Ouvrir le concepteur d'armes |
| D | Démétre soldat / Concepteur d'Avions (Hangars) |
| Esc | Retour au géoscape |

### Écran d'Équipement

| Touche | Action |
|--------|--------|
| ↑/↓ | Sélectionner un soldat |
| Tab | Cycler les objets disponibles |
| 1 | Emplacement Arme |
| 2 | Emplacement Armure |
| 3 | Emplacement Inventaire |
| Space | Équiper l'objet sélectionné |
| G | Ouvrir le concepteur d'armes |
| A | Auto-équiper tous les soldats |
| Esc | Retour |

### Concepteur d'Armes

Appuyez sur `G` depuis la Base, l'onglet Soldats ou l'écran d'Équipement.

| Paramètre | Options | Effet |
|-----------|---------|-------|
| Canon | Court / Standard / Long / Allongé | Portée, précision, poids, coût UT |
| Optique | Aucun / Viseur Fer / Lunette / Avancée | Précision, poids, coût UT |
| Mode de Tir | Semi / Full-Auto | Mode full-auto |
| Munitions | Standard / Perforant / Incendiaire / Explosif | Dégâts, poids, coût UT |
| Crosse | Aucune / Légère / Lourde | Précision, poids, coût UT |

### Concepteur d'Avions (Intercepteurs sur Mesure)

Tous les intercepteurs sont conçus et construits via le **Concepteur d'Avions**.
Appuyez sur `D` depuis l'onglet Hangars pour l'ouvrir. Configurez votre appareil :

| Paramètre | Plage | Ce qu'il affecte |
|-----------|-------|------------------|
| **Longueur** | Court (3) → Long (7) | Points de coque, masse, vitesse |
| **Envergure** | Court (1) → Large (4) | Maniabilité, masse |
| **Moteurs** | 1–3 | Vitesse, capacité carburant, masse |
| **Carburant** | 20–100 | Portée opérationnelle |
| **Arme** | Canon / Stingray / Avalanche / Plasma | Puissance de feu, poids, coût |
| **Blindage** | Aucun / Alliage Léger / Alliage Lourd / Blindage Alien | Bonus coque, réduction dégâts, masse |

Le concepteur calcule les stats dérivées (vitesse, puissance de feu, coque,
rapport masse/poussée) à partir de votre configuration et affiche un aperçu ASCII
coloré. Les conceptions plus lourdes sont plus résistantes mais plus lentes —
équilibrez durabilité contre vitesse d'interception.

**Armes d'avion :**

| Arme | Dégâts | Précision | Portée | Cadence | Coût |
|------|--------|-----------|--------|---------|------|
| Canon | 15 | 85 % | 25 | 3 tirs | 5 K$ |
| Stingray | 25 | 70 % | 45 | 2 tirs | 8 K$ |
| Avalanche | 40 | 55 % | 60 | 1 tir | 12 K$ |
| Plasma | 60 | 50 % | 50 | 1 tir | 20 K$ |

**Blindage d'avion :**

| Blindage | Bonus Coque | Réduction Dégâts | Coût |
|----------|-------------|------------------|------|
| Aucun | 0 | 0 % | Gratuit |
| Alliage Léger | +10 | 10 % | 8 K$ |
| Alliage Lourd | +25 | 25 % | 18 K$ |
| Blindage Alien | +40 | 40 % | 35 K$ |

### Champ de Bataille

| Touche | Action |
|--------|--------|
| Touches fléchées / WASD / hjkl | Déplacer le curseur |
| Space / Entrée | Sélectionner l'unité / confirmer |
| q | Cycler les soldats |
| f | Tirer |
| Tab | Cycler le mode de tir |
| r | Recharger |
| e / n | Fin du tour |
| c | S'accroupir |
| g | Lancer une grenade |
| m | Mode déplacement |
| h | Utiliser le médikit |
| p | Attaque psi |
| y | Scanner de mouvement |
| t | Poser une mine de proximité |
| v | Cycler le mode de vision |
| o | Options |
| ? | Aide |
| Esc | Annuler / désélectionner |

### Contrôles Tactiles Mobiles

Sur navigateur avec écran étroit (cols < 100) ou quand `touch_mode` est activé :

| Geste | Action |
|-------|--------|
| Tap | Sélectionner, déplacer, tirer |
| Appui long (500 ms) | Annuler |
| Glisser vertical | Défiler |

Un bouton `[=]` ouvre un menu de contrôle tactile à l'écran.

---

## Conseils & Stratégie

### Début de Partie

1. Recherchez **Alliages Extraterrestres** en premier — débloque Armes Laser et Armures.
2. Construisez un second **Radar** — plus de détection, plus de financement mensuel.
3. Engagez 2–4 soldats supplémentaires pour remplir vos escouades.
4. Utilisez le **Concepteur d'Armes** pour créer des fusils sur mesure — vous pouvez
   construire de meilleures armes que les modèles standard avec les bons composants.
5. N'ignorez pas les autopsies — certaines technos d'armes les requièrent.
6. Concevez vos intercepteurs dans le **Concepteur d'Avions** — une conception
   équilibrée (longueur/envergure moyenne, 2 moteurs, missiles Stingray) surpasse
   les intercepteurs par défaut.

### Combat

- **Utilisez la couverture** — murs (80 %) > roches (70 %) > arbres (60 %) > buissons (40 %)
- **Accroupissez-vous** avant de tirer pour meilleure précision et réduction des dégâts
- **Grenades** contournent la couverture et détruisent les murs — parfait pour les ennemis retranchés
- **Apprenez les résistances alien** — consultez l'encyclopédie après les premiers kills
- **Ne vous étirez pas trop** — les aliens tirent en réaction quand vous bougez dans leur ligne de vue
- **Gardez un médecin** — un soldat avec médikit peut sauver des vies
- **Gérez l'encombrement** — ne surchargez pas les soldats d'équipement lourd

### Économie

- Vendez l'excès de cadavres et butin alien pour du cash
- Les salaires mensuels s'accumulent — équilibrez votre effectif contre les revenus
- Les missions Conseil paient un bonus de 100 K$ — priorisez-les
- Fabriquez des objets à vendre pour profit en début de partie

### Chemin de Recherche

Alliages → Armes Laser → Armure Personnelle → Autopsies → Élérium → Armes à Plasma

Milieu de partie : Tenue Intermédiaire, Plasma Lourd.
Fin de partie : Tenue de Puissance/Vol, Contrôle Mental.

### Construction de Base

- Les Radars se rentabilisent (+50 K$/mois chacun)
- Construisez un Entrepôt tôt — vous le remplirez vite
- Le Conteneur à Extraterrestres est nécessaire pour captures vivantes et bonus d'interrogatoire
- Les installations adjacentes se boostent mutuellement — planifiez la disposition de votre base
- Construisez un Labo Psi si vous voulez des capacités psi


