/*
 * AppManager API
 *
 * HTTP REST API to connect to the AppManager
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package appManagerApiClient

import (
	"time"
)

type Failure struct {
	Message string `json:"message,omitempty"`
	Reason string `json:"reason,omitempty"`
	FailedAt time.Time `json:"failedAt,omitempty"`
}
