//go:build generate
// +build generate

//go:generate rm -rf ../manifests/crd/bases
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.1 object:headerFile="../hack/copyright.go.txt" paths="./..."
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.1 crd:trivialVersions=true,preserveUnknownFields=false rbac:roleName=manager-role output:crd:artifacts:config=../manifests/crd/bases paths="./..."
//go:generate cp ../manifests/crd/bases/octopinger.io_octopingers.yaml ../charts/octopinger/templates/crds/crd-octopinger.yaml

package api

import (
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen" //nolint:typecheck
)
