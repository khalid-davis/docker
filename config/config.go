package config

var Config struct{
	RootURL string
}

func init() {
	Config.RootURL = "/root/docker-exp/busybox"
}
