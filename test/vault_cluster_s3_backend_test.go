package test

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

const VAULT_CLUSTER_S3_BACKEND_PATH = "examples/vault-s3-backend"

const VAR_S3_BUCKET_NAME = "s3_bucket_name"
const VAR_FORCE_DESTROY_S3_BUCKET = "force_destroy_s3_bucket"

// Test the Vault with S3 Backend example by:
//
// 1. Copy the code in this repo to a temp folder so tests on the Terraform code can run in parallel without the
//    state files overwriting each other.
// 2. Build the AMI in the vault-consul-ami example with the given build name
// 3. Deploy that AMI using the example Terraform code
// 4. SSH to a Vault node and initialize the Vault cluster
// 5. SSH to each Vault node and unseal it
// 6. Connect to the Vault cluster via the ELB
func runVaultWithS3BackendClusterTest(t *testing.T, amiId string, awsRegion, sshUserName string) {
	examplesDir := test_structure.CopyTerraformFolderToTemp(t, REPO_ROOT, VAULT_CLUSTER_S3_BACKEND_PATH)

	defer test_structure.RunTestStage(t, "teardown", func() {
		teardownResources(t, examplesDir)
	})

	defer test_structure.RunTestStage(t, "log", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, examplesDir)
		keyPair := test_structure.LoadEc2KeyPair(t, examplesDir)

		getVaultLogs(t, "vaultClusterWithS3Backend", terraformOptions, amiId, awsRegion, sshUserName, keyPair)
	})

	test_structure.RunTestStage(t, "deploy", func() {
		uniqueId := random.UniqueId()
		terraformVars := map[string]interface{}{
			VAR_S3_BUCKET_NAME:          s3BucketName(uniqueId),
			VAR_FORCE_DESTROY_S3_BUCKET: true,
			VAR_CONSUL_CLUSTER_NAME:     fmt.Sprintf("consul-test-%s", uniqueId),
			VAR_CONSUL_CLUSTER_TAG_KEY:  fmt.Sprintf("consul-test-%s", uniqueId),
		}
		deployCluster(t, amiId, awsRegion, examplesDir, uniqueId, terraformVars)
	})

	test_structure.RunTestStage(t, "validate", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, examplesDir)
		keyPair := test_structure.LoadEc2KeyPair(t, examplesDir)

		cluster := initializeAndUnsealVaultCluster(t, OUTPUT_VAULT_CLUSTER_ASG_NAME, sshUserName, terraformOptions, awsRegion, keyPair)
		testVaultUsesConsulForDns(t, cluster)
	})
}
