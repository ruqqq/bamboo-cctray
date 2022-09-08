module github.com/ruqqq/bamboo-cctray

go 1.16

require (
	github.com/hashicorp/go-retryablehttp v0.7.0
	github.com/rcarmstrong/go-bamboo v0.0.0-20201005173404-72bad9513ccc
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/rcarmstrong/go-bamboo v0.0.0-20201005173404-72bad9513ccc => github.com/ruqqq/go-bamboo v0.0.0-20220908052732-0d6162818a15
