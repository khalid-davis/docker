package cgroups

import (
	"docker/cgroups/subsystems"
)

type CgroupManager struct {
	//  cgroup在hierachy中的路径，相当于创建的cgroup目录相对于root cgroup目录的路径
	Path string
	// 资源配置
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string, res *subsystems.ResourceConfig) *CgroupManager {
	return &CgroupManager{
		Path: path,
		Resource: res,
	}
}

// 将进程加入到这个cgroup中，这个cgroup会有不同的subsystem
func (c *CgroupManager) Apply(pid int) error {
	for _, subsystemIns := range subsystems.SubsystemsIns {
		if err := subsystemIns.Apply(c.Path, pid); err != nil {
			return err
		}
	}
	return nil
}

// 设置cgroup资源限制
func (c *CgroupManager) Set() error {
	for _, subsystemIns := range subsystems.SubsystemsIns {
		if err := subsystemIns.Set(c.Path, c.Resource); err != nil {
			return err
		}
	}
	return nil
}

// 释放cgroup
func (c *CgroupManager) Destroy() error {
	for _, subsystemIns := range subsystems.SubsystemsIns {
		if err := subsystemIns.Remove(c.Path); err != nil {
			return err
		}
	}
	return nil
}