
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <asm/uaccess.h>
#include <linux/hugetlb.h>
#include <linux/kernel.h>
#include <linux/module.h>
#include <linux/init.h>
#include <linux/sched/signal.h>
#include <linux/sched.h>
#include <linux/fs.h>
 
MODULE_LICENSE("GPL");
MODULE_DESCRIPTION("Escribe información de ram");
MODULE_AUTHOR("Jerson Villatoro - 201442819");
 
struct task_struct *task;
struct task_struct *task_child;
struct list_head *list;

static char * getEstado(long s){
    if(s == 0){
        return "runing";
    }else if(s == 1){
        return "sleeping";
    }else if(s == 2){
        return "uninterruptible";
    }else if(s == 4){
        return "zombie";
    }
    else if(s == 8){
        return "stopped";
    }
    else if(s == 1026){
        return "noload";
    }
    else{
        return "extranio";
    }
}

static int escribir_archivo(struct seq_file *m, void *v)
{
    seq_printf(m,"Nombre Estudiante 1: JERSON VILLATORO\n");
	seq_printf(m,"Carnet 1: 2014-42819\n\n\n");

    for_each_process( task ){
        seq_printf(m,"\nPID: %d NOMBRE: %s ESTADO: %s\n",task->pid, task->comm, getEstado(task->state));
        seq_printf(m,"HIJOS:\n");
        list_for_each(list, &task->children)
        {
            task_child = list_entry(list, struct task_struct, sibling );
            seq_printf(m,"PID: %d NOMBRE: %s ESTADO: %s\n",task_child->pid, task_child->comm, getEstado(task_child->state));
        }
        seq_printf(m,"-----------------------------------------------------\n");
    }
	return 0;
}


static int al_abrir(struct inode *inode, struct file *file){
	return single_open(file, escribir_archivo, NULL);
}

static struct file_operations operaciones = 
{
	.open = al_abrir,
	.read = seq_read
};

static int iniciar(void)
{
	proc_create("cpu_201442819", 0, NULL, &operaciones);
	printk(KERN_INFO "Carnet: 2014-42819\n");
	return 0;
}

static void salir(void)
{
	remove_proc_entry("cpu_201442819", NULL);
	printk(KERN_INFO "Vacaciones Junio 2020: Sistemas Operativo 1");
}
 
module_init(iniciar);
module_exit(salir);