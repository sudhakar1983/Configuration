package main

//configuration
import (
	"bytes"
	"encoding/json"
	"fmt"
	isEmpty "github.com/24COMS/go.isempty"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Properties struct {
	Database            database
	Authenticator       authenticator
	ConfigurationServer configuration_server
	Configurations      configurations
	Services            Services
	NATS                NATS
}

type authenticator struct {
	Key string `yaml:"Key"`
}

type configuration_server struct {
	Url string `yaml:"Url"`
}

type Services struct {
	UserApi_FetchUser_Url string `yaml:"UserApi_FetchUser_Url"`
}

type database struct {
	Url                string `yaml:"Url"`
	Environment        string `yaml:"Environment"`
	ProjectId          string `yaml:"ProjectId"`
	AuthenticationFile string `yaml:"AuthenticationFile"`
}

type NATS struct {
	URL       string
	ClusterID string
	ClientID  string
}

var Prop *Properties

func Initialize() {
	println("init()")
	Prop := initializeConfiguration()
	println("Config server url :" + Prop.ConfigurationServer.Url)

	var err error
	Prop.Configurations, err = LoadConfigurationFromServer(Prop.ConfigurationServer.Url)
	if len(Prop.Configurations.DomainAndBuckets) < 1 {
		panic("Failed to load configurations from Configuration Server")
	}
	if err != nil {
		panic(err)
	}
}

func initializeConfiguration() *Properties {
	var app_home string = getEnv("APP_HOME", "")
	if isEmpty.Value(app_home) {
		panic("Missing env APP_HOME!!")
	}

	var profile string = getEnv("PROFILE", "")
	if isEmpty.Value(profile) {
		panic("Missing env PROFILE!!")
	}

	var configFilePath = app_home + "/envs/" + profile + "/config.yaml"
	println("Loading properties configuration from :" + configFilePath)
	data, err := os.ReadFile(configFilePath)
	viper.SetConfigType("yaml")
	err = viper.ReadConfig(bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("%v", err)
	}

	err = viper.Unmarshal(&Prop)
	if err != nil {
		fmt.Printf("%v", err)
	}
	return Prop
}

func IsProd() bool {
	var Result bool = false
	var profile string = getEnv("PROFILE", "")
	if "dev" != profile {
		fmt.Println("PROD mode")
		Result = true
	}
	return Result
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

type config_server_response struct {
	Response configurations `json:"Response"`
}

type configurations struct {
	DomainAndBuckets map[string]string `json:"DomainAndBuckets"`
	ContactFormApi   ContactFormApi
	SuperChatApi     SuperChatApi
	UserApi          UserApi
}

type ContactFormApi struct {
	ValidDomains             []string
	DomainAndServicesettings map[string]Servicesettings `json:"DomainAndServicesettings"`
}

type SuperChatApi struct {
	ValidDomains             []string
	DomainAndServicesettings map[string]Servicesettings `json:"DomainAndServicesettings"`
}

type UserApi struct {
	ValidDomains             []string
	DomainAndServicesettings map[string]Servicesettings `json:"DomainAndServicesettings"`
}

type Servicesettings struct {
	MessageExpiryInDays      int
	EmailNotificationEnabled int
	PublishToIndexer         int
}

func (prop *Properties) FetchServicesettings(domain string) Servicesettings {
	return prop.Configurations.ContactFormApi.DomainAndServicesettings[domain]
}

func (config *configurations) IsValidDomain(domain string) bool {
	return slices.Contains(config.ContactFormApi.ValidDomains, domain)
}

func LoadConfigurationFromServer(url string) (configurations, error) {
	var config = configurations{}
	var serverResponse = config_server_response{}
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return config, err
	}

	err = json.Unmarshal(body, &serverResponse)
	config = serverResponse.Response
	if err != nil {
		println(err)
		return config, err
	}
	return config, err
}

func GetStorageBucketName(domain string, properties *Properties) string {
	return properties.Configurations.DomainAndBuckets[domain]
}

func IsValidDomain(domain string, properties *Properties) bool {
	isValid := true
	if isEmpty.Value(GetStorageBucketName(domain, properties)) {
		isValid = false
	}
	return isValid
}
