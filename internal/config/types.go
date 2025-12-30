package config

type Config struct {
	Debug     bool `yaml:"debug"`
	GitConfig Git  `yaml:"git"`
}

type Git struct {
	CloneDir        string `yaml:"clone_dir"`
	CreateDeployKey bool   `yaml:"create_deploy_key"`
	KeyDir          string `yaml:"key_dir"`
	Repo            string `yaml:"repo"`
	Branch          string `yaml:"branch"`
	PollingInterval int    `yaml:"polling_interval"`
}
