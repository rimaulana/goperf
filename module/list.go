package module

import (
	"encoding/json"
	"net/http"
	"time"
)

type List struct {
	Name    string   `json:"name"`
	Address []string `json:"address"`
}

type Lists map[string][]List

type Prefix struct {
	Ip_prefix string `json:"ip_prefix"`
	Region    string `json:"region"`
	Service   string `json:"service"`
}

type Prefixes map[string][]Prefix
type IPs []string

func getAwsIpRange(url string) (IPs, error) {
	var ips IPs
	var nc = &http.Client{
		Timeout: time.Second * 10,
	}

	r, err := nc.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var pfxs Prefixes
	json.NewDecoder(r.Body).Decode(&pfxs)
	for _, ip := range pfxs["prefixes"] {
		if ip.Region == "ap-southeast-1" {
			ips = append(ips, ip.Ip_prefix)
		}
	}

	return ips, nil
}

func (s Server) ImportLists() error {
	url := "https://ip-ranges.amazonaws.com/ip-ranges.json"
	ips, err := getAwsIpRange(url)
	for _, ip := range ips {
		s.SendCommand("/ip/firewall/address-list/add;=list=AWS_list;=address=" + ip)
	}

	cls, err := s.GetCommand("/ip/firewall/address-list/print")
	for _, cl := range cls.Re {
		ips.synchronize(s, cl.Map["address"], cl.Map["list"], cl.Map[".id"])
	}

	return err
}

func (ips IPs) synchronize(s Server, cip string, cl string, id string) bool {
	for _, ip := range ips {
		if ip == cip && cl == "AWS_list" {
			return true
		}
	}
	s.SendCommand("/ip/firewall/address-list/remove;=.id=" + id)
	return false
}
