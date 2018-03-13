package module

import (
	"encoding/json"
	"io/ioutil"
)

type List struct {
	Name    string   `json:"name"`
	Address []string `json:"address"`
}

type Lists map[string][]List

func (s Server) ImportLists(p string) error {
	var ls Lists
	raw, err := ioutil.ReadFile(p)
	err = json.Unmarshal(raw, &ls)

	for _, l := range ls["List"] {
		for _, a := range l.Address {
			s.SendCommand("/ip/firewall/address-list/add;=list=" + l.Name + ";=address=" + a)
		}
	}

	cls, err := s.GetCommand("/ip/firewall/address-list/print")
	for _, cl := range cls.Re {
		ls.synchronize(s, cl.Map["address"], cl.Map["list"], cl.Map[".id"])
	}

	return err
}

func (ls Lists) synchronize(s Server, ca string, cl string, id string) bool {
	for _, l := range ls["List"] {
		for _, a := range l.Address {
			if a == ca && l.Name == cl {
				return true
			}
		}
	}
	s.SendCommand("/ip/firewall/address-list/remove;=.id=" + id)
	return false
}
