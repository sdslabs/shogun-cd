package config

type Config struct {
	Debug     bool `yaml:"debug"`
	GitConfig Git  `yaml:"git"`
	ApiConfig Api  `yaml:"api"`
}

type Git struct {
	CloneDir        string `yaml:"clone_dir"`
	Repo            string `yaml:"repo"`
	Branch          string `yaml:"branch"`
	PollingInterval int    `yaml:"polling_interval"`
}

type Api struct {
	Admin Admin `yaml:"admin"`
	Port  int   `yaml:"port"`
	JWT   JWT   `yaml:"jwt"`
}

type Admin struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type JWT struct {
	Secret          string `yaml:"secret"`
	ExpirationHours int    `yaml:"expiration_hours"`
}
