package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	errors      appError
	dataChannel chan int
)

type appError struct {
	Err   error
	Stack []error
}

func (a *appError) Add(err error) {
	a.Err = err
	a.Stack = append(a.Stack, err)
}

func (a *appError) Error() string {
	return a.Err.Error()
}

type Service struct {
	Id   int
	Name string
}

type ServiceDetails struct {
	Id     int
	Name   string
	Active bool
}

func main() {
	var active_services map[string]ServiceDetails
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://catalog.compamy.com/api/v1/services", nil)
	resp, _ := client.Do(req)
	buf, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var services []Service
	err := json.Unmarshal(buf, &services)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	for _, svc := range services {
		go func() {
			resp, err := http.Get(fmt.Sprintf("http://catalog.compamy.com/api/v1/service/%d", svc.Id))
			if err != nil {
				errors.Add(err)
			} else {
				data, _ := ioutil.ReadAll(resp.Body)
				var details ServiceDetails
				if err := json.Unmarshal(data, &details); err != nil {
					errors.Add(err)
					return
				}

				if details.Active {
					active_services[details.Name] = details
				}
			}
		}()
	}

	if errors.Err != nil {
		fmt.Println("Error:", errors)
		os.Exit(1)
	}

	for _, svc := range active_services {
		fmt.Printf("Service %s is active\n", svc.Name)
	}
}
