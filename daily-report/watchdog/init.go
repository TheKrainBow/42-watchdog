package watchdog

import (
	"fmt"
	"os"
	"watchdog/apiManager"
	"watchdog/config"
)

func init() {
	AllUsers = make(map[int]User)
}

func initAccessControlAPI() error {
	APIClient, err := apiManager.NewAPIClient(config.AccessControl, apiManager.APIClientInput{
		AuthType: apiManager.AuthTypeBasic,
		Username: config.ConfigData.AccessControl.Username,
		Password: config.ConfigData.AccessControl.Password,
		Endpoint: config.ConfigData.AccessControl.Endpoint,
		TestPath: config.ConfigData.AccessControl.TestPath,
	})
	if err != nil {
		return fmt.Errorf("couldn't create access control api client: %w", err)
	}
	err = APIClient.TestConnection()
	if err != nil {
		return fmt.Errorf("api connection test to access control failed: %w", err)
	}
	return nil
}

func init42AttendanceAPI() error {
	APIClient, err := apiManager.NewAPIClient(config.FTAttendance, apiManager.APIClientInput{
		AuthType:     apiManager.AuthTypePassword,
		TokenURL:     config.ConfigData.Attendance42.TokenUrl,
		Endpoint:     config.ConfigData.Attendance42.Endpoint,
		TestPath:     config.ConfigData.Attendance42.TestPath,
		ClientID:     config.ConfigData.Attendance42.Uid,
		ClientSecret: config.ConfigData.Attendance42.Secret,
		Username:     config.ConfigData.Attendance42.Username,
		Password:     config.ConfigData.Attendance42.Password,
	})
	if err != nil {
		return fmt.Errorf("couldn't create attendance api client: %w", err)
	}
	err = APIClient.TestConnection()
	if err != nil {
		return fmt.Errorf("api connection test to attendance failed: %w", err)
	}
	return nil
}

func init42v2API() error {
	APIClient, err := apiManager.NewAPIClient(config.FTv2, apiManager.APIClientInput{
		AuthType:     apiManager.AuthTypeClientCredentials,
		TokenURL:     config.ConfigData.ApiV2.TokenUrl,
		Endpoint:     config.ConfigData.ApiV2.Endpoint,
		TestPath:     config.ConfigData.ApiV2.TestPath,
		ClientID:     config.ConfigData.ApiV2.Uid,
		ClientSecret: config.ConfigData.ApiV2.Secret,
		Scope:        config.ConfigData.ApiV2.Scope,
	})
	if err != nil {
		return fmt.Errorf("couldn't create 42v2 api client: %w", err)
	}
	err = APIClient.TestConnection()
	if err != nil {
		return fmt.Errorf("api connection test to 42v2 failed: %w", err)
	}
	return nil
}

func Init(logPath string) error {
	err := InitLogs(logPath)
	if err != nil {
		fmt.Printf("ERROR: couldn't init logs")
		os.Exit(1)
	}
	err = initAccessControlAPI()
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s\n", err.Error()))
		os.Exit(1)
	}
	err = init42v2API()
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s\n", err.Error()))
		os.Exit(1)
	}
	err = init42AttendanceAPI()
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s\n", err.Error()))
		os.Exit(1)
	}
	return nil
}
