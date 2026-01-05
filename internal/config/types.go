package config

type Config struct {
	Debug     bool   `yaml:"debug"`
	DataDir   string `yaml:"data_dir"`
	GitConfig Git    `yaml:"git"`
}

type Git struct {
	CreateDeployKey bool   `yaml:"create_deploy_key"`
	Repo            string `yaml:"repo"`
	Branch          string `yaml:"branch"`
	PollingInterval int    `yaml:"polling_interval"`
}
