// Copyright 2021 Authors of Cilium
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

package quay

import (
	"fmt"
	"time"
)

type Time struct {
	time.Time
}

func (d *Time) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	ot, err := time.Parse(fmt.Sprintf(`"%s"`, time.RFC1123Z), string(b))
	if err != nil {
		return err
	}
	d.Time = ot
	return nil
}

type Tag struct {
	Name           string `json:"name"`
	Reversion      bool   `json:"reversion"`
	EndTS          uint64 `json:"end_ts"`
	StartTS        uint64 `json:"start_ts"`
	ImageID        string `json:"image_id"`
	LastModified   Time   `json:"last_modified"`
	Expiration     Time   `json:"expiration"`
	ManifestDigest string `json:"manifest_digest"`
	DockerImageID  string `json:"docker_image_id"`
	IsManifestList bool   `json:"is_manifest_list"`
	Size           uint64 `json:"size"`
}

type TagsGet struct {
	HasAdditional bool  `json:"has_additional"`
	Page          int   `json:"page"`
	Tags          []Tag `json:"tags"`
}
