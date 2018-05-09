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

package worker

import "github.com/thunderdb/ThunderDB/proto"

// Service is a database worker side RPC implementation
type Service struct {
}

// NewService will return a new Service
func NewService() *Service {
	return &Service{}
}

// bootstrap Service instance
// 1. send local WorkerMeta to super nodes
// 2. fetch database topology from super nodes
// 3. init database storage for issued databases
// 4. form database replication chains with peer nodes
func (worker *Service) bootstrap() {

}

// issue Service meta update
// 1. sending meta update request to super nodes
// 2. super nodes will sending meta update callback to UpdateServiceMeta method
func (worker *Service) updateMeta() {

}

// PrepareQuery is used to process prepare query from client nodes
func (worker *Service) PrepareQuery(req *proto.PrepareQueryReq, resp *proto.PrepareQueryResp) (err error) {
	return
}

// ExecuteQuery is used to process real query from client nodes
func (worker *Service) ExecuteQuery(req *proto.ExecuteQueryReq, resp *proto.ExecuteQueryResp) (err error) {
	return
}

// ConfirmQuery is used to process confirm query from client nodes
func (worker *Service) ConfirmQuery(req *proto.ConfirmQueryReq, resp *proto.ConfirmQueryResp) (err error) {
	return
}

// UpdateService is used to process service update from super nodes
func (worker *Service) UpdateService(req *proto.UpdateServiceReq, resp *proto.UpdateServiceResp) {
	return
}
