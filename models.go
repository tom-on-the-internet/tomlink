package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

// Redirect is a user created redirect.
type Redirect struct {
	id         int
	Link       string
	URL        string
	AccessCode string
	CreatedAt  time.Time
	Visits     []Visit
}

func (r *Redirect) SafeFullLink(host string) template.URL {
	return template.URL(host + "/" + r.Link)
}

func (r Redirect) VisitCount() int {
	return len(r.Visits)
}

// Visit is the record of a client using the redirect.
type Visit struct {
	IPAddress  string
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	ISP        string `json:"isp"`
	CreatedAt  time.Time
}

type ViewData struct {
	Name       string
	Redirect   Redirect
	CurrentURL template.URL
	LinkURL    template.URL
	Host       string
}

func createVisitFromIPAddress(ipAddress string) (Visit, error) {
	// locally always local host
	if os.Getenv("mock_ip") != "" {
		ipAddress = os.Getenv("mock_ip")
	}

	visit := Visit{IPAddress: ipAddress}

	r, err := http.Get("http://ip-api.com/json/" + ipAddress)
	if err != nil {
		log.Println(err)

		return visit, err
	}

	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&visit)
	if err != nil {
		log.Println(err)

		return visit, err
	}

	return visit, nil
}

func newViewData(name string) ViewData {
	host := os.Getenv("host")
	if host == "" {
		log.Println("Failed to fetch host")
	}

	return ViewData{Host: host, Name: name}
}
