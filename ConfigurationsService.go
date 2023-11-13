package Configuration

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
)

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
	BaseSecretKey                string
	GlobalKey                    string //change this to deactivate and force issuing new token
	TokenValidityInMins          int
	MessageExpiryInDays          int
	EmailNotificationEnabled     int
	PublishToIndexer             int
	EnableUsrTokenAuthentication int
}

type config_server_response struct {
	Response configurations `json:"Response"`
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
