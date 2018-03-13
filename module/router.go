package module

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"gopkg.in/routeros.v2"
)

type Server struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Servers map[string]Server

func ImportConfig(p string) (Servers, error) {
	var s Servers
	raw, err := ioutil.ReadFile(p)
	err = json.Unmarshal(raw, &s)

	return s, err
}

func (s Server) dial() (*routeros.Client, error) {
	c, err := routeros.Dial(s.Address, s.Username, s.Password)
	if err != nil {
		return nil, err
	}
	c.Async()

	return c, nil
}

func (s Server) SendCommand(cmd string) error {
	c, err := s.dial()
	_, err = c.RunArgs(strings.Split(cmd, ";"))
	defer c.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s Server) GetCommand(cmd string) (*routeros.Reply, error) {
	c, err := s.dial()
	r, err := c.RunArgs(strings.Split(cmd, ";"))
	defer c.Close()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s Server) isRouting(cmt string) (*routeros.Reply, error) {
	c, err := s.dial()
	cmd := "/ip/route/print;?comment=" + cmt
	r, err := c.RunArgs(strings.Split(cmd, ";"))
	defer c.Close()
	if err != nil {
		return nil, err
	}

	return r, err
}

func (s Server) SetDefaultRoute(gw string) error {
	var cmd []string
	c, err := s.dial()
	gw = "=gateway=" + gw
	r, err := s.isRouting("default route")
	if r.Re != nil {
		cmd = strings.Split("/ip/route/set;=dst-address=0.0.0.0;=comment=default route;=.id="+r.Re[0].Map[".id"], ";")
	} else {
		cmd = strings.Split("/ip/route/add;=dst-address=0.0.0.0;=comment=default route", ";")
	}
	cmd = append(cmd, gw)
	_, err = c.RunArgs(cmd)
	defer c.Close()
	if err != nil {
		return err
	}

	return nil
}
