package main

import (
	"encoding/xml"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var bambooConfigFilename = "bamboo.yml"
var CACHE_EXPIRY_DURATION time.Duration = time.Second * 30
var lastCacheTime = time.Time{}
var cacheValue []byte = nil

func cctrayHandler(w http.ResponseWriter, req *http.Request) {
	if cacheValue == nil || lastCacheTime.Add(CACHE_EXPIRY_DURATION).Before(time.Now()) {
		log.Println("Retrieving from Bamboo")
		result, err := getCcTrayProjects(bambooConfigFilename)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}

		lastCacheTime = time.Now()
		cacheValue = result
	} else {
		log.Printf("Using cached data. Last refreshed on: %s\n", lastCacheTime.Format(time.RFC3339))
	}

	w.Header().Add("Content-Type", "application/xml")
	w.WriteHeader(200)
	w.Write(cacheValue)
}

func main() {
	if len(os.Args) >= 2 {
		bambooConfigFilename = os.Args[1]
	}
	if len(os.Args) == 3 && os.Args[2] != "" {
		seconds, err := strconv.Atoi(os.Args[2])
		if err != nil {
			seconds = 30
		}
		CACHE_EXPIRY_DURATION = time.Second * time.Duration(seconds)
	}

	log.Printf("Bamboo Config File: %s\n", bambooConfigFilename)
	log.Printf("Cache expiry: %s\n", CACHE_EXPIRY_DURATION.String())

	http.HandleFunc("/dashboard/cctray.xml", cctrayHandler)

	log.Println("Starting server on :7000")
	http.ListenAndServe(":7000", nil)
}

func getCcTrayProject(client *BambooClient, buildKey string, output chan CcTrayProject) {
	result, err := client.getLatestResult(buildKey)
	if err != nil {
		log.Printf("error: %v", err)
	}

	if result == nil {
		output <- CcTrayProject{}
		return
	}

	output <- bambooToCcTrayProject(result)
}

func getCcTrayProjects(filename string) ([]byte, error) {
	ccTrayProjects := make([]CcTrayProject, 0)
	clients, err := createBambooClients(filename)
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		buildKeys, err := client.getBuildKeys()
		if err != nil {
			return nil, err
		}

		projectsChan := make(chan CcTrayProject, len(buildKeys))

		for _, buildKey := range buildKeys {
			go getCcTrayProject(client, buildKey, projectsChan)
		}

		for range buildKeys {
			ccTrayProject := <-projectsChan
			if ccTrayProject.Name != "" {
				ccTrayProjects = append(ccTrayProjects, ccTrayProject)
			}
		}
	}

	ccTray := CcTrayRoot{
		Projects: ccTrayProjects,
	}
	out, err := xml.MarshalIndent(ccTray, " ", "  ")
	if err != nil {
		return nil, err
	}

	return out, nil
}
