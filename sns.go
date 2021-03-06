package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/service/sns"
)

func confirmSubscription(URL string) {
	_, err := http.Get(URL)
	if err != nil {
		fmt.Printf("Unable to confirm subscription")
	} else {
		fmt.Printf("Subscription confirmed")
	}
}

func snsHandler(w http.ResponseWriter, r *http.Request) {
	var f interface{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Unable to parse body")
		return
	}

	bodyStr := string(body)

	log.Printf(bodyStr)

	if bodyStr != "" {
		err = json.Unmarshal(body, &f)
		if err != nil {
			log.Printf("Unable to unmarshal request")
			return
		}

		data := f.(map[string]interface{})
		log.Println(data["Type"].(string))

		if data["Type"].(string) == "SubscriptionConfirmation" {
			URL := data["SubscribeURL"].(string)
			go confirmSubscription(URL)
		} else if data["Type"].(string) == "Notification" {
			log.Println("Received SNS Notification : ", data["Message"].(string))

			var payload = []byte(data["Message"].(string))
			var event map[string]interface{}

			err = json.Unmarshal(payload, &event)
			if err != nil {
				log.Println(err)
				return
			}

			// If the action is update, and we are on the target branch, trigger a server update
			if event["action"] == "update" && event["branch"] == serverBranch {
				log.Println("Received server update event for", serverBranch, "branch")

				processMonitorQuit <- true

				time.Sleep(time.Second * 5)

				if err := killServer(); err != nil {
					log.Println(err)
				}

				time.Sleep(time.Second * 5)

				updateServer(serverGameIniURI)
				startServer()

				go processMonitor()
			}
		}

		fmt.Fprintf(w, "Success")
	}
}

func subscribe(topicArn string) {
	publicIP, err := getPublicIPAddress()
	if err != nil {
		log.Panic(err)
	}

	protocol := "http"
	endpoint := fmt.Sprintf("%s://%s:8081/", protocol, publicIP)

	input := &sns.SubscribeInput{
		Endpoint: &endpoint,
		Protocol: &protocol,
		TopicArn: &topicArn,
	}

	out, err := snsService.Subscribe(input)
	if err != nil {
		fmt.Println("Unable to subscribe", err)
		return
	}

	log.Printf(*out.SubscriptionArn)
}
