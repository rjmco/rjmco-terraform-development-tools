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

resource "google_project" "test" {
  auto_create_network = false
  billing_account     = var.billing_account
  folder_id           = var.parent_folder_id
  labels              = var.default_resource_labels
  name                = format("%s-tf-dev-sandbox-%s", var.display_prefix, random_string.unique_id.result)
  project_id          = format("%s-tf-dev-sandbox-%s", var.display_prefix, random_string.unique_id.result)
}

resource "google_service_account" "terraform_google_provider_test" {
  account_id   = "terraform-google-provider-test"
  description  = "Service Account used to Terraform Google Provider development testing"
  display_name = "Terraform Google Provider Testing SA"
  project      = google_project.test.project_id
}

resource "google_project_iam_member" "editor_role_to_test_sa_on_project" {
  member  = format("serviceAccount:%s", google_service_account.terraform_google_provider_test.email)
  project = google_project.test.project_id
  role    = "roles/editor"
}