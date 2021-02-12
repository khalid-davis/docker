package main

import (
	"docker/cgroups"
	"docker/cgroups/subsystems"
	"docker/container"
	"github.com/sirupsen/logrus"
)

func Run(tty bool, command string, res *subsystems.ResourceConfig) {
	parent := container.NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		logrus.Error(err, "maybe you need to input 'mount -t proc proc /proc' on terminal")
		return
	}
	// 创建cgroup manager， 并通过调用set和apply设置资源限制并使限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup", res)
	defer cgroupManager.Destroy()

	err := cgroupManager.Set()
	if err != nil {
		logrus.Error("error: ", err)
		return
	}
	err = cgroupManager.Apply(parent.Process.Pid)
	if err != nil {
		logrus.Error("error: ", err)
		return
	}
	parent.Wait()
}

