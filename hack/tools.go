//go:build tools
// +build tools

/*
 * Copyright 2020 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tools

import (
	_ "knative.dev/eventing/hack"
	_ "knative.dev/hack"
	_ "knative.dev/hack/cmd/script" // Allow calling embedded scripts from hack
	_ "knative.dev/pkg/hack"

	// eventshub is a cloudevents sender/receiver utility for e2e testing.
	_ "knative.dev/reconciler-test/cmd/eventshub"

	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
