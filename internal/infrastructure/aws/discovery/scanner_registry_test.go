package discovery

import (
	"testing"

	"github.com/elip/WeaveLens/internal/infrastructure/aws/client"
)

func TestBuildRegionScanners_RegisteredScanners(t *testing.T) {
	scanners := BuildRegionScanners(&client.Clients{}, "us-east-1")
	wanted := map[string]bool{
		"Network": false, "NetworkAttachments": false, "EC2": false, "RDS": false,
		"ALB": false, "S3": false, "IAM": false, "Lambda": false, "SQS": false,
		"SNS": false, "ECR": false, "SecretsManager": false, "DynamoDB": false,
		"Elasticache": false, "CloudWatchLogs": false, "EventBridge": false,
		"CloudFront": false, "Route53": false, "TransitGateway": false,
		"APIGateway": false, "KMS": false, "StepFunctions": false,
	}
	for _, scanner := range scanners {
		name := scanner.Name()
		if _, ok := wanted[name]; !ok {
			t.Errorf("unexpected registered scanner %q", name)
			continue
		}
		if wanted[name] {
			t.Errorf("scanner %q registered more than once", name)
		}
		wanted[name] = true
	}
	if len(scanners) != len(wanted) {
		t.Errorf("BuildRegionScanners() returned %d scanners, want %d", len(scanners), len(wanted))
	}
	for name, found := range wanted {
		if !found {
			t.Errorf("scanner %q was not registered", name)
		}
	}
}

func TestBuildRegionScanners_NilClients(t *testing.T) {
	if scanners := BuildRegionScanners(nil, "us-east-1"); scanners != nil {
		t.Fatalf("BuildRegionScanners(nil) = %v, want nil", scanners)
	}
}
