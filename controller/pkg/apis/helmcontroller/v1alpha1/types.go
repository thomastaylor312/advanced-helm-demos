package v1alpha1

import (
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Helm is a specification for a Helm resource
type Helm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmSpec   `json:"spec"`
	Status HelmStatus `json:"status"`
}

// HelmSpec is the spec for a Helm resource
type HelmSpec struct {
	ChartName string `json:"chartName"`
	Replicas  int32  `json:"replicas"`
}

// HelmStatus is the status for a Helm resource
type HelmStatus struct {
	ReleaseStatus release.Status `json:"releaseStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelmList is a list of Helm resources
type HelmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Helm `json:"items"`
}
