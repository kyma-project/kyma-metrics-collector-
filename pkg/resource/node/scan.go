package node

import (
	"errors"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/kyma-project/kyma-metrics-collector/pkg/config"
	"github.com/kyma-project/kyma-metrics-collector/pkg/resource"
	"github.com/kyma-project/kyma-metrics-collector/pkg/runtime"
)

const nodeInstanceTypeLabel = "node.kubernetes.io/instance-type"

var _ resource.ScanConverter = &Scan{}

type Scan struct {
	provider runtime.ProviderType
	nodes    corev1.NodeList
}

func (s *Scan) UM(duration time.Duration) (resource.UMMeasurement, error) {
	return resource.UMMeasurement{}, nil
}

func (s *Scan) EDP(specs *config.PublicCloudSpecs) (resource.EDPMeasurement, error) {
	edp := resource.EDPMeasurement{}
	var errs []error
	vmTypes := make(map[string]int)

	for _, node := range s.nodes.Items {
		nodeType := node.Labels[nodeInstanceTypeLabel]
		nodeType = strings.ToLower(nodeType)

		vmFeature := specs.GetFeature(string(s.provider), nodeType)
		if vmFeature == nil {
			errs = append(errs, fmt.Errorf("providerType: %s and nodeType: %s does not exist in the map", s.provider, nodeType))
			continue
		}

		edp.ProvisionedCPUs += vmFeature.CpuCores
		edp.ProvisionedRAMGb += vmFeature.Memory
		vmTypes[nodeType] += 1
	}

	for vmType, count := range vmTypes {
		edp.VMTypes = append(edp.VMTypes, resource.VMType{
			Name:  vmType,
			Count: count,
		})
	}

	return edp, errors.Join(errs...)
}