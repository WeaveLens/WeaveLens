package client

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Clients struct {
	APIGateway     APIGatewayAPI
	CloudFront     CloudFrontAPI
	EC2            EC2API
	EC2Attachments EC2AttachmentAPI
	RDS            RDSAPI
	ELBv2          ELBv2API
	S3             S3API
	IAM            IAMAPI
	KMS            KMSAPI
	Lambda         LambdaAPI
	SQS            SQSAPI
	SNS            SNSAPI
	STS            STSAPI
	ECR            ECRAPI
	SecretsMgr     SecretsManagerAPI
	DynamoDB       DynamoDBAPI
	Elasticache    ElasticacheAPI
	CloudWatchLogs CloudWatchLogsAPI
	EventBridge    EventBridgeAPI
	Route53        Route53API
	StepFunctions  StepFunctionsAPI
	TransitGateway TransitGatewayAPI
}

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) BuildClients(cfg aws.Config) *Clients {
	return &Clients{
		APIGateway:     apigateway.NewFromConfig(cfg),
		CloudFront:     cloudfront.NewFromConfig(cfg),
		EC2:            ec2.NewFromConfig(cfg),
		EC2Attachments: ec2.NewFromConfig(cfg),
		RDS:            rds.NewFromConfig(cfg),
		ELBv2:          elasticloadbalancingv2.NewFromConfig(cfg),
		S3:             s3.NewFromConfig(cfg),
		IAM:            iam.NewFromConfig(cfg),
		KMS:            kms.NewFromConfig(cfg),
		Lambda:         lambda.NewFromConfig(cfg),
		SQS:            sqs.NewFromConfig(cfg),
		SNS:            sns.NewFromConfig(cfg),
		STS:            sts.NewFromConfig(cfg),
		ECR:            ecr.NewFromConfig(cfg),
		SecretsMgr:     secretsmanager.NewFromConfig(cfg),
		DynamoDB:       dynamodb.NewFromConfig(cfg),
		Elasticache:    elasticache.NewFromConfig(cfg),
		CloudWatchLogs: cloudwatchlogs.NewFromConfig(cfg),
		EventBridge:    eventbridge.NewFromConfig(cfg),
		Route53:        route53.NewFromConfig(cfg),
		StepFunctions:  sfn.NewFromConfig(cfg),
		TransitGateway: ec2.NewFromConfig(cfg),
	}
}
