package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpusetSubSystem struct {
	enable bool
}

func (s *CpusetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.CpuSet != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"), []byte(res.CpuSet), 0644); err != nil {
				return fmt.Errorf("set cgroups cpuset fail %v", err)
			} else {
				s.enable = true
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	if !s.enable {
		return nil
	}
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroups proc fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroups %s error: %v", cgroupPath, err)
	}
}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}
