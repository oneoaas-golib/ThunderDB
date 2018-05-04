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
