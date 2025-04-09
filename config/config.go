package config

import (
	"os"
	"watchdog/apiManager"

	"gopkg.in/yaml.v2"
)

const AccessControl string = "access-control"
const FTv2 string = "42-v2"
const FTAttendance string = "42-attendance"

var ConfigData configFile

type configFile struct {
	AccessControl struct {
		Endpoint string `yaml:"endpoint"`
		TestPath string `yaml:"testpath"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"AccessControl"`
	ApiV2 struct {
		TokenUrl           string   `yaml:"tokenUrl"`
		Endpoint           string   `yaml:"endpoint"`
		TestPath           string   `yaml:"testpath"`
		Uid                string   `yaml:"uid"`
		Secret             string   `yaml:"secret"`
		Scope              string   `yaml:"scope"`
		CampusID           string   `yaml:"campusId"`
		ApprenticeProjects []string `yaml:"apprenticeProjects"`
	} `yaml:"42apiV2"`
	Attendance42 struct {
		AutoPost bool   `yaml:"autoPost"`
		TokenUrl string `yaml:"tokenUrl"`
		Endpoint string `yaml:"endpoint"`
		TestPath string `yaml:"testpath"`
		Uid      string `yaml:"uid"`
		Secret   string `yaml:"secret"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"42Attendance"`
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

	_, err = apiManager.NewAPIClient(FTv2, apiManager.APIClientInput{
		AuthType:     apiManager.AuthTypeClientCredentials,
		TokenURL:     ConfigData.ApiV2.TokenUrl,
		Endpoint:     ConfigData.ApiV2.Endpoint,
		TestPath:     ConfigData.ApiV2.TestPath,
		ClientID:     ConfigData.ApiV2.Uid,
		ClientSecret: ConfigData.ApiV2.Secret,
		Scope:        ConfigData.ApiV2.Scope,
	})
	if err != nil {
		return err
	}

	_, err = apiManager.NewAPIClient(FTAttendance, apiManager.APIClientInput{
		AuthType:     apiManager.AuthTypePassword,
		TokenURL:     ConfigData.Attendance42.TokenUrl,
		Endpoint:     ConfigData.Attendance42.Endpoint,
		TestPath:     ConfigData.Attendance42.TestPath,
		ClientID:     ConfigData.Attendance42.Uid,
		ClientSecret: ConfigData.Attendance42.Secret,
		Username:     ConfigData.Attendance42.Username,
		Password:     ConfigData.Attendance42.Password,
	})
	if err != nil {
		return err
	}
	return nil
}
