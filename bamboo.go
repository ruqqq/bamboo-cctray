package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	bamboo "github.com/rcarmstrong/go-bamboo"
	"gopkg.in/yaml.v2"
)

type BambooConfig struct {
	Url       string
	BasicAuth struct {
		Username string
		Password string
	} `yaml:"basic_auth"`
	BuildKeys []string `yaml:"build_keys"`
	Projects  []string
}

type BambooClient struct {
	Client *bamboo.Client
	Config BambooConfig
}

// PlanResponse encapsultes a response from the plan service
type PlanResponse struct {
	*bamboo.ResourceMetadata
	Plans *Plans `json:"plans"`
}

// Plans is a collection of Plan objects
type Plans struct {
	*bamboo.CollectionMetadata
	PlanList []*Plan `json:"plan"`
}

// Plan is the definition of a single plan
type Plan struct {
	ShortName string          `json:"shortName,omitempty"`
	ShortKey  string          `json:"shortKey,omitempty"`
	Type      string          `json:"type,omitempty"`
	Enabled   bool            `json:"enabled,omitempty"`
	Link      *bamboo.Link    `json:"link,omitempty"`
	Key       string          `json:"key,omitempty"`
	Name      string          `json:"name,omitempty"`
	PlanKey   *bamboo.PlanKey `json:"planKey,omitempty"`
	IsActive  bool            `json:"isActive,omitempty"`
}

// Result represents all the information associated with a build result
type Result struct {
	*bamboo.ResourceMetadata
	bamboo.ChangeSet       `json:"changes"`
	ID                     int    `json:"id"`
	PlanName               string `json:"planName"`
	ProjectName            string `json:"projectName"`
	BuildResultKey         string `json:"buildResultKey"`
	LifeCycleState         string `json:"lifeCycleState"`
	BuildStartedTime       string `json:"buildStartedTime"`
	BuildCompletedTime     string `json:"buildCompletedTime"`
	BuildDurationInSeconds int    `json:"buildDurationInSeconds"`
	VcsRevisionKey         string `json:"vcsRevisionKey"`
	BuildTestSummary       string `json:"buildTestSummary"`
	SuccessfulTestCount    int    `json:"successfulTestCount"`
	FailedTestCount        int    `json:"failedTestCount"`
	QuarantinedTestCount   int    `json:"quarantinedTestCount"`
	SkippedTestCount       int    `json:"skippedTestCount"`
	Finished               bool   `json:"finished"`
	Successful             bool   `json:"successful"`
	BuildReason            string `json:"buildReason"`
	ReasonSummary          string `json:"reasonSummary"`
	Key                    string `json:"key"`
	State                  string `json:"state"`
	BuildState             string `json:"buildState"`
	Number                 int    `json:"number"`
	BuildNumber            int    `json:"buildNumber"`
	Plan                   *Plan  `json:"plan"`
}

func (client *BambooClient) getBuildKeys() ([]string, error) {
	if len(client.Config.BuildKeys) > 0 {
		return client.Config.BuildKeys, nil
	}

	buildKeys := make([]string, 0)
	for _, project := range client.Config.Projects {
		plans, _, err := client.Client.Projects.ProjectPlans(project)
		if err != nil {
			return nil, err
		}

		for _, plan := range plans {
			buildKeys = append(buildKeys, plan.Key)
		}
	}

	return buildKeys, nil
}

func (client *BambooClient) getPlan(buildKey string) (*Plan, error) {
	result := Plan{}
	path := fmt.Sprintf("plan/%s", buildKey)
	request, err := client.Client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Client.Do(request, &result)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("GET %s returned statusCode %d", path, response.StatusCode))
	}

	return &result, nil
}

func (client *BambooClient) getLatestResult(buildKey string) (*Result, error) {
	result := Result{}
	path := fmt.Sprintf("result/%s/latest?expand=plan", buildKey)
	request, err := client.Client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Client.Do(request, &result)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("GET %s returned statusCode %d", path, response.StatusCode))
	}

	return &result, nil
}

func getBambooConfigs(filename string) (map[string]BambooConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	m := make([]map[string]BambooConfig, 0)
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]BambooConfig)
	for _, v := range m {
		for name, config := range v {
			ret[name] = config
		}
	}

	return ret, nil
}

func createBambooClients(configFilename string) (map[string]*BambooClient, error) {
	configs, err := getBambooConfigs(configFilename)
	if err != nil {
		return nil, err
	}

	clients := make(map[string]*BambooClient)
	for name, config := range configs {
		retryClient := retryablehttp.NewClient()
		retryClient.RetryMax = 2
		retryClient.RetryWaitMin = time.Millisecond * 100
		client := bamboo.NewSimpleClient(retryClient.StandardClient(), config.BasicAuth.Username, config.BasicAuth.Password)
		client.SetURL(config.Url)
		clients[name] = &BambooClient{
			Client: client,
			Config: config,
		}
	}

	return clients, nil
}
