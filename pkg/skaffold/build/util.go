/*
Copyright 2019 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package build

// MergeWithPreviousBuilds merges previous or prebuilt build artifacts with
// builds. If an artifact is already present in builds, the same artifact from
// previous will be ignored.
func MergeWithPreviousBuilds(builds, previous []Artifact) []Artifact {
	updatedBuilds := map[string]bool{}
	for _, build := range builds {
		updatedBuilds[build.ImageName] = true
	}

	var merged []Artifact
	merged = append(merged, builds...)

	for _, b := range previous {
		if !updatedBuilds[b.ImageName] {
			merged = append(merged, b)
		}
	}

	return merged
}
