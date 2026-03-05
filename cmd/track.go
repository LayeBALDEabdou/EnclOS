/*
Copyright © 2026 Abdoulaye BALDE <[EMAIL_ADDRESS]>
*/
package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"

	"github.com/LayeBALDEabdou/enclos/bpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/spf13/cobra"
)

// Étape 1 — Struct Go miroir de event_t dans execve_tracker.c
// L'ordre et les types doivent correspondre exactement au C.
type Event struct {
	Pid      uint32
	Ppid     uint32
	Filename [256]byte
}

// trackCmd represents the track command
var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "track les dependances systeme d'une commande via eBPF",
	Long:  `Avec cette commande, tu peux tracker les dependances systeme d'une commande via eBPF`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Erreur : Veuillez spécifier une commande a tracker (ex : enclos track ls -l)")
			os.Exit(1)
		}

		// Étape 2a — Lever la limite mémoire verrouillée (requis pour eBPF)
		// Sans ça, le kernel refuse de créer les maps eBPF avec EPERM.
		if err := rlimit.RemoveMemlock(); err != nil {
			fmt.Println("Erreur RemoveMemlock :", err)
			os.Exit(1)
		}

		// Étape 2b — Charger les objets eBPF (programme + maps) dans le kernel
		// LoadBpfObjects lit le .o embarqué dans bpf_bpfel.go et l'injecte dans le kernel.
		// Nécessite les droits root ou CAP_BPF.
		var objs bpf.BpfObjects
		if err := bpf.LoadBpfObjects(&objs, nil); err != nil {
			fmt.Println("Erreur chargement eBPF (root requis ?) :", err)
			os.Exit(1)
		}
		defer objs.Close()

		// Étape 3 — Attacher le programme au tracepoint sys_enter_execve
		// A partir de ce moment, chaque execve() sur la machine déclenche notre programme C.
		tp, err := link.Tracepoint("syscalls", "sys_enter_execve", objs.TraceExecve, nil)
		if err != nil {
			fmt.Println("Erreur attachement tracepoint :", err)
			os.Exit(1)
		}
		defer tp.Close()

		// Étape 4 — Lancer la commande cible normalement (sans strace)
		// On utilise Start() et non Run() pour ne pas bloquer : on doit lire le ringbuf en parallèle.
		sysCmd := exec.Command(args[0], args[1:]...)
		sysCmd.Stdout = os.Stdout
		sysCmd.Stderr = os.Stderr

		fmt.Println("tracking eBPF en cours pour la commande :", args)
		fmt.Println("------------------------------------------------------------")

		if err := sysCmd.Start(); err != nil {
			fmt.Println("Erreur lancement commande :", err)
			os.Exit(1)
		}

		cmdPID := sysCmd.Process.Pid

		// Étape 5 — Créer le lecteur du Ring Buffer
		// C'est la map "events" déclarée dans execve_tracker.c que le programme C remplit.
		rd, err := ringbuf.NewReader(objs.Events)
		if err != nil {
			fmt.Println("Erreur création ringbuf reader :", err)
			os.Exit(1)
		}

		dependances := make(map[string]bool)

		// Canal pour savoir quand la goroutine a fini de traiter les derniers événements
		done := make(chan struct{})

		go func() {
			defer close(done)
			// Set de tous les PIDs appartenant à l'arbre de processus tracké.
			// On commence avec cmdPID et on l'étend dynamiquement à chaque enfant détecté.
			trackedPIDs := map[uint32]bool{uint32(cmdPID): true}

			for {
				record, err := rd.Read()
				if err != nil {
					// Le reader a été fermé (après Wait) → on arrête proprement
					return
				}

				// Décoder les bytes bruts en struct Go
				var event Event
				if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &event); err != nil {
					continue
				}

				// Filtrer : on ne garde que les processus dont le parent est dans l'arbre tracké.
				// Si le ppid est connu, ce processus fait partie de l'arbre → on l'ajoute au set.
				if !trackedPIDs[event.Ppid] && !trackedPIDs[event.Pid] {
					continue
				}
				trackedPIDs[event.Pid] = true

				// Convertir le tableau de bytes en string (jusqu'au premier octet nul)
				filename := string(bytes.TrimRight(event.Filename[:], "\x00"))
				if filename != "" {
					dependances[filename] = true
				}
			}
		}()

		// Attendre que la commande se termine, puis fermer le reader.
		// La fermeture du reader provoque le retour de rd.Read() avec une erreur → la goroutine s'arrête.
		sysCmd.Wait()
		rd.Close()
		<-done // attendre que la goroutine ait fini de traiter les derniers événements

		// Étape 6 — Écrire enclave.lock avec les dépendances collectées
		fmt.Println("analyse des dependances ...")

		file, errFile := os.Create("enclave.lock")
		if errFile != nil {
			fmt.Println("erreur lors de la creation du fichier enclave.lock :", errFile)
			os.Exit(1)
		}
		defer file.Close()

		file.WriteString("# fichier genere automatiquement par enclos\n")
		file.WriteString("dependencies:\n")
		for chemin := range dependances {
			file.WriteString(fmt.Sprintf("  - %s\n", chemin))
		}

		fmt.Println("fichier enclave.lock genere avec succes !")
		fmt.Println("\ntracking termine avec succes")
	},
}

func init() {
	rootCmd.AddCommand(trackCmd)
	trackCmd.DisableFlagParsing = true
}
