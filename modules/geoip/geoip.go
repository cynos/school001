package geoip

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

type GeoIP struct {
	Ip          string  `json:"ip"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name""`
	RegionCode  string  `json:"region_code"`
	RegionName  string  `json:"region_name"`
	City        string  `json:"city"`
	Zipcode     string  `json:"zipcode"`
	TimeZone    string  `json:"time_zone"`
	Lat         float32 `json:"latitude"`
	Lon         float32 `json:"longitude"`
	MetroCode   int     `json:"metro_code"`
}

func Search(ip string) (GeoIP, error) {
	var instance GeoIP

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://freegeoip.app/json/%s", ip), nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return instance, err
	}
	defer res.Body.Close()

	// read the data in to a byte slice(string)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return instance, err
	}

	// Unmarshal the JSON byte slice to a GeoIP struct
	err = json.Unmarshal(body, &instance)
	if err != nil {
		return instance, err
	}

	return instance, nil
}

func GetIP(r *http.Request) (string, error) {
	//Get IP from the X-Real-Ip header
	ip := r.Header.Get("X-Real-Ip")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	//Get IP from X-Forwarded-For header
	ips := r.Header.Get("X-Forwarded-For")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}
	return "", fmt.Errorf("No valid ip found")
}
