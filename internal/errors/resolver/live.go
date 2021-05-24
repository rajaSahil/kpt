// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resolver

import (
	"github.com/GoogleContainerTools/kpt/internal/cmdliveinit"
	"github.com/GoogleContainerTools/kpt/internal/cmdutil"
	"github.com/GoogleContainerTools/kpt/internal/errors"
	"sigs.k8s.io/cli-utils/pkg/apply/taskrunner"
	"sigs.k8s.io/cli-utils/pkg/inventory"
)

//nolint:gochecknoinits
func init() {
	AddErrorResolver(&liveErrorResolver{})
}

const (
	noInventoryObjErrorMsg = `
Error: Package uninitialized. Please run "kpt live init" command.

The package needs to be initialized to generate the template
which will store state for resource sets. This state is
necessary to perform functionality such as deleting an entire
package or automatically deleting omitted resources (pruning).
`
	multipleInventoryObjErrorMsg = `
Error: Package has multiple inventory object templates.

The package should have one and only one inventory object template.
`
	//nolint:lll
	timeoutErrorMsg = `
Error: Timeout after {{printf "%.0f" .err.Timeout.Seconds}} seconds waiting for {{printf "%d" (len .err.TimedOutResources)}} out of {{printf "%d" (len .err.Identifiers)}} resources to reach condition {{ .err.Condition}}:{{ printf "\n" }}

{{- range .err.TimedOutResources}}
{{printf "%s/%s %s %s" .Identifier.GroupKind.Kind .Identifier.Name .Status .Message }}
{{- end}}
`

	resourceGroupCRDInstallErrorMsg = `
Error: Unable to install the ResourceGroup CRD.

{{- if gt (len .cause) 0 }}
{{ printf "\nDetails:" }}
{{ printf "%s" .cause }}
{{- end }}
`
	//nolint:lll
	noResourceGroupCRDMsg = `
Error: The ResourceGroup CRD was not found in the cluster. Please install it either by using the '--install-resource-group' flag or the 'kpt live install-resource-group' command.
`

	//nolint:lll
	invInfoAlreadyExistsMsg = `
Error: Inventory information has already been added to the package Kptfile. Changing it after a package has been applied to the cluster can lead to undesired results. Use the --force flag to suppress this error.
`

	TimeoutErrorExitCode = 3
)

// liveErrorResolver is an implementation of the ErrorResolver interface
// that can resolve error types used in the live functionality.
type liveErrorResolver struct{}

func (*liveErrorResolver) Resolve(err error) (ResolvedResult, bool) {
	tmplArgs := map[string]interface{}{
		"err": err,
	}

	var noInventoryObjError *inventory.NoInventoryObjError
	if errors.As(err, &noInventoryObjError) {
		return ResolvedResult{
			Message: ExecuteTemplate(noInventoryObjErrorMsg, tmplArgs),
		}, true
	}

	var multipleInventoryObjError *inventory.MultipleInventoryObjError
	if errors.As(err, &multipleInventoryObjError) {
		return ResolvedResult{
			Message: ExecuteTemplate(multipleInventoryObjErrorMsg, tmplArgs),
		}, true
	}

	var timeoutError *taskrunner.TimeoutError
	if errors.As(err, &timeoutError) {
		return ResolvedResult{
			Message:  ExecuteTemplate(timeoutErrorMsg, tmplArgs),
			ExitCode: TimeoutErrorExitCode,
		}, true
	}

	var resourceGroupCRDInstallError *cmdutil.ResourceGroupCRDInstallError
	if errors.As(err, &resourceGroupCRDInstallError) {
		return ResolvedResult{
			Message: ExecuteTemplate(resourceGroupCRDInstallErrorMsg, map[string]interface{}{
				"cause": resourceGroupCRDInstallError.Err.Error(),
			}),
		}, true
	}

	var noResourceGroupCRDError *cmdutil.NoResourceGroupCRDError
	if errors.As(err, &noResourceGroupCRDError) {
		return ResolvedResult{
			Message: ExecuteTemplate(noResourceGroupCRDMsg, tmplArgs),
		}, true
	}

	var invExistsError *cmdliveinit.InvExistsError
	if errors.As(err, &invExistsError) {
		return ResolvedResult{
			Message: ExecuteTemplate(invInfoAlreadyExistsMsg, tmplArgs),
		}, true
	}

	return ResolvedResult{}, false
}
