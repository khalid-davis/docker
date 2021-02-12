package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
	"testing"
)

const (
	// 挂载了memory subsystem的hierarchy的根目录位置
	cgroupMemoryHierarchyMount = "/sys/fs/cgroups/memory"
)

func TestCgroup(t *testing.T) {
	if os.Args[0] == "/proc/self/exe" {
		// 容器进程
		fmt.Printf("current pid %d \n", syscall.Getpid())
		cmd := exec.Command("sh", "-c", `stress --vm-bytes 200m --vm-keep -m 1`)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	//得到fork出来进程映射在外部命名空间的pid
	fmt.Printf("%+v\n", cmd.Process.Pid)

	// 在系统默认创建挂载了memory subsystem的hierarchy上创建cgroup
	newCgroup := path.Join(cgroupMemoryHierarchyMount, "cgroups-demo-memory-3")
	// 如果想删除cgroup的目录，需要先umount <dir-name>，再删除
	//if err := os.Mkdir(newCgroup, 0755); err != nil {
	//	panic(err)
	//}
	// 将容器进程放到子cgroup中
	if err := ioutil.WriteFile(path.Join(newCgroup, "tasks"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		panic(err)
	}
	// 限制cgroup的内存使用
	// 这里最大分配400m，上面的stress会模拟压力
	if err := ioutil.WriteFile(path.Join(newCgroup, "memory.limit_in_bytes"), []byte("400m"), 0644); err != nil {
		panic(err)
	}
	cmd.Process.Wait()
}
