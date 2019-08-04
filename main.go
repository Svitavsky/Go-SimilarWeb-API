package main

import (
	"encoding/csv"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Configuration struct {
	ApiKey         string   `yaml:"apiKey"`
	OutputFileName string   `yaml:"outputFilename"`
	NoDuplicates   bool     `yaml:"noDuplicates"`
	Websites       []string `yaml:"websites"`
	//DateFrom string `yaml:"date_from"`
	//DateTo   string `yaml:"date_to"`
}

type ApiData struct {
	Responses []Response
}

type Response struct {
	SimilarSites []SimilarSite `yaml:"similar_sites"`
}

type SimilarSite struct {
	Url   string `yaml:"url"`
	Score string `yaml:"score"`
}

const ConfigFileName = "config.yml"
const SimilarWebApiUrl = "https://api.similarweb.com/v1/website/%s/similar-sites/similarsites?api_key=%s"

var configuration Configuration

func main() {
	responses := request()

	if responses == nil {
		log.Fatal("API response is empty!")
	}

	generateFile(responses)
}

func request() (responses []*http.Response) {
	yamlFile, err := ioutil.ReadFile(ConfigFileName)
	if err != nil {
		log.Printf("No configuration file with name %s provided!", ConfigFileName)
	}

	err = yaml.Unmarshal(yamlFile, &configuration)
	if err != nil {
		log.Fatalf("Cannot read configuration file: %v", err)
	}

	key := configuration.ApiKey

	for _, website := range configuration.Websites {
		url := fmt.Sprintf(SimilarWebApiUrl, website, key)

		response, err := http.Get(url)
		if err != nil {
			log.Printf("API request error: %s", err)
			continue
		}

		responses = append(responses, response)
	}

	return
}

func generateFile(responses []*http.Response) {
	file, err := os.Create(configuration.OutputFileName)
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	var writeData [][]string
	for _, response := range responses {
		var data = Response{}

		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal("API request error: ", err)
		}

		err = yaml.Unmarshal(body, &data)
		if err != nil {
			log.Fatalf("Cannot parse API data: %v", err)
		}

		for _, similarSite := range data.SimilarSites {
			writeData = append(writeData, []string{similarSite.Url, similarSite.Score})
		}
	}

	if len(writeData) > 0 {
		file, err := os.Create(configuration.OutputFileName)
		if err != nil {
			log.Fatal("Cannot create file", err)
		}

		defer file.Close()

		w := csv.NewWriter(file)
		defer w.Flush()
		w.WriteAll(writeData)

		if err := w.Error(); err != nil {
			log.Fatalln("Error writing csv:", err)
		}
	} else {
		log.Fatal("No data for writing!")
	}
}
