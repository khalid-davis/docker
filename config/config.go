package config

var Config struct{
	RootURL string
	MntURL string
}

func init() {
	Config.RootURL = "/root/docker-exp/"
	Config.MntURL = "/root/docker-exp/mnt/"
}
