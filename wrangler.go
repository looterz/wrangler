package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sns"
)

var steamcmdPath = "C:\\steamcmd"
var steamcmdBin = fmt.Sprintf("%s\\steamcmd.exe", steamcmdPath)
var serverBin = fmt.Sprintf("%s\\steamapps\\common\\The Isle Dedicated Server\\TheIsleServer.exe", steamcmdPath)
var ec2MetaClient *ec2metadata.EC2Metadata
var awsSession *session.Session
var ec2Service *ec2.EC2
var snsService *sns.SNS

var serverProcess *os.Process
var processName = "TheIsleServer.exe"
var serverBranch string
var serverName string
var serverMap string
var serverMaxPlayers string

func main() {
	// Setup logging
	logFile, err := os.OpenFile("wrangler.log", os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	// Setup AWS EC2 Metadata client for fetching information about this instance
	// session.New works without a region, session.NewSession requires a region?
	ec2MetaClient = ec2metadata.New(session.New())

	// Setup AWS session using this instances IAM role
	region, _ := regionInstanceID()
	awsSession, err = session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		log.Fatal("Unable to establish AWS Session", err)
	}

	// Setup the AWS client services
	ec2Service = ec2.New(awsSession)
	snsService = sns.New(awsSession)

	// Parse this EC2 instances tags
	serverBranch, err = getTagValue("Server_Branch")
	if err != nil {
		log.Panic(err)
	}

	serverName, err = getTagValue("Server_Name")
	if err != nil {
		log.Panic(err)
	}

	serverMap, err = getTagValue("Server_Map")
	if err != nil {
		log.Panic(err)
	}

	serverMaxPlayers, err = getTagValue("Server_MaxPlayers")
	if err != nil {
		log.Panic(err)
	}

	snsTopic, err := getTagValue("SNS_TOPIC")
	if err != nil {
		log.Panic(err)
	}

	// Setup SNS listener
	subscribe(snsTopic)

	http.HandleFunc("/", snsHandler)
	go http.ListenAndServe(":8081", nil)

	// Run steamcmd to update the server
	updateServer()
	startServer()

	// Setup the main loop
	programQuit = make(chan bool)
	processMonitorQuit = make(chan bool)

	go processMonitor()

	<-programQuit
}
