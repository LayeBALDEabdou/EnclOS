/*
Copyright © 2026 Abdoulaye BALDE <twoylit@gmail.com>
*/
package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/LayeBALDEabdou/enclos/bpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/spf13/cobra"
)

// Event correspond à event_t dans execve_tracker.c (même ordre de champs).
type Event struct {
	Pid      uint32
	Ppid     uint32
	Filename [256]byte
}

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Détecte les dépendances système d'une commande",
	Long:  `Lance une commande et observe tous les binaires et librairies qu'elle utilise.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Erreur : spécifie une commande (ex: enclos track ./build.sh)")
			return
		}

		if err := rlimit.RemoveMemlock(); err != nil {
			fmt.Println("Erreur : enclos n'a pas les permissions nécessaires (es tu root ?)")
			return
		}

		var objs bpf.BpfObjects
		if err := bpf.LoadBpfObjects(&objs, nil); err != nil {
			fmt.Println("Erreur : enclos n'a pas pu démarrer (tu es root ?)")
			return
		}
		defer objs.Close()

		tpExecve, err := link.Tracepoint("syscalls", "sys_enter_execve", objs.TraceExecve, nil)
		if err != nil {
			fmt.Println("Erreur : enclos n'a pas pu s'attacher au kernel :", err)
			return
		}
		defer tpExecve.Close()

		tpOpenat, err := link.Tracepoint("syscalls", "sys_enter_openat", objs.TraceOpenat, nil)
		if err != nil {
			fmt.Println("Erreur : enclos n'a pas pu intercepter les librairies :", err)
			return
		}
		defer tpOpenat.Close()

		// Le reader est créé AVANT de lancer la commande pour ne rater aucun événement.
		// Il est fermé explicitement après Wait() pour arrêter la goroutine proprement.
		rd, err := ringbuf.NewReader(objs.Events)
		if err != nil {
			fmt.Println("Erreur : enclos n'a pas pu démarrer l'écoute :", err)
			return
		}

		sysCmd := exec.Command(args[0], args[1:]...)
		sysCmd.Stdout = os.Stdout
		sysCmd.Stderr = os.Stderr

		fmt.Println("enclos analyse :", args)
		fmt.Println("Ctrl+C pour arrêter l'analyse et générer enclave.lock")
		fmt.Println("--------------------------------------------")

		if err := sysCmd.Start(); err != nil {
			rd.Close()
			fmt.Println("Erreur : impossible de lancer la commande :", err)
			return
		}

		cmdPID := sysCmd.Process.Pid
		binaires := make(map[string]bool)
		librairies := make(map[string]bool)
		done := make(chan struct{})

		go func() {
			defer close(done)
			trackedPIDs := map[uint32]bool{uint32(cmdPID): true}

			for {
				record, err := rd.Read()
				if err != nil {
					return
				}

				var event Event
				if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &event); err != nil {
					continue
				}

				if !trackedPIDs[event.Ppid] && !trackedPIDs[event.Pid] {
					continue
				}
				trackedPIDs[event.Pid] = true

				filename := string(bytes.TrimRight(event.Filename[:], "\x00"))
				if filename == "" || filename == "/etc/ld.so.cache" {
					continue
				}

				if strings.Contains(filename, ".so.") || strings.HasSuffix(filename, ".so") {
					librairies[filename] = true
				} else {
					binaires[filename] = true
				}
			}
		}()

		// Écouter Ctrl+C pour les serveurs et commandes longues
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigChan)

		// Canal pour la fin naturelle de la commande
		cmdFini := make(chan error, 1)
		go func() { cmdFini <- sysCmd.Wait() }()

		// Attendre soit la fin de la commande, soit Ctrl+C
		select {
		case err := <-cmdFini:
			if err != nil {
				fmt.Println("Avertissement : la commande s'est terminée avec une erreur :", err)
			}
		case <-sigChan:
			fmt.Println("\nAnalyse arrêtée.")
			sysCmd.Process.Signal(syscall.SIGTERM)
			<-cmdFini
		}

		rd.Close()
		<-done

		fmt.Println("\nenclos génère enclave.lock ...")

		file, err := os.Create("enclave.lock")
		if err != nil {
			fmt.Println("Erreur : impossible de créer enclave.lock :", err)
			return
		}
		defer file.Close()

		file.WriteString("# Généré automatiquement par enclos\n")
		file.WriteString("dependencies:\n")
		file.WriteString("  binaries:\n")
		for chemin := range binaires {
			if _, err := fmt.Fprintf(file, "    - %s\n", chemin); err != nil {
				fmt.Println("Erreur : impossible d'écrire dans enclave.lock :", err)
				return
			}
		}
		file.WriteString("  libraries:\n")
		for chemin := range librairies {
			if _, err := fmt.Fprintf(file, "    - %s\n", chemin); err != nil {
				fmt.Println("Erreur : impossible d'écrire dans enclave.lock :", err)
				return
			}
		}

		fmt.Println("enclave.lock généré avec succès !")
	},
}

func init() {
	rootCmd.AddCommand(trackCmd)
	// DisableFlagParsing transmet les flags à la commande trackée (ex: enclos track ls -la)
	trackCmd.DisableFlagParsing = true
}
