package client

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

func TestFactoryBuildClients(t *testing.T) {
	factory := NewFactory()
	cfg := aws.Config{
		Region: "us-east-1",
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", "test")),
	}

	clients := factory.BuildClients(cfg)
	if clients == nil {
		t.Fatal("BuildClients() returned nil")
	}
	if clients.EC2 == nil {
		t.Error("BuildClients() EC2 client is nil")
	}
	if clients.RDS == nil {
		t.Error("BuildClients() RDS client is nil")
	}
	if clients.ELBv2 == nil {
		t.Error("BuildClients() ELBv2 client is nil")
	}
	if clients.STS == nil {
		t.Error("BuildClients() STS client is nil")
	}
}
