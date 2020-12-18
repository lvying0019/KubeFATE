/*
 * Copyright 2019-2020 VMware, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package modules

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"sigs.k8s.io/yaml"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	uuid "github.com/satori/go.uuid"
)

type Cluster struct {
	// UUID is created by github.com/satori/go A combination of 36 bit strings generated by. UUID
	Uuid      string `json:"uuid" gorm:"type:varchar(36);index:uuid;unique"`
	Name      string `json:"name" gorm:"type:varchar(255);not null"`
	NameSpace string `json:"namespaces" gorm:"type:varchar(255);not null"`

	ChartName    string `json:"chart_name" gorm:"type:varchar(255)"`
	ChartVersion string `json:"chart_version" gorm:"type:varchar(255);not null"`

	//Values field storage cluster.yaml File content of
	Values string `json:"values" gorm:"type:text"`
	//Spec corresponding to values(of Cluster) is decoded into interface by yaml
	Spec MapStringInterface `json:"Spec,omitempty" gorm:"type:blob"`

	// Cluster revision
	Revision int8 `json:"revision" gorm:"size:8"`
	// Cluster revision of helm
	HelmRevision int8 `json:"helm_revision" gorm:"size:8"`
	//Values through the values in the chart file- template.yaml The file generates the corresponding helm values file
	ChartValues MapStringInterface `json:"chart_values" gorm:"type:blob"`

	//The status of the cluster, including: "Creating","Deleting","Updating","Running","Unavailable","Deleted"
	Status ClusterStatus `json:"status"  gorm:"size:8"`

	//Info is the corresponding information of cluster in k8s
	Info MapStringInterface `json:"Info,omitempty" gorm:"type:blob"`

	gorm.Model
}

type MapStringInterface map[string]interface{}

type Clusters []Cluster

type ClusterStatus int8

const (
	ClusterStatusPending ClusterStatus = iota + 1
	ClusterStatusCreating
	ClusterStatusDeleting
	ClusterStatusUpdating
	ClusterStatusRunning
	ClusterStatusUnavailable
	ClusterStatusDeleted
	ClusterStatusRollback
	ClusterStatusFailed
	ClusterStatusUnknown
)

func (s ClusterStatus) String() string {
	names := map[ClusterStatus]string{
		ClusterStatusPending:     "Pending",
		ClusterStatusCreating:    "Creating",
		ClusterStatusDeleting:    "Deleting",
		ClusterStatusUpdating:    "Updating",
		ClusterStatusRunning:     "Running",
		ClusterStatusUnavailable: "Unavailable",
		ClusterStatusDeleted:     "Deleted",
		ClusterStatusRollback:    "Rollback",
		ClusterStatusFailed:      "Failed",
		ClusterStatusUnknown:     "Unknown",
	}

	return names[s]
}

// MarshalJSON convert cluster status to string
func (s *ClusterStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON sets *m to a copy of data.
func (s *ClusterStatus) UnmarshalJSON(data []byte) error {
	if s == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	var ClusterStatus ClusterStatus
	switch string(data) {
	case "\"Pending\"":
		ClusterStatus = ClusterStatusPending
	case "\"Creating\"":
		ClusterStatus = ClusterStatusCreating
	case "\"Deleting\"":
		ClusterStatus = ClusterStatusDeleting
	case "\"Updating\"":
		ClusterStatus = ClusterStatusUpdating
	case "\"Running\"":
		ClusterStatus = ClusterStatusRunning
	case "\"Unavailable\"":
		ClusterStatus = ClusterStatusUnavailable
	case "\"Deleted\"":
		ClusterStatus = ClusterStatusDeleted
	case "\"Rollback\"":
		ClusterStatus = ClusterStatusRollback
	case "\"Failed\"":
		ClusterStatus = ClusterStatusFailed
	case "\"Unknown\"":
		ClusterStatus = ClusterStatusUnknown
	default:
		return errors.New("data can't UnmarshalJSON")
	}
	*s = ClusterStatus
	return nil
}

// NewCluster create cluster object with basic argument
func NewCluster(name string, nameSpaces, chartName, chartVersion, values string) (*Cluster, error) {
	var spec MapStringInterface
	err := yaml.Unmarshal([]byte(values), &spec)
	if err != nil {
		return nil, err
	}
	cluster := &Cluster{
		Uuid:         uuid.NewV4().String(),
		Name:         name,
		NameSpace:    nameSpaces,
		Revision:     0,
		Status:       ClusterStatusPending,
		ChartName:    chartName,
		ChartVersion: chartVersion,
		Values:       values,
		Spec:         spec,
	}

	return cluster, nil
}

func (s MapStringInterface) Value() (driver.Value, error) {
	bJson, err := json.Marshal(s)
	return bJson, err
}

func (s *MapStringInterface) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), s)
}
