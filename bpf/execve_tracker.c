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


struct  execve_args{
    unsigned short common_type;
    unsigned char common_flag;
    unsigned char common_preempt_count;
    int common_pid;
    int __syscall_nr;
    const char *filename;
    const char *const *argv;
    const char *const *evenp;

};

SEC("tracepoint/syscalls/sys_enter_execve")
int trace_execve(struct execve_args *ctx)
{
    struct event_t *event=bpf_ringbuf_reserve(&events,sizeof(struct event_t) ,0);

    if (!event){
        return 0;
    }


    __u64 pid_tgid=bpf_get_current_pid_tgid();
    event->pid=pid_tgid >> 32;

    struct task_struct *task=(struct task_struct *)bpf_get_current_task();
    event->ppid = BPF_CORE_READ(task, real_parent, tgid);

    bpf_probe_read_user_str(event->filename,sizeof(event->filename),ctx->filename);
    bpf_ringbuf_submit(event,0);
    return 0;
};

char LICENSE[] SEC("license")="Dual BSD/GPL";

