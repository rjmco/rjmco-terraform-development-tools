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

output "project_id" {
    description = "Created project's project_id."
    value = google_project.test.project_id
}

output "random_id" {
    description = "Generated random ID for resources."
    value = random_string.unique_id.result
}

output "service_account_email" {
    description = "Terraform Google Provider SA's email address"
    value = google_service_account.terraform_google_provider_test.email
}