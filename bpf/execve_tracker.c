//go:build ignore

// Le header vmlinux.h contient toutes les structures du noyau.
// IMPORTANT : Il ne faut SURTOUT PAS inclure <linux/bpf.h> ou d'autres headers <linux/...> 
// en même temps que vmlinux.h, sinon il y a des conflits de définitions.
#include "vmlinux.h"

// Headers eBPF standards fournis par libbpf
#include <bpf/bpf_helpers.h>

#include <bpf/bpf_core_read.h>

#define PATH_MAX_LEN 256

struct  event_t
{
    __u32 pid;
    __u32 ppid;
    __u8 filename[PATH_MAX_LEN];
};


struct 
{
    __uint(type,BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries,256*1024);
} events SEC(".maps");


// Arguments du tracepoint sys_enter_execve
struct execve_args {
    unsigned short common_type;
    unsigned char common_flag;
    unsigned char common_preempt_count;
    int common_pid;
    int __syscall_nr;
    const char *filename;
    const char *const *argv;
    const char *const *envp;
};

// Arguments du tracepoint sys_enter_openat.
// Sur x86_64, les arguments syscall sont stockés en 8 octets (long),
// même si leur type réel est int. Utiliser long pour dfd et flags
// garantit le bon alignement et donc la bonne adresse pour filename.
struct openat_args {
    unsigned short common_type;
    unsigned char common_flag;
    unsigned char common_preempt_count;
    int common_pid;
    int __syscall_nr;
    long dfd;
    const char *filename;
    long flags;
    unsigned short mode;
};

// Remplit et envoie un event dans le ring buffer pour un chemin donné.
// Retourne 0 dans tous les cas (convention eBPF).
static __always_inline int envoyer_event(const char *chemin_utilisateur)
{
    struct event_t *event = bpf_ringbuf_reserve(&events, sizeof(struct event_t), 0);
    if (!event)
        return 0;

    __u64 pid_tgid = bpf_get_current_pid_tgid();
    event->pid = pid_tgid >> 32;

    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    event->ppid = BPF_CORE_READ(task, real_parent, tgid);

    bpf_probe_read_user_str(event->filename, sizeof(event->filename), chemin_utilisateur);
    bpf_ringbuf_submit(event, 0);
    return 0;
}

// Intercepte chaque exécution de programme (ex: /usr/bin/python, /usr/bin/gcc...)
SEC("tracepoint/syscalls/sys_enter_execve")
int trace_execve(struct execve_args *ctx)
{
    return envoyer_event(ctx->filename);
}

// Intercepte chaque ouverture de fichier.
// On filtre sur ".so" pour ne garder que les librairies partagées.
SEC("tracepoint/syscalls/sys_enter_openat")
int trace_openat(struct openat_args *ctx)
{
    char chemin[PATH_MAX_LEN];
    bpf_probe_read_user_str(chemin, sizeof(chemin), ctx->filename);

    // Parcourir le chemin pour détecter la présence de ".so"
    for (int i = 0; i < PATH_MAX_LEN - 3; i++) {
        if (chemin[i] == '.' && chemin[i+1] == 's' && chemin[i+2] == 'o') {
            // C'est une librairie .so → on envoie l'événement
            return envoyer_event(ctx->filename);
        }
        if (chemin[i] == '\0')
            break;
    }

    return 0;
}

char LICENSE[] SEC("license") = "Dual BSD/GPL";

