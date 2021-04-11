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

variable "billing_account" {
  type        = string
  description = "The alphanumeric ID of the billing account the project belongs to."
}

variable "default_resource_labels" {
  type        = map(any)
  description = "A set of key/value label pairs to assign to the project."
}

variable "display_prefix" {
  type        = string
  description = "String prefix to append to each resource name and displayed name."
}

variable "parent_folder_id" {
  type        = string
  description = "The numeric ID of the folder this project should be created under."
}