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

resource "random_string" "unique_id" {
  length    = 4
  min_lower = 4
}

resource "google_folder" "base" {
  display_name = format("%s TF Dev Sandbox %s", var.display_prefix, random_string.unique_id.result)
  parent       = format("folders/%s", var.parent_folder_id)
}

resource "google_folder_organization_policy" "skip_default_network_creation" {
  constraint = "constraints/compute.skipDefaultNetworkCreation"
  boolean_policy {
    enforced = true
  }
  folder = google_folder.base.name
}

resource "time_sleep" "wait_after_org_policy" {
  depends_on      = [google_folder_organization_policy.skip_default_network_creation]
  create_duration = "15s"
}

resource "google_project" "test" {
  auto_create_network = false
  billing_account     = var.billing_account
  folder_id           = basename(google_folder.base.name)
  labels              = var.default_resource_labels
  name                = format("%s-tf-dev-sandbox-%s", var.display_prefix, random_string.unique_id.result)
  project_id          = format("%s-tf-dev-sandbox-%s", var.display_prefix, random_string.unique_id.result)
  depends_on          = [time_sleep.wait_after_org_policy]
}

module "goole_project_services" {
  source  = "terraform-google-modules/project-factory/google//modules/project_services"
  version = "10.3.1"

  activate_api_identities = [
    {
      api = "secretmanager.googleapis.com"
      roles = [
        "roles/cloudkms.cryptoKeyEncrypterDecrypter",
      ]
    },
  ]

  activate_apis = [
    "cloudkms.googleapis.com",
    "secretmanager.googleapis.com",
  ]
  disable_services_on_destroy = false
  enable_apis                 = true
  project_id                  = google_project.test.project_id
}

resource "google_service_account" "terraform_google_provider_test" {
  account_id   = "terraform-google-provider-test"
  description  = "Service Account used to Terraform Google Provider development testing"
  display_name = "Terraform Google Provider Testing SA"
  project      = google_project.test.project_id
}

resource "google_service_account_iam_member" "access_token_creator_on_tf_goog_provider_test_sa_on_client_id" {
  member             = length(regexall("gserviceaccount.com", data.google_client_openid_userinfo.me.email)) > 0 ? format("serviceAccount:%s", data.google_client_openid_userinfo.me.email) : format("user:%s", data.google_client_openid_userinfo.me.email)
  service_account_id = google_service_account.terraform_google_provider_test.name
  role               = "roles/iam.serviceAccountTokenCreator"
}

resource "google_project_iam_member" "editor_role_to_test_sa_on_project" {
  member  = format("serviceAccount:%s", google_service_account.terraform_google_provider_test.email)
  project = google_project.test.project_id
  role    = "roles/editor"
}