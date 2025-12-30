package config

type Config struct {
	Debug     bool `yaml:"debug"`
	GitConfig Git  `yaml:"git"`
}

type Git struct {
	CreateDeployKey bool   `yaml:"create_deploy_key"`
	KeyDir          string `yaml:"key_dir"`
	Repo            string `yaml:"repo"`
	Branch          string `yaml:"branch"`
	CloneDir        string `yaml:"clone_dir"`
	PollingInterval int    `yaml:"polling_interval"`
}
