package main

import (
	"docker/cgroups/subsystems"
	"docker/container"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Crate a container with naemspace and cgroups limit mydocker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name: "ti",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name: "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name: "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name: "cpuset",
			Usage: "cpuset limit",
		},
	},
	Action: func(context *cli.Context) error {
		logrus.Info("run command action start")
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		cmd := context.Args().Get(0)
		tty := context.Bool("ti")
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet: context.String("cpuset"),
			CpuShare: context.String("cpushare"),
		}
		logrus.Info("resConf: ", resConf.MemoryLimit)
		Run(tty, cmd, resConf)
		logrus.Info("run command action end")
		return nil
	},
}

var initCommand = cli.Command{
	Name: "init",
	Usage: `Init container process run user's process in container. Do not call it outside`,
	Action: func(ctx *cli.Context) error {
		logrus.Info("init command action start")
		cmd := ctx.Args().Get(0)
		logrus.Infof("init command %s", cmd)
		err := container.RunContainerInitProcess(cmd,nil)
		if err != nil {
			logrus.Errorf("init command actio err: %s", err)
			return err
		}
		return nil
	},
}