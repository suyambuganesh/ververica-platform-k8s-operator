package converters

import (
	"errors"

	"github.com/fintechstudios/ververica-platform-k8s-operator/api/v1beta1"
	platformApiClient "github.com/fintechstudios/ververica-platform-k8s-operator/platform-api-client"
)

var ErrorInvalidNamespaceLifecyclePhase = errors.New("origin must be one of: LIFECYCLE_PHASE_ACTIVE, LIFECYCLE_PHASE_TERMINATING, UNRECOGNIZED, LIFECYCLE_PHASE_INVALID")

func NamespaceLifecyclePhaseToNative(savepointOrigin string) (v1beta1.NamespaceLifecyclePhase, error) {
	switch savepointOrigin {
	case string(v1beta1.InvalidNamespaceLifecyclePhase):
		return v1beta1.InvalidNamespaceLifecyclePhase, nil
	case string(v1beta1.ActiveNamespaceLifecyclePhase):
		return v1beta1.ActiveNamespaceLifecyclePhase, nil
	case string(v1beta1.TerminatingNamespaceLifecyclePhase):
		return v1beta1.TerminatingNamespaceLifecyclePhase, nil
	case string(v1beta1.UnrecognizedNamespaceLifecyclePhase):
		return v1beta1.UnrecognizedNamespaceLifecyclePhase, nil
	default:
		return "", ErrorInvalidNamespaceLifecyclePhase
	}
}

func NamespaceLifecyclePhaseFromNative(vpSavepointOrigin v1beta1.NamespaceLifecyclePhase) (string, error) {
	switch vpSavepointOrigin {
	case v1beta1.InvalidNamespaceLifecyclePhase:
		return string(v1beta1.InvalidNamespaceLifecyclePhase), nil
	case v1beta1.ActiveNamespaceLifecyclePhase:
		return string(v1beta1.ActiveNamespaceLifecyclePhase), nil
	case v1beta1.TerminatingNamespaceLifecyclePhase:
		return string(v1beta1.TerminatingNamespaceLifecyclePhase), nil
	case v1beta1.UnrecognizedNamespaceLifecyclePhase:
		return string(v1beta1.UnrecognizedNamespaceLifecyclePhase), nil
	default:
		return "", ErrorInvalidNamespaceLifecyclePhase
	}
}

func NamespaceRoleBindingsFromNative(nativeBindings []v1beta1.NamespaceRoleBinding) []platformApiClient.RoleBinding {
	bindings := make([]platformApiClient.RoleBinding, len(nativeBindings))
	for i, nativeBinding := range nativeBindings {
		bindings[i] = platformApiClient.RoleBinding{
			Members: nativeBinding.Members,
			Role:    nativeBinding.Role,
		}
	}
	return bindings
}

func NamespaceRoleBindingsToNative(bindings []platformApiClient.RoleBinding) []v1beta1.NamespaceRoleBinding {
	nativeBindings := make([]v1beta1.NamespaceRoleBinding, len(bindings))
	for i, nativeBinding := range bindings {
		nativeBindings[i] = v1beta1.NamespaceRoleBinding{
			Members: nativeBinding.Members,
			Role:    nativeBinding.Role,
		}
	}
	return nativeBindings
}
