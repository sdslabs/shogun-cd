package config

type Config struct {
	Debug     bool `yaml:"debug"`
	GitConfig Git  `yaml:"git"`
}

type Git struct {
	CloneDir        string `yaml:"clone_dir"`
	Repo            string `yaml:"repo"`
	Branch          string `yaml:"branch"`
	PollingInterval int    `yaml:"polling_interval"`
}
