/*
 * AppManager API
 *
 * HTTP REST API to connect to the AppManager
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package ververicaplatformapi

type DeploymentMetadata struct {
	Id              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	Namespace       string            `json:"namespace,omitempty"`
	CreatedAt       int64             `json:"createdAt,omitempty"`
	ModifiedAt      int64             `json:"modifiedAt,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	ResourceVersion int32             `json:"resourceVersion,omitempty"`
}
