/*
 * AppManager API
 *
 * HTTP REST API to connect to the AppManager
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package ververicaplatformapi

type JobDeploymentTarget struct {
	Kubernetes *KubernetesTarget `json:"kubernetes,omitempty"`
}
