package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"

	"golang.org/x/net/context"

	"googlemaps.github.io/maps"
)

type Configuration struct {
	Address string
	Lat     float64
	Lng     float64
	ApiKey  string `json:"api_key"`
}

var (
	configPath    string
	configuration Configuration
)

func setupAddress() {
	fmt.Println("To determine your location, we use Google Geocoding API for server. Please get your API Key at https://console.developers.google.com/apis/credentials/wizard?api=geocoding_backend")
	fmt.Print("Type your API Key: ")
	var apiKey string
	fmt.Scan(&apiKey)

	fmt.Print("Type your location: ")
	var address string
	fmt.Scan(&address)

	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}

	r := &maps.GeocodingRequest{
		Address: address,
	}

	locations, err := c.Geocode(context.Background(), r)
	if err != nil {
		log.Fatal(err)
	}

	if len(locations) == 0 {
		log.Fatal("no locations found")
	}

	location := locations[0]
	if len(locations) > 1 {
		for i, location := range locations {
			fmt.Printf("%d) %s\n", i+1, location.FormattedAddress)
		}

		choice := 0
		for choice < 1 || choice > len(locations) {
			fmt.Printf("\rWhich is your address? ")
			_, err := fmt.Scanf("%d", &choice)
			if err != nil {
				fmt.Printf("Please type %d to %d", 1, len(locations))
				choice = 0
			}
		}

		location = locations[choice-1]
	}

	configuration := Configuration{
		Address: location.FormattedAddress,
		Lat:     location.Geometry.Location.Lat,
		Lng:     location.Geometry.Location.Lng,
		ApiKey:  apiKey}

	data, err := json.Marshal(configuration)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(configPath, data, 0660)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Configuration saved to " + configPath)
}

func main() {
	// setup location
	// NOTE: ignoring error from user.Current
	usr, _ := user.Current()
	configPath = path.Join(usr.HomeDir, ".saveConfig.json")

	if _, err := os.Stat(configPath); err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
		setupAddress()
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Configuration loaded from " + configPath)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Api Key \t: %s\n", configuration.ApiKey)
	fmt.Printf("Address \t: %s\n", configuration.Address)
	fmt.Printf("Lat \t\t: %f\n", configuration.Lat)
	fmt.Printf("Lng \t\t: %f\n", configuration.Lng)
}
