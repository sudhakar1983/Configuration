package ServerConfig

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
	"io/ioutil"
	"net/http"
)

type ConfigurationService struct {
	DomainAndBuckets map[string]string `json:"DomainAndBuckets"`
	DomainSettings   map[string]Settings
	ContactFormApi   ContactFormApi
	SuperChatApi     SuperChatApi
	UserApi          UserApi
	CheckSum         string
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
	BaseSecretKey                string
	GlobalKey                    string //change this to deactivate and force issuing new token
	TokenValidityInMins          int
	MessageExpiryInDays          int
	EmailNotificationEnabled     int
	PublishToIndexer             int
	EnableUsrTokenAuthentication int
}

type config_server_response struct {
	Response ConfigurationService `json:"Response"`
}

func LoadConfigurationFromServer(url string) (ConfigurationService, error) {
	var config = ConfigurationService{}
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

func IsValidDomain(domain string, validDomainsForTheService []string) bool {
	return slices.Contains(validDomainsForTheService, domain)
}

func GenerateCheckSum(configJsonStr string) string {
	sha1 := sha1.Sum([]byte(configJsonStr))
	return base64.StdEncoding.EncodeToString(sha1[:])
}
