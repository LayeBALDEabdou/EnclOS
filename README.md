<p align="center">
  <img src="./enclos-logo.svg" alt="Enclos Logo" width="300" />
</p>

# enclos

**Un outil Linux pour détecter automatiquement les dépendances de tes projets, par l'observation directe du kernel.**

---

## Le problème

Tu clones un projet. Le `README` dit :

```text
Prérequis : Node 18, Python 3.11
```

Tu installes tout. Tu lances le build. Ça plante :

```text
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

`enclos` **observe** ce que ton projet utilise réellement pendant son exécution et génère la liste complète automatiquement.

```bash
$ enclos track ./build.sh

Analyse en cours via eBPF...

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

Le tracking se fait au niveau du **kernel Linux**, de manière moderne et sécurisée grâce à **eBPF** :

```text
┌─────────────────────────────────────┐
│         Ton script build.sh         │
└─────────────────┬───────────────────┘
                  │ exécute gcc, python, ffmpeg...
                  ▼
┌─────────────────────────────────────┐
│           Kernel Linux              │
│  ┌───────────────────────────────┐  │
│  │   Programme eBPF (Sandboxé)   │  │
│  │   - Intercepte execve()       │  │
│  │   - Intercepte mmap() (.so)   │  │
│  │   - Intercepte openat()       │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
                  │
                  ▼
         Fichier enclave.lock
         (liste complète et prouvée)
```

Le kernel voit **tout**. Impossible d'oublier une dépendance.

---

## Pourquoi l'observabilité Kernel (eBPF) ?

| Niveau | Ce qu'il peut voir | Risque de crash ? |
|--------|-------------------|-------------------|
| npm/pip | Seulement ses propres packages | Aucun |
| Script bash | Peut rater des sous-dépendances | Aucun |
| Module Kernel (.ko) | **Tout** : binaires, librairies... | **Élevé** (Kernel panic) |
| **eBPF (enclos)** | **Tout** : binaires, librairies... | **Aucun** (Vérifié et sandboxé par Linux) |

---

## Utilisation

### 1. Détecter les dépendances de n'importe quel projet

`enclos` s'en fiche du langage ou de l'outil de build. Il observe la réalité de l'exécution pour **n'importe quelle commande Linux** :

```bash
# Pour un script Bash
$ enclos track ./build.sh

# Pour un projet Python
$ enclos track python main.py

# Pour un projet Node.js / Frontend
$ enclos track npm run build

# Pour un projet C/C++
$ enclos track make all

# Le fichier enclave.lock généré liste la vérité absolue des binaires et librairies utilisés
```

### 2. Le "Killer Feature" : Exporter la configuration

Pourquoi réécrire un fichier Docker à la main quand `enclos` connait les dépendances exactes ?

```bash
$ enclos export docker > Dockerfile
# Génère un Dockerfile (ex: Alpine/Ubuntu) incluant toutes les dépendances système

$ enclos export nix > flake.nix
# Génère une configuration Nix reproductible
```

### 3. Installer les dépendances sur une autre machine

```bash
$ git clone projet.git
$ cd projet
$ enclos install
# Installe localement tout ce qui est dans enclave.lock
```

### 4. Gérer manuellement

```bash
$ enclos add node@18
$ enclos list
- node 18.0.0
- python 3.11.2
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
  libraries:
    - name: libssl
      version: 3.0.2
```

Tu le commit dans git → tout le monde a le même environnement de départ.

---

## Ce qui existe déjà vs enclos

| Outil | Ce qu'il fait | Limitation |
|-------|---------------|------------|
| mise/asdf | Gère les versions | Tu dois lister les dépendances toi-même |
| Docker | Environnement isolé | Tu dois écrire le Dockerfile toi-même |
| Nix | Reproductibilité | Tu dois écrire le flake.nix toi-même |
| ReproZip | Trace l'exécution | Lourd, orienté data-science, archive tout |
| **enclos** | **Détection eBPF + Export** | **Génère pour toi (Docker/Nix) par la preuve** |

---

## Architecture (eBPF + Go)

```text
enclos/
├── ebpf/                # Programmes C eBPF (compilés vers BPF bytecode)
│   ├── intercept_exec.c # Hook sur sys_enter_execve
│   ├── intercept_mmap.c # Hook pour les librairies partagées
│   └── Makefile
├── cli/                 # Outil en ligne de commande (Go)
│   ├── main.go          # Point d'entrée
│   ├── parser/          # Analyse des traces et résolution des paquets
│   ├── bpf_loader.go    # Charge et communique avec le code eBPF via cilium/ebpf
│   ├── cmd/             # Commandes (track, export, install...)
│   └── go.mod
└── shell/               # Intégration shell optionnelle
    └── enclos.sh
```

---

## Installation (Développement)

```bash
# Dépendances requises : Go, Clang/LLVM (pour compiler le C vers eBPF), kernel headers
sudo dnf install clang llvm golang kernel-devel

# Cloner le projet
git clone <url-du-repo>
cd enclos

# Compiler l'outil (Go compile et embarque automatiquement le bytecode eBPF avec 'go generate')
cd cli
go generate ./...
go build -o enclos
sudo mv enclos /usr/local/bin/
```

---

## Roadmap

### Phase 1 - MVP & Le Pilote "strace"
- [ ] Créer le CLI de base (`add`, `list`) en **Go**.
- [ ] Piloter `strace -f -e trace=execve` sous le capot pour simuler l'interception.
- [ ] Parser la sortie de strace pour générer un premier `enclave.lock`.

### Phase 2 - Les Exports Magiques
- [ ] Commande `enclos export docker` pour générer un Dockerfile depuis le `.lock`.
- [ ] Commande `enclos export nix` pour s'interfacer avec l'écosystème Nix.

### Phase 3 - Moteur natif eBPF
- [ ] Remplacer l'appel `strace` par un programme **eBPF** (en C restreint).
- [ ] Charger le programme eBPF depuis Go (`cilium/ebpf`).
- [ ] Intercepter `openat` et `mmap` pour tracer avec précision les bibliothèques C (`.so`).

### Phase 4 - Intégration & Résolution
- [ ] Mapping intelligent entre les binaires appelés et les noms de paquets (ex: `gcc` -> paquet `gcc` ou `build-essential`).
- [ ] Support multi-distro.

---

## Stack technique

- **Moteur d'interception** : eBPF (C)
- **CLI & Logique** : Go (avec `cilium/ebpf`)
- **Cible** : Linux (Fedora, Ubuntu, Arch...)
- **Tests** : Fedora 43

---

## Documentation & Ressources eBPF

Si tu souhaites comprendre la magie sous le capot d'Enclos, voici les meilleures ressources pour découvrir eBPF :

- [eBPF.io (Site Officiel)](https://ebpf.io/) - L'introduction parfaite à la technologie par la fondation eBPF.
- [Cilium eBPF Go Library](https://ebpf-go.dev/) - La documentation de la librairie Go que nous utilisons pour charger nos programmes eBPF.
- [BPF Compiler Collection (BCC)](https://github.com/iovisor/bcc) - Le dépôt historique contenant d'excellents exemples et tutos (notamment `execsnoop`).
- [Learning eBPF par Liz Rice](https://isovalent.com/books/learning-ebpf/) - Le livre de référence absolu pour comprendre comment observer le kernel Linux.

---

## Collaboration
- Pour toute envie de collaboration, contactez-moi !

*enclos - Parce que documenter ses dépendances manuellement, c'est has been.*
