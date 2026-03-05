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
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/spf13/cobra"
)

// Event correspond exactement à event_t dans execve_tracker.c.
// L'ordre des champs doit être identique pour que le décodage binaire fonctionne.
type Event struct {
	Pid      uint32
	Ppid     uint32
	Filename [256]byte
}

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Détecte les dépendances système d'une commande",
	Long:  `Lance une commande et observe tous les binaires qu'elle exécute.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Erreur : spécifie une commande (ex: enclos track ./build.sh)")
			return
		}

		// 1. Autoriser l'utilisation de mémoire nécessaire à l'observation kernel
		if err := rlimit.RemoveMemlock(); err != nil {
			fmt.Println("Erreur : enclos n'a pas les permissions nécessaires (tu es root ?)")
			return
		}

		// 2. Démarrer l'observateur kernel (nécessite root)
		var objs bpf.BpfObjects
		if err := bpf.LoadBpfObjects(&objs, nil); err != nil {
			fmt.Println("Erreur : enclos n'a pas pu démarrer (tu es root ?)")
			return
		}
		defer objs.Close()

		// 3. Activer l'interception de chaque exécution de programme
		tp, err := link.Tracepoint("syscalls", "sys_enter_execve", objs.TraceExecve, nil)
		if err != nil {
			fmt.Println("Erreur : enclos n'a pas pu s'attacher au kernel :", err)
			return
		}
		defer tp.Close()

		// 4. Préparer la réception des événements AVANT de lancer la commande
		// pour ne rater aucune exécution dès le démarrage.
		rd, err := ringbuf.NewReader(objs.Events)
		if err != nil {
			fmt.Println("Erreur : enclos n'a pas pu démarrer l'écoute :", err)
			return
		}
		// rd est fermé explicitement plus bas, après sysCmd.Wait(),
		// ce qui provoque l'arrêt propre de la goroutine de lecture.

		// 5. Lancer la commande à analyser
		sysCmd := exec.Command(args[0], args[1:]...)
		sysCmd.Stdout = os.Stdout
		sysCmd.Stderr = os.Stderr

		fmt.Println("enclos analyse :", args)
		fmt.Println("--------------------------------------------")

		if err := sysCmd.Start(); err != nil {
			rd.Close()
			fmt.Println("Erreur : impossible de lancer la commande :", err)
			return
		}

		cmdPID := sysCmd.Process.Pid

		// 6. Lire les événements dans une goroutine parallèle pendant que la commande tourne
		dependances := make(map[string]bool)
		done := make(chan struct{})

		go func() {
			defer close(done)

			// On suit tous les PIDs de la commande et de ses processus enfants.
			trackedPIDs := map[uint32]bool{uint32(cmdPID): true}

			for {
				record, err := rd.Read()
				if err != nil {
					// rd.Close() a été appelé → sortie propre
					return
				}

				// Décoder les bytes reçus en struct Event
				var event Event
				if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &event); err != nil {
					continue
				}

				// Ignorer tout processus qui n'appartient pas à l'arbre de la commande trackée
				if !trackedPIDs[event.Ppid] && !trackedPIDs[event.Pid] {
					continue
				}
				trackedPIDs[event.Pid] = true

				// Convertir le nom de fichier (tableau C de bytes) en string Go
				filename := string(bytes.TrimRight(event.Filename[:], "\x00"))
				if filename != "" {
					dependances[filename] = true
				}
			}
		}()

		// Attendre la fin de la commande
		if err := sysCmd.Wait(); err != nil {
			fmt.Println("Avertissement : la commande s'est terminée avec une erreur :", err)
		}

		// Fermer l'écoute → fait sortir la goroutine
		rd.Close()
		// Attendre que tous les événements restants soient traités
		<-done

		// 7. Écrire le fichier enclave.lock
		fmt.Println("\nenclos génère enclave.lock ...")

		file, err := os.Create("enclave.lock")
		if err != nil {
			fmt.Println("Erreur : impossible de créer enclave.lock :", err)
			return
		}
		defer file.Close()

		file.WriteString("# Généré automatiquement par enclos\n")
		file.WriteString("dependencies:\n")
		for chemin := range dependances {
			if _, err := file.WriteString(fmt.Sprintf("  - %s\n", chemin)); err != nil {
				fmt.Println("Erreur : impossible d'écrire dans enclave.lock :", err)
				return
			}
		}

		fmt.Println("enclave.lock généré avec succès !")
	},
}

func init() {
	rootCmd.AddCommand(trackCmd)
	trackCmd.DisableFlagParsing = true
}
