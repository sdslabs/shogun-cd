package config

type Config struct {
	Debug     bool   `yaml:"debug"`
	DataDir   string `yaml:"data_dir"`
	GitConfig Git    `yaml:"git"`
	ApiConfig Api    `yaml:"api"`
	DBConfig  DB     `yaml:"database"`
}

type Git struct {
	CreateDeployKey bool   `yaml:"create_deploy_key"`
	Repo            string `yaml:"repo"`
	Branch          string `yaml:"branch"`
	PollingInterval int    `yaml:"polling_interval"`
}

type Api struct {
	Admin      Admin `yaml:"admin"`
	Port       int   `yaml:"port"`
	JWT        JWT   `yaml:"jwt"`
	EnableSpec bool  `yaml:"enable_api_spec"`	
}

type Admin struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type JWT struct {
	Secret          string `yaml:"secret"`
	ExpirationHours int    `yaml:"expiration_hours"`
}

type DB struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	DBName         string `yaml:"dbname"`
	SSLMode        string `yaml:"ssl_mode"`
	MasterKey      string `yaml:"master_key"`
	EncryptionSalt string `yaml:"encryption_salt"`
	DerivedKey     []byte `yaml:"-"`
}
