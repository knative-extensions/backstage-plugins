/*
Copyright 2020 The Knative Authors

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

package feature

import (
	"fmt"
)

type Timing uint8

const (
	// Prerequisite timing allows having steps that are asserting whether the feature should run
	// or not, a failed Prerequisite step will cause the other steps to be skipped.
	Prerequisite Timing = iota
	Setup
	Requirement
	Assert
	Teardown
)

func (t Timing) String() string {
	return timingMapping[t]
}

func (t Timing) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.String())), nil
}

func Timings() []Timing {
	return []Timing{Prerequisite, Setup, Requirement, Assert, Teardown}
}

var timingMapping = map[Timing]string{
	Prerequisite: "Prerequisite",
	Setup:        "Setup",
	Requirement:  "Requirement",
	Assert:       "Assert",
	Teardown:     "Teardown",
}
