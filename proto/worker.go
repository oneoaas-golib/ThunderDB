/*
 * Copyright 2018 The ThunderDB Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the “License”);
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an “AS IS” BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package proto

import "time"

// ServiceUpdateType database service update operation type
type ServiceUpdateType uint

// database service update operation type values
const (
	ServiceUpdateTypeCreate ServiceUpdateType = iota
	ServiceUpdateTypeUpdate
	ServiceUpdateTypeDrop
)

// QueryID query id definition
type QueryID string

// PrepareQueryReq prepare query request entity
type PrepareQueryReq struct {
	Node    Node          // issuing client node info
	QueryID QueryID       // query id
	Timeout time.Duration // query expire time duration
}

// PrepareQueryResp prepare query response entity
type PrepareQueryResp struct {
	Success bool
	Msg     string // response status text
	Envelope
}

// ExecuteQueryReq execute query request entity
type ExecuteQueryReq struct {
	Node    Node
	QueryID QueryID
	Query   string
	Envelope
}

// ExecuteQueryResp execute query response entity
type ExecuteQueryResp struct {
	Success bool
	Msg     string // response status text
	Envelope
}

// ConfirmQueryReq confirm query request entity
type ConfirmQueryReq struct {
	Node    Node
	QueryID string
	Envelope
}

// ConfirmQueryResp confirm query response entity
type ConfirmQueryResp struct {
	Success bool
	Msg     string // response status text
	Envelope
}

// UpdateServiceReq update service request entity
type UpdateServiceReq struct {
	Node          Node              // requiring super nodes
	DatabaseID    string            // database identifier in UUID
	ReservedSpace uint64            // database reserved max space in bytes
	Peers         []Node            // database service nodes (if op == add/update, current node itself is included)
	OpType        ServiceUpdateType // service update operation type
}

// UpdateServiceResp update service response entity
type UpdateServiceResp struct {
	Success bool
	Msg     string // response status text
	Envelope
}
