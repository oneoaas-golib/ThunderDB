/*
 * MIT License
 *
 * Copyright (c) 2016-2018. ThunderDB
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the “Software”), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
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
