# enclos

**Un outil Linux pour détecter automatiquement les dépendances de tes projets**

---

## Le problème

Tu clones un projet. Le README dit :

```
Prérequis : Node 18, Python 3.11
```

Tu installes tout. Tu lances le build. Ça plante :

```
Error: libssl.so.3 not found
Error: ffmpeg command not found
```

**Le projet avait des dépendances cachées que personne n'avait documentées.**

---

## Pourquoi ça arrive ?

```bash
# build.sh
npm install
python process.py
gcc -o output main.c    # Personne n'a noté que gcc était nécessaire
ffmpeg -i video.mp4     # Personne n'a noté que ffmpeg était nécessaire
```

Les développeurs oublient de documenter les dépendances système.
Résultat : "Ça marche sur ma machine" mais pas ailleurs.

---

## La solution : enclos

enclos **observe** ce que ton projet utilise réellement et génère la liste complète automatiquement.

```bash
$ enclos track ./build.sh

Analyse en cours...

Dépendances détectées :
  - node 18.0.0
  - python 3.11.2
  - gcc 13.2
  - ffmpeg 6.0
  - openssl 3.0.2 (librairie)
  - libc 2.38 (librairie)

Fichier enclave.lock généré.
```

**Tu n'oublies plus rien. Le système a tout vu.**

---

## Comment ça marche ?

Le tracking se fait au niveau du **kernel Linux** :

```
┌─────────────────────────────────────┐
│         Ton script build.sh         │
└─────────────────┬───────────────────┘
                  │ exécute gcc, python, ffmpeg...
                  ▼
┌─────────────────────────────────────┐
│           Kernel Linux              │
│  ┌───────────────────────────────┐  │
│  │   Module enclos               │  │
│  │   - Voit chaque binaire lancé │  │
│  │   - Voit chaque .so chargé    │  │
│  │   - Voit chaque fichier lu    │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
                  │
                  ▼
        Fichier enclave.lock
        (liste complète)
```

Le kernel voit **tout**. Impossible d'oublier une dépendance.

---

## Pourquoi le kernel ?

| Niveau | Ce qu'il peut voir |
|--------|-------------------|
| npm/pip | Seulement ses propres packages |
| Script bash | Peut rater des sous-dépendances |
| **Kernel** | **Tout** : binaires, librairies, fichiers |

---

## Utilisation

### 1. Détecter les dépendances d'un projet

```bash
$ cd mon-projet
$ enclos track ./build.sh
# Génère enclave.lock
```

### 2. Installer les dépendances sur une autre machine

```bash
$ git clone projet.git
$ cd projet
$ enclos install
# Installe tout ce qui est dans enclave.lock
```

### 3. Ajouter des dépendances manuellement

```bash
$ enclos add node@18
$ enclos add python@3.11
```

### 4. Voir les dépendances actuelles

```bash
$ enclos list
- node 18.0.0
- python 3.11.2
- gcc 13.2
```

---

## Le fichier enclave.lock

```yaml
name: mon-projet
tracked: true
dependencies:
  binaries:
    - name: node
      version: 18.0.0
    - name: python
      version: 3.11.2
    - name: gcc
      version: 13.2
    - name: ffmpeg
      version: 6.0
  libraries:
    - name: libssl
      version: 3.0.2
    - name: libc
      version: 2.38
```

Tu le commit dans git → tout le monde a le même environnement.

---

## Ce qui existe déjà vs enclos

| Outil | Ce qu'il fait | Limitation |
|-------|---------------|------------|
| mise/asdf | Gère les versions | Tu dois lister les dépendances toi-même |
| Docker | Environnement isolé | Tu dois écrire le Dockerfile toi-même |
| Nix | Reproductibilité | Tu dois écrire le flake.nix toi-même |
| **enclos** | Détection auto | Le kernel trouve tout pour toi |

**L'innovation d'enclos = la détection automatique via le kernel.**

---

## Architecture

```
enclos/
├── kernel/              # Module kernel Linux (C)
│   ├── enclos.c         # Hook sur execve, open, mmap
│   └── Makefile
├── cli/                 # Outil en ligne de commande (Go)
│   ├── main.go          # Point d'entrée
│   ├── cmd/             # Commandes (track, install, add...)
│   └── go.mod
└── shell/               # Intégration shell
    └── enclos.sh        # Activation auto au cd
```

---

## Installation (Fedora)

```bash
# Dépendances
sudo dnf install kernel-devel kernel-headers gcc make golang

# Compiler le module kernel
cd kernel/
make
sudo insmod enclos.ko

# Compiler et installer la CLI
cd ../cli/
go build -o enclos
sudo mv enclos /usr/local/bin/
```

---

## Roadmap

### Phase 1 - CLI basique (sans kernel)
- [ ] Commandes `enclos add`, `enclos list`, `enclos install`
- [ ] Format `enclave.lock`
- [ ] Gestion des packages dans `~/.enclos/`

### Phase 2 - Module kernel
- [ ] Hook sur `execve` (détecter les binaires lancés)
- [ ] Hook sur `open` (détecter les fichiers lus)
- [ ] Hook sur `mmap` (détecter les .so chargés)
- [ ] Communication kernel ↔ userspace

### Phase 3 - Intégration
- [ ] Activation automatique au `cd`
- [ ] Résolution et téléchargement des packages
- [ ] Support multi-distro

---

## Stack technique

- **Module kernel** : C
- **CLI** : Go
- **Cible** : Linux (Fedora, Ubuntu, Arch...)
- **Tests** : Fedora 43

---

## Collaboration
-pour toute envie de collaboration contactez moi

*enclos - Parce que documenter ses dépendances manuellement, c'est has been.*
