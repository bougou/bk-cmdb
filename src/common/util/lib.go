/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/storage/dal"

	"github.com/emicklei/go-restful"
	"github.com/rs/xid"
)

func InStrArr(arr []string, key string) bool {
	for _, a := range arr {
		if key == a {
			return true
		}
	}
	return false
}

func GetLanguage(header http.Header) string {
	return header.Get(common.BKHTTPLanguage)
}

func GetUser(header http.Header) string {
	return header.Get(common.BKHTTPHeaderUser)
}

func GetOwnerID(header http.Header) string {
	return header.Get(common.BKHTTPOwnerID)
}

// set supplier id and account in head
func SetOwnerIDAndAccount(req *restful.Request) {
	owner := req.Request.Header.Get(common.BKHTTPOwner)
	if "" != owner {
		req.Request.Header.Set(common.BKHTTPOwnerID, owner)
	}
}

// GetHTTPCCRequestID return config center request id from http header
func GetHTTPCCRequestID(header http.Header) string {
	rid := header.Get(common.BKHTTPCCRequestID)
	return rid
}

// GetSupplierID return supplier_id from http header
func GetSupplierID(header http.Header) (int64, error) {
	return GetInt64ByInterface(header.Get(common.BKHTTPSupplierID))
}

// IsExistSupplierID check supplier_id  exist from http header
func IsExistSupplierID(header http.Header) bool {
	if "" == header.Get(common.BKHTTPSupplierID) {
		return false
	}
	return true
}

// GetHTTPCCTransaction return config center request id from http header
func GetHTTPCCTransaction(header http.Header) string {
	rid := header.Get(common.BKHTTPCCTransactionID)
	return rid
}

// GetDBContext returns a new context that contains JoinOption
func GetDBContext(parent context.Context, header http.Header) context.Context {
	return context.WithValue(parent, common.CCContextKeyJoinOption, dal.JoinOption{
		RequestID: header.Get(common.BKHTTPCCRequestID),
		TxnID:     header.Get(common.BKHTTPCCTransactionID),
	})
}

// IsNil returns whether value is nil value, including map[string]interface{}{nil}, *Struct{nil}
func IsNil(value interface{}) bool {
	rflValue := reflect.ValueOf(value)
	if rflValue.IsValid() {
		return rflValue.IsNil()
	}
	return true
}

type AtomicBool int32

func NewBool(yes bool) *AtomicBool {
	var n = AtomicBool(0)
	if yes {
		n = AtomicBool(1)
	}
	return &n
}

func (b *AtomicBool) SetIfNotSet() bool {
	return atomic.CompareAndSwapInt32((*int32)(b), 0, 1)
}

func (b *AtomicBool) Set() {
	atomic.StoreInt32((*int32)(b), 1)
}

func (b *AtomicBool) UnSet() {
	atomic.StoreInt32((*int32)(b), 0)
}

func (b *AtomicBool) IsSet() bool {
	return atomic.LoadInt32((*int32)(b)) == 1
}

func (b *AtomicBool) SetTo(yes bool) {
	if yes {
		atomic.StoreInt32((*int32)(b), 1)
	} else {
		atomic.StoreInt32((*int32)(b), 0)
	}
}

type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func Ptrue() *bool {
	tmp := true
	return &tmp
}
func Pfalse() *bool {
	tmp := false
	return &tmp
}

// RunForever will run the function forever and rerun the f function if any panic happened
func RunForever(name string, f func() error) {
	for {
		if err := runNoPanic(f); err != nil {
			blog.Errorf("[%s] return %v, retry 3s later", name, err)
			time.Sleep(time.Second * 3)
		}
	}
}

func runNoPanic(f func() error) (err error) {
	defer func() {
		if err != nil {
			return
		}
		if syserr := recover(); err != nil {
			err = fmt.Errorf("panic with error: %v, stack: \n%s", syserr, debug.Stack())
		}
	}()

	err = f()
	return err
}

func GenerateRID() string {
	unused := "0000"
	id := xid.New()
	return fmt.Sprintf("cc%s%s", unused, id.String())
}

// Int64Join []int64 to string
func Int64Join(data []int64, separator string) string {
	var ret string
	for _, item := range data {
		ret += strconv.FormatInt(item, 10) + separator
	}
	return strings.Trim(ret, separator)
}
