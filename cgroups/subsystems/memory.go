package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type MemorySubsystem struct {
	enable bool
}

func (s *MemorySubsystem) Name() string {
	return "memory"
}

func (s *MemorySubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		// 设置这个cgroup的内存限制，即将限制写入到cgroup对应目录 的memory.limit_in_bytes文件中
		if res.MemoryLimit != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroups memory failed %v", err)
			} else {
				s.enable = true
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *MemorySubsystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	 } else {
	 	return err
	}
}

func (s *MemorySubsystem) Apply(cgroupPath string, pid int) error {
	if !s.enable {
		return nil
	}
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		// 把进程的PID写到cgroup的虚拟文件系统对应目录 下的"tasks"文件
		fmt.Println("path: ", path.Join(subsysCgroupPath, "tasks"), "pid",  pid)
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			fmt.Println("err1: ", err)
			return fmt.Errorf("set cgroups proc fail %v", err)
		}
		return nil
	} else {
		fmt.Println("err1: ", err)
		return fmt.Errorf("get cgroups %s error: %v", cgroupPath, err)
	}
}