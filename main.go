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

func init() {
	// setup location
	// NOTE: ignoring error from user.Current
	usr, _ := user.Current()
	configPath = path.Join(usr.HomeDir, ".saveConfig.json")

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			setupAddress()
		} else {
			log.Fatal(err)
		}
	}

	configFile, err := os.Open(configPath)
	fmt.Println("Configuration loaded from " + configPath)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func setupAddress() (configFile *os.File) {
	fmt.Println("To determine your location, we use Google Geocoding API for server. Please get your API Key at https://console.developers.google.com/apis/credentials/wizard?api=geocoding_backend")
	fmt.Print("Type your API Key: ")
	ApiKey := ""
	fmt.Scan(&ApiKey)

	fmt.Print("Type your location: ")
	address := ""
	fmt.Scan(&address)

	c, err := maps.NewClient(maps.WithAPIKey(ApiKey))
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

	location := locations[0]
	if len(locations) > 1 {
		for i, location := range locations {
			fmt.Fprintf(os.Stdout, "%d) %s\n", i+1, location.FormattedAddress)
		}

		choice := 0
		for choice < 1 || choice > len(locations) {
			fmt.Printf("\rWhich is your address? ")
			_, err := fmt.Scanf("%d", &choice)
			if err != nil {
				choice = 0
			}
		}

		location = locations[choice-1]
	}

	configuration := Configuration{location.FormattedAddress, location.Geometry.Location.Lat, location.Geometry.Location.Lng, ApiKey}

	// NOTE: ignoring error from json.Marshal
	data, _ := json.Marshal(configuration)

	err = ioutil.WriteFile(configPath, []byte(data), 0666)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Configuration saved to " + configPath)

	return
}

func main() {
	fmt.Printf("Api Key \t: %s\n", configuration.ApiKey)
	fmt.Printf("Address \t: %s\n", configuration.Address)
	fmt.Printf("Lat \t\t: %f\n", configuration.Lat)
	fmt.Printf("Lng \t\t: %f\n", configuration.Lng)
}
