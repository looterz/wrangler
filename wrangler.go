package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
)

type config struct {
	Steamcmd       string
	Server         string
	Process        string
	UseS3Bucket    bool
	S3Bucket       string
	S3BucketPrefix string
	S3Folder       string
	GameFolder     string
	AppID          string
	ServerConfig string
}

var Config config

var steamcmdBin string
var serverPath string
var serverBin string
var processName string

var serverProcess *os.Process

var ec2MetaClient *ec2metadata.EC2Metadata
var awsSession *session.Session
var ec2Service *ec2.EC2
var snsService *sns.SNS
var s3Service *s3.S3

var serverBranch string
var serverName string
var serverMap string
var serverMaxPlayers string
var serverGame string
var serverGameIniURI string

func main() {
	// Setup logging
	logFile, err := os.OpenFile("wrangler.log", os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	// Load toml configuration
	if _, err := toml.DecodeFile("wrangler.toml", &Config); err != nil {
		log.Panic(err)
	}

	steamcmdBin = Config.Steamcmd
	serverPath = Config.Server
	processName = Config.Process
	serverBin = fmt.Sprintf("%s\\%s", serverPath, processName)

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
	s3Service = s3.New(awsSession)

	// Parse this EC2 instances tags
	serverBranch, err = getTagValue("Server_Branch")
	if err != nil {
		log.Println("Failed to get tag value for Server_Branch")
		log.Panic(err)
	}

	serverName, err = getTagValue("Server_Name")
	if err != nil {
		log.Println("Failed to get tag value for Server_Name")
		log.Panic(err)
	}

	serverMap, err = getTagValue("Server_Map")
	if err != nil {
		log.Println("Failed to get tag value for Server_Map")
		log.Panic(err)
	}

	serverMaxPlayers, err = getTagValue("Server_MaxPlayers")
	if err != nil {
		log.Println("Failed to get tag value for Server_MaxPlayers")
		log.Panic(err)
	}

	serverGame, err = getTagValue("Server_Game")
	if err != nil {
		log.Println("Failed to get tag value for Server_Game")
		log.Panic(err)
	}

	snsTopic, err := getTagValue("SNS_TOPIC")
	if err != nil {
		log.Println("Failed to get tag value for SNS_TOPIC")
		log.Panic(err)
	}

	serverGameIniURI, err := getTagValue("GAME_INI_URI")
	if err != nil {
		log.Println("Custom game ini path undefined, using default")
	}

	// Setup SNS listener
	subscribe(snsTopic)

	http.HandleFunc("/", snsHandler)
	go http.ListenAndServe(":8081", nil)

	// Run steamcmd to update the server
	updateServer(serverGameIniURI)
	startServer()

	// Setup the main loop
	programQuit = make(chan bool)
	processMonitorQuit = make(chan bool)

	go processMonitor()

	<-programQuit
}
