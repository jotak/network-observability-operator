// Package constants defines some values that are shared across multiple packages
package constants

const (
	GoflowKubeName = "goflow-kube"
	DeploymentKind = "Deployment"
	DaemonSetKind  = "DaemonSet"
)

var Labels = []string{"SrcK8S_Namespace", "DstK8S_Namespace", "FlowDirection"}
