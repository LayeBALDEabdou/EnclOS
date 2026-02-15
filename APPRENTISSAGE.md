# Ce que tu dois apprendre pour créer enclos

---

## Partie 1 : Go (pour la CLI)

### Niveau 1 - Les bases (1-2 semaines)

| Concept | Pourquoi c'est nécessaire | Exemple |
|---------|---------------------------|---------|
| Variables et types | Stocker des données | `var name string = "node"` |
| Fonctions | Organiser le code | `func track(cmd string) {}` |
| Structures (struct) | Représenter une dépendance | `type Dependency struct { Name, Version string }` |
| Slices et maps | Listes et dictionnaires | `deps := []Dependency{}` |
| Gestion d'erreurs | Go n'a pas d'exceptions | `if err != nil { return err }` |
| Packages | Organiser en modules | `package main` |

### Niveau 2 - Pour la CLI (1 semaine)

| Concept | Pourquoi c'est nécessaire | Exemple |
|---------|---------------------------|---------|
| Arguments CLI | Lire `enclos add node@18` | `os.Args` |
| Lire/écrire fichiers | Gérer enclave.lock | `os.ReadFile()`, `os.WriteFile()` |
| JSON/YAML | Parser enclave.lock | `json.Unmarshal()` |
| Exécuter commandes | Lancer des programmes | `exec.Command("node", "--version")` |

### Niveau 3 - Avancé (selon besoins)

| Concept | Pourquoi c'est nécessaire |
|---------|---------------------------|
| Goroutines | Télécharger plusieurs packages en parallèle |
| Channels | Communication entre goroutines |
| Interfaces | Code flexible et testable |

### Ressources Go

| Ressource | Lien |
|-----------|------|
| Tour of Go (officiel) | https://go.dev/tour |
| Go by Example | https://gobyexample.com |
| Effective Go | https://go.dev/doc/effective_go |

---

## Partie 2 : C (pour le module kernel)

### Niveau 1 - Les bases du C (2 semaines)

| Concept | Pourquoi c'est nécessaire | Exemple |
|---------|---------------------------|---------|
| Variables et types | Stocker des données | `int pid = 1234;` |
| Pointeurs | Manipuler la mémoire | `char *filename` |
| Structures | Représenter des données | `struct dependency {}` |
| Allocation mémoire | Gérer la mémoire | `malloc()`, `free()` |
| Préprocesseur | Macros kernel | `#include`, `#define` |

### Niveau 2 - Spécifique kernel (2-3 semaines)

| Concept | Pourquoi c'est nécessaire | Exemple |
|---------|---------------------------|---------|
| Modules kernel | Créer un .ko chargeable | `module_init()`, `module_exit()` |
| Kprobes | Intercepter des fonctions kernel | Hook sur `execve` |
| Netlink | Communiquer kernel ↔ userspace | Envoyer les dépendances à la CLI |
| Listes kernel | Stocker des données | `struct list_head` |
| Spinlocks | Éviter les race conditions | `spin_lock()` |

### Les syscalls à comprendre

| Syscall | Ce qu'il fait | Pourquoi on l'intercepte |
|---------|---------------|--------------------------|
| `execve` | Lance un programme | Détecter les binaires exécutés |
| `open` | Ouvre un fichier | Détecter les fichiers lus |
| `mmap` | Mappe en mémoire | Détecter les .so chargés |

### Ressources C et Kernel

| Ressource | Lien |
|-----------|------|
| Learn C | https://www.learn-c.org |
| The Linux Kernel Module Programming Guide | https://sysprog21.github.io/lkmpg |
| Linux Kernel Development (livre) | Robert Love |

---

## Ordre d'apprentissage recommandé

```
Semaine 1-2    →  Bases de Go
Semaine 3      →  CLI en Go (arguments, fichiers, JSON)
Semaine 4-5    →  Bases du C
Semaine 6-7    →  Modules kernel Linux
Semaine 8+     →  Kprobes et Netlink
```

---

## Projets pratiques pour apprendre

### Go - Avant de coder enclos

| Projet | Ce que tu apprends |
|--------|-------------------|
| 1. CLI todo list | Arguments, fichiers, JSON |
| 2. Téléchargeur de fichiers | HTTP, goroutines |
| 3. Parseur de package.json | Lecture JSON, structures |

### C - Avant de coder le module kernel

| Projet | Ce que tu apprends |
|--------|-------------------|
| 1. Programme qui liste les fichiers | Syscalls de base |
| 2. Module kernel "hello world" | Structure d'un module |
| 3. Module qui log les execve | Kprobes |

---

## Ce que tu peux ignorer (pour l'instant)

### Go - Pas besoin pour ce projet
- Génériques (Go 1.18+)
- Reflection
- CGO (appeler du C depuis Go)
- Web/HTTP servers

### C - Pas besoin pour ce projet
- Réseaux (sockets)
- Threads POSIX
- GUI

---

## Estimation du temps

| Si tu connais... | Temps pour être prêt |
|------------------|---------------------|
| Aucun langage | 2-3 mois |
| Python/JS | 1-2 mois |
| C ou Go déjà | 2-4 semaines |
| Les deux | Tu peux commencer |

---

## Chemin le plus court

Si tu veux aller vite :

1. **Commence par la CLI en Go** (plus simple)
2. **Fais une version sans module kernel** (utilise `strace` pour tracker)
3. **Ajoute le module kernel après** (quand la CLI marche)

```bash
# Version 1 : sans module kernel (plus simple)
strace -f -e execve ./build.sh 2>&1 | enclos parse

# Version 2 : avec module kernel (plus tard)
enclos track ./build.sh
```

---

*Tu n'as pas besoin de tout maîtriser avant de commencer. Apprends en construisant.*
