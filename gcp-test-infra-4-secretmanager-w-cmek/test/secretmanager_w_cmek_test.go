// Copyright 2021 Ricardo Cordeiro <ricardo.cordeiro@tux.com.pt>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package _test

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/googleapis/gax-go/v2"
	"github.com/gruntwork-io/terratest/modules/gcp"
	"github.com/gruntwork-io/terratest/modules/terraform"
	option "google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/grpc/codes"
)

// Define a default location for the Cloud KMS and Secret Manager resources.
const (
	defaultLocation = "global"
)

// testKmsResourceCreation tests the creation of KMS KeyRing and CryptoKey and returns the CryptoKey.
func testKmsResourceCreation(ctx context.Context, t *testing.T, projectID string,
	opts ...option.ClientOption) *kmspb.CryptoKey {
	t.Helper()

	// Create a connection with the Cloud KMS API
	kmsClient, err := kms.NewKeyManagementClient(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	// Close the connection once done
	defer kmsClient.Close()

	// To avoid "rpc error: code = Unauthenticated desc = transport: impersonate: status code 403:" errors, setup a Retryer
	// on Unauthenticated code returns to overcome the eventual consistent nature of setting up IAM policies on GCP
	// resources
	retryFunc := func() gax.Retryer {
		return gax.OnCodes([]codes.Code{codes.Unauthenticated}, gax.Backoff{
			Initial:    time.Second,
			Max:        time.Second * 30,
			Multiplier: 2,
		})
	}

	// Setup a deadline to 1 minutes after which the retries should stop and an error returned.
	ctxWithTimeout, cancel := context.WithDeadline(ctx, time.Now().Add(time.Minute*2))
	defer cancel()

	// Define the KMS KeyRing creation request
	keyRingID := gcp.RandomValidGcpName()
	keyRingReq := &kmspb.CreateKeyRingRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", projectID, defaultLocation),
		KeyRing: &kmspb.KeyRing{ //nolint:exhaustivestruct
			Name: keyRingID,
		},
		KeyRingId: keyRingID,
	}

	// Create the KMS KeyRing, retrying on Unauthenticated failures until deadline is reached
	keyRingResp, err := kmsClient.CreateKeyRing(ctxWithTimeout, keyRingReq, gax.WithRetry(retryFunc))
	if err != nil {
		t.Fatal(err)
	}

	// Define the KMS CryptoKey creation request
	cryptoKeyReq := &kmspb.CreateCryptoKeyRequest{
		Parent:      keyRingResp.Name,
		CryptoKeyId: gcp.RandomValidGcpName(),
		CryptoKey: &kmspb.CryptoKey{ //nolint:exhaustivestruct
			Purpose: kmspb.CryptoKey_ENCRYPT_DECRYPT,
		},
		SkipInitialVersionCreation: false,
	}

	// Create the KMS CryptoKey on the created KeyRing
	cryptoKeyResp, err := kmsClient.CreateCryptoKey(ctx, cryptoKeyReq)
	if err != nil {
		t.Fatal(err)
	}

	return cryptoKeyResp
}

// Test the creation of a Secret Manager Secret protected by the KMS CryptoKey.
func testProtectedSecretCreation(ctx context.Context, t *testing.T, projectID string, cryptoKey *kmspb.CryptoKey,
	opts ...option.ClientOption) {
	t.Helper()

	// Create a connection with the Secret Manager API
	smClient, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	// Close the connection once done
	defer smClient.Close()

	// Define the Secret Manager Secret creation request
	smSecretRequest := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", projectID),
		SecretId: gcp.RandomValidGcpName(),
		Secret: &secretmanagerpb.Secret{ //nolint:exhaustivestruct
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{
						CustomerManagedEncryption: &secretmanagerpb.CustomerManagedEncryption{
							KmsKeyName: cryptoKey.Name,
						},
					},
				},
			},
		},
	}

	// Create the Secret Manager Secret protected by the created KMS CryptoKey
	smSecret, err := smClient.CreateSecret(ctx, smSecretRequest)
	if err != nil {
		t.Fatal(err)
	}

	// Generate random data to encrypt in a new Secret Version
	data := make([]byte, 1)
	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}

	// Define the new Secret Version request
	smSecretVersionRequest := &secretmanagerpb.AddSecretVersionRequest{
		Parent: smSecret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: data,
		},
	}

	// Add a new SecretVersion to the Secret Manager Secret
	if _, err := smClient.AddSecretVersion(ctx, smSecretVersionRequest); err != nil {
		t.Fatal(err)
	}
}

// TestCMEKProtectedSecretManagerSecrets tests Terraform Google Provider Test infrastructure's ability to create CMEK
// encrypted Secret Manager Secrets
//
// The test procedure first creates the test infrastructure and then"
// 1. Creates a Cloud KMS KeyRing and CryptoKey with an initial CryptoKeyVersion;
// 2. Creates a Secret Manager Secret protected by the CryptoKey's encryption;.
func TestCMEKProtectedSecretManagerSecrets(t *testing.T) {
	ctx := context.Background()

	t.Parallel()

	// Set the terraform test directory and variables
	terraformOptions := &terraform.Options{ //nolint:exhaustivestruct
		TerraformDir: "..",
	}

	// When finished with the test, destroy the infrastructure
	defer terraform.Destroy(t, terraformOptions)

	// Build the test infrastructure
	terraform.InitAndPlan(t, terraformOptions)
	terraform.Apply(t, terraformOptions)

	// Get created project ID and service account name
	tfOutputs := make(map[string]string)
	for k, v := range terraform.OutputAll(t, terraformOptions) {
		tfOutputs[fmt.Sprintf("%v", k)] = fmt.Sprintf("%v", v)
	}

	// test the creation of KMS KeyRing and CryptoKey and returns the CryptoKey
	cryptoKeyResp := testKmsResourceCreation(ctx, t, tfOutputs["project_id"],
		option.ImpersonateCredentials(tfOutputs["service_account_email"]))

	// Test the creation of a Secret Manager Secret protected by the KMS CryptoKey
	testProtectedSecretCreation(ctx, t, tfOutputs["project_id"], cryptoKeyResp,
		option.ImpersonateCredentials(tfOutputs["service_account_email"]))

	// Steps to destroy the Cloud KMS and Secret Manager resources are skipped because the infrastructure destroy procedure
	// should delete the project that hosts these resources.
} //nolint:wsl
