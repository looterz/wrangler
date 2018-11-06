package main

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func regionInstanceID() (region string, instanceID string) {
	ec2InstanceIdentifyDocument, err := ec2MetaClient.GetInstanceIdentityDocument()
	if err != nil {
		log.Println(err)
	}

	region = ec2InstanceIdentifyDocument.Region
	instanceID = ec2InstanceIdentifyDocument.InstanceID

	return
}

func getPublicIPAddress() (address string, err error) {
	_, instanceID := regionInstanceID()

	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	}

	resp, err := ec2Service.DescribeInstances(params)
	if err != nil {
		log.Printf("%s", err)
		return "", err
	}
	if len(resp.Reservations) == 0 {
		return "", err
	}

	for idx := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			if *inst.InstanceId == instanceID {
				return *inst.PublicIpAddress, nil
			}
		}
	}

	return "", errors.New("Failed to find public ip address")
}

func getTagValue(name string) (value string, err error) {
	_, instanceID := regionInstanceID()

	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	}

	resp, err := ec2Service.DescribeInstances(params)
	if err != nil {
		log.Printf("%s", err)
		return "", err
	}
	if len(resp.Reservations) == 0 {
		return "", err
	}

	for idx := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			for _, tag := range inst.Tags {
				if (name != "") && (*tag.Key == name) {
					//log.Println(*tag.Key, "=", *tag.Value)
					return *tag.Value, nil
				}
			}
		}
	}

	return "", errors.New("Failed to find tag value")
}
