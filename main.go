package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strconv"
	"time"
)

const iperfInterval = "1"
const iperfDuration = "2"

type IperfReport struct {
	Time      int64   `json:"time"`
	Bandwidth float64 `json:"average_bandwidth"`
}

type IperfReports struct {
	StartTime string                 `json:"Start Time"`
	Results   map[string]IperfReport `json:"Results"`
}

func (i *IperfReports) AddReport(route Route, report IperfReport, mode string) {
	i.Results[route.Name+"-"+mode] = report
}

var IperfMeasurementMode = [2]string{"Sending", "Receiving"}

type Route struct {
	Name string `json:"name"`
	Dev  string `json:"dev"`
	Ip   string `json:"ip"`
	Dns  string `json:"dns"`
}

func (r Route) MakeRouteDefault(dest Server) error {
	err := exec.Command("ip", "route", "replace", dest.Ip, "via", r.Ip, "dev", r.Dev).Run()
	if err != nil {
		return err
	}
	return nil
}

type Routes []Route

func ReadConfigFromFile(filepath string) (Routes, Servers, error) {
	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, nil, err
	}

	var configJson map[string]*json.RawMessage
	err = json.Unmarshal(raw, &configJson)
	if err != nil {
		return nil, nil, err
	}
	routesJson, ok := configJson["routes"]
	if !ok {
		return nil, nil, errors.New("There is no 'routes' path in json file " + filepath)
	}
	serversJson, ok := configJson["servers"]
	if !ok {
		return nil, nil, errors.New("There is no 'servers' path in json file " + filepath)
	}
	var routes Routes
	var servers Servers

	err = json.Unmarshal(*routesJson, &routes)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(*serversJson, &servers)
	if err != nil {
		return nil, nil, err
	}

	return routes, servers, nil
}

type Server struct {
	Username string `json:"username"`
	Ip       string `json:"ip"`
	Port     string `json:"port"`
}

func (s Server) CheckConnection(r Route) error {
	_, err := net.DialTimeout("tcp", r.Dns+":53", 5*time.Second)
	if err != nil {
		_, err2 := net.DialTimeout("udp", r.Dns+":53", 5*time.Second)
		if err2 != nil {
			return err2
		}
	}
	_, err = net.DialTimeout("tcp", s.Ip+":"+s.Port, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func MeasureBWIperf(route Route, server Server, mode string) (IperfReport, error) {
	var res []byte
	var err error
	var iperfReport IperfReport
	var iperfDataStruct map[string]interface{}
	var cmd *exec.Cmd
	cmd = exec.Command("timeout", "5", "iperf3", "-c", server.Ip, "-p", server.Port, "-i", iperfInterval, "-t", iperfDuration, "-J")
	if mode == "Receiving" {
		cmd = exec.Command("timeout", "5", "iperf3", "-c", server.Ip, "-p", server.Port, "-i", iperfInterval, "-t", iperfDuration, "-J", "-R")
	}
	log.Println(cmd.Args)
	res, err = cmd.Output()

	if err != nil {
		log.Println(err)
		return iperfReport, err
	}

	// parse iperf data
	err = json.Unmarshal(res, &iperfDataStruct)
	if err != nil {
		return iperfReport, err
	}

	if iperfDataStruct["error"] != nil {
		iperfDataError := fmt.Sprint(iperfDataStruct["error"])
		return iperfReport, errors.New(iperfDataError)
	}

	iperfDataTimeString := fmt.Sprint(iperfDataStruct["start"].(map[string]interface{})["timestamp"].(map[string]interface{})["timesecs"])
	iperfDataTimestamp, err := strconv.ParseInt(iperfDataTimeString, 10, 64)
	if err != nil {
		return iperfReport, errors.New("Cannot parse timestamp from Iperf Report data")
	}
	var iperfDataBWString string
	if mode == "Sending" {
		iperfDataBWString = fmt.Sprint(iperfDataStruct["end"].(map[string]interface{})["sum_sent"].(map[string]interface{})["bits_per_second"])
	} else if mode == "Receiving" {
		iperfDataBWString = fmt.Sprint(iperfDataStruct["end"].(map[string]interface{})["sum_received"].(map[string]interface{})["bits_per_second"])
	}
	iperfDataBW, err := strconv.ParseFloat(iperfDataBWString, 64)
	if err != nil {
		return iperfReport, errors.New("Cannot parse bandwidth from Iperf Report data")
	}

	iperfReport.Time = iperfDataTimestamp
	iperfReport.Bandwidth = iperfDataBW / (1024 * 1024)

	return iperfReport, nil
}

type Servers []Server

type ISPRuntime struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type ISPRuntimes map[string]ISPRuntime

func ReadRuntimeFromFile(filepath string) (ISPRuntimes, error) {
	rawFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var ispRuntimes ISPRuntimes
	err = json.Unmarshal(rawFile, &ispRuntimes)
	if err != nil {
		return nil, err
	}
	return ispRuntimes, nil
}

func (i *ISPRuntimes) UpdateRuntime(route Route, status string) {
	if (*i)[route.Name].Status == "up" && status == "down" {
		newStatus := ISPRuntime{Status: "down", Count: 0}
		(*i)[route.Name] = newStatus
	} else if (*i)[route.Name].Status == "down" {
		if status == "up" {
			newStatus := ISPRuntime{Status: "up", Count: 0}
			(*i)[route.Name] = newStatus
		} else if status == "down" {
			newStatus := ISPRuntime{Status: "down", Count: (*i)[route.Name].Count + 1}
			(*i)[route.Name] = newStatus
		}
	}
}

func (i ISPRuntimes) WriteToFile(filepath string) error {
	jsonfile, err := json.Marshal(i)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, jsonfile, 360)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	startTime := time.Now().Format(time.RFC3339)

	var iperfReports IperfReports
	iperfReports.StartTime = startTime

	configFile := "./config.json"
	runtimeFile := "./runtime.json"
	routes, servers, err := ReadConfigFromFile(configFile)
	if err != nil {
		log.Fatal("Cannot read routes or servers from file " + configFile + ". Error message: '" + fmt.Sprint(err) + "'")
	}
	runtimes, err := ReadRuntimeFromFile(runtimeFile)
	if err != nil {
		log.Fatal("Cannot read runtimes from file " + runtimeFile + ". Error message: '" + fmt.Sprint(err) + "'")
	}

	for _, server := range servers {
		log.Printf("Measuring bandwidth to server %s:%s\n", server.Ip, server.Port)
		for _, route := range routes {

			log.Println("Change default route to : ", route.Ip, " device ", route.Dev)

			err := route.MakeRouteDefault(server)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Check connection to server %s:%s\n", server.Ip, server.Port)
			err = server.CheckConnection(route)
			if err != nil {
				log.Printf("Connection to server %s:%s is DOWN\n", server.Ip, server.Port)
				runtimes.UpdateRuntime(route, "down")
			} else {
				log.Printf("Connection to server %s:%s is UP\n", server.Ip, server.Port)
				runtimes.UpdateRuntime(route, "up")
			}
			for _, mode := range IperfMeasurementMode {
				log.Printf("Measure bandwidth via route %s mode %s\n", route.Name, mode)
				res, err := MeasureBWIperf(route, server, mode)
				if err != nil {
					log.Fatal(err)
					iperfReports.Results[route.Name+"-"+mode] = res
				} else {
					log.Println("OK")
					log.Println(res)
					iperfReports.Results[route.Name+"-"+mode] = res
				}
			}
		}
		log.Println(iperfReports)
	}
}
