package main

//configuration
import (
	"bytes"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"context"
	"encoding/json"
	"fmt"
	isEmpty "github.com/24COMS/go.isempty"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
	"golang.org/x/oauth2/google"
	"hash/crc32"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Properties struct {
	Database            database
	AuthenticationKeys  map[string]string
	ConfigurationServer configuration_server
	Configurations      configurations
	Services            Services
	NATS                NATS
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

	Prop.AuthenticationKeys = map[string]string{}
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
	DomainSettings   map[string]Settings
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

type Settings struct {
	BaseSecretKey            string
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
		log.Error().Msgf("LoadConfigurationFromServer :%v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Msgf("LoadConfigurationFromServer :%v", err)
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

func (prop *Properties) GetSecretValue(domain string, key string) (string, error) {
	mapKey := domain + "-" + key
	secretValue := prop.AuthenticationKeys[mapKey]

	if _, ok := prop.Configurations.DomainSettings[domain]; !ok {
		return "", fmt.Errorf("DomainSettings is not configured for domain " + domain)
	}

	log.Debug().Msg("BaseSecretKey :" + prop.Configurations.DomainSettings[domain].BaseSecretKey)

	if isEmpty.Value(secretValue) {
		val, err := AccessSecret(prop.Configurations.DomainSettings[domain].BaseSecretKey, domain+"-"+key)
		if err != nil {
			return "", err
		}
		secretValue = val
		prop.AuthenticationKeys[mapKey] = secretValue
	}
	return secretValue, nil
}

func AccessSecret(baseKey, name string) (string, error) {
	secretValue := ""
	// Create the client.
	ctx := context.Background()

	SECRET_KEY := strings.Replace(baseKey, "{key}", name, -1)

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return secretValue, fmt.Errorf("failed to create secretmanager client: %w", err)
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: SECRET_KEY,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return secretValue, fmt.Errorf("failed to access secret version: %w", err)
	}

	// Verify the data checksum.
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		return secretValue, fmt.Errorf("Data corruption detected.")
	}

	secretValue = string(result.Payload.Data)
	return secretValue, nil
}

func main() {
	Initialize()
	initialize_projectid()
	secretVal, err := Prop.GetSecretValue("knowme", "authentication-key")
	if err != nil {
		log.Error().Msgf("GetSecretValue error : %v", err)
	}
	println("secretVal :" + secretVal)
}

func initialize_projectid() {
	ctx := context.Background()
	computeScope := "https://www.googleapis.com/auth/compute"
	credentials, err := google.FindDefaultCredentials(ctx, computeScope)
	if err != nil {
		fmt.Println(err)
	}

	println(credentials.ProjectID)
}

// APP_HOME=/Users/sudhakarduraiswamy/projects/git/GoLang/Configuration;GOOGLE_APPLICATION_CREDENTIALS=/Users/sudhakarduraiswamy/projects/git/GoLang/UserApi/.gcloud/imageapi_datastoreuser.json;PROFILE=dev
