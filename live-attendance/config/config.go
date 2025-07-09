package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

const AccessControl string = "access-control"
const FTv2 string = "42-v2"
const FTAttendance string = "42-attendance"

var ConfigData ConfigFile

type ConfigWatchtime struct {
	Monday    [][]string `yaml:"monday"`
	Tuesday   [][]string `yaml:"thuesday"`
	Wednesday [][]string `yaml:"wednesday"`
	Thursday  [][]string `yaml:"thursday"`
	Friday    [][]string `yaml:"friday"`
	Saturday  [][]string `yaml:"saturday"`
	Sunday    [][]string `yaml:"sunday"`
}

type ConfigAccessControl struct {
	Endpoint string `yaml:"endpoint"`
	TestPath string `yaml:"testpath"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type ConfigAPIV2 struct {
	TokenUrl           string   `yaml:"tokenUrl"`
	Endpoint           string   `yaml:"endpoint"`
	TestPath           string   `yaml:"testpath"`
	Uid                string   `yaml:"uid"`
	Secret             string   `yaml:"secret"`
	Scope              string   `yaml:"scope"`
	CampusID           string   `yaml:"campusId"`
	ApprenticeProjects []string `yaml:"apprenticeProjects"`
}

type ConfigAttendance42 struct {
	AutoPost bool   `yaml:"autoPost"`
	TokenUrl string `yaml:"tokenUrl"`
	Endpoint string `yaml:"endpoint"`
	TestPath string `yaml:"testpath"`
	Uid      string `yaml:"uid"`
	Secret   string `yaml:"secret"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type ConfigMailer struct {
	SmtpServer string   `yaml:"smtp_server"`
	SmtpPort   int      `yaml:"smtp_port"`
	SmtpAuth   bool     `yaml:"smtp_auth"`
	SmtpUser   string   `yaml:"smtp_user"`
	SmtpPass   string   `yaml:"smtp_pass"`
	SmtpTLS    bool     `yaml:"smtp_tls"`
	Helo       string   `yaml:"helo"`
	FromName   string   `yaml:"from_name"`
	FromMail   string   `yaml:"from_mail"`
	Recipients []string `yaml:"recipients"`
}

type ConfigFile struct {
	AccessControl ConfigAccessControl `yaml:"AccessControl"`
	ApiV2         ConfigAPIV2         `yaml:"42apiV2"`
	Attendance42  ConfigAttendance42  `yaml:"42Attendance"`
	Mailer        ConfigMailer        `yaml:"mailer"`
	Watchtime     ConfigWatchtime     `yaml:"watchtime"`
}

func LoadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&ConfigData)
	if err != nil {
		return err
	}
	return nil
}
