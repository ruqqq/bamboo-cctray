package main

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type CcTrayProject struct {
	LastBuildLabel  string `xml:"lastBuildLabel,attr"`
	LastBuildTime   string `xml:"lastBuildTime,attr"`
	Name            string `xml:"name,attr"`
	WebUrl          string `xml:"webUrl,attr"`
	Activity        string `xml:"activity,attr"`
	LastBuildStatus string `xml:"lastBuildStatus,attr"`
}

type CcTrayRoot struct {
	XMLName  xml.Name        `xml:"Projects"`
	Projects []CcTrayProject `xml:"Project"`
}

func bambooToCcTrayProject(result *Result) CcTrayProject {
	activity := "Sleeping"
	if result.Plan.IsActive {
		activity = "Building"
	}

	status := "unknown"
	if result.State == "Successful" {
		status = "Success"
	} else if result.State == "Failed" {
		status = "Failure"
	}

	lastBuildTime, err := time.Parse(time.RFC3339, result.BuildCompletedTime)
	if err != nil {
		lastBuildTime = time.Now()
	}

	webUrl := ""
	if result.ResourceMetadata != nil {
		webUrl = strings.Replace(result.ResourceMetadata.Link.HREF, "rest/api/latest/result", "browse", 1)
	}

	return CcTrayProject{
		LastBuildLabel:  fmt.Sprintf("%d", result.Number),
		LastBuildTime:   lastBuildTime.UTC().Format("2006-01-02T15:04:05Z"),
		Name:            result.Plan.ShortName,
		WebUrl:          webUrl,
		Activity:        activity,
		LastBuildStatus: status,
	}
}
