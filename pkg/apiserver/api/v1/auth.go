/*
Copyright 2019 The KubeCarrier Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/authorizer"
)

type ServerGVRGetter interface {
	GetGVR() schema.GroupVersionResource
}

func (req *ListRequest) GetAuthOption() authorizer.AuthorizationOption {
	return authorizer.AuthorizationOption{
		Namespace: req.Account,
		Verb:      authorizer.RequestList,
	}
}

func (req *ListRequest) GetGVR(server ServerGVRGetter) schema.GroupVersionResource {
	if gvrSrv, ok := server.(ServerGVRGetter); ok {
		return gvrSrv.GetGVR()
	}
	return schema.GroupVersionResource{}
}

func (req *GetRequest) GetAuthOption() authorizer.AuthorizationOption {
	return authorizer.AuthorizationOption{
		Name:      req.Name,
		Namespace: req.Account,
		Verb:      authorizer.RequestGet,
	}
}

func (req *GetRequest) GetGVR(server ServerGVRGetter) schema.GroupVersionResource {
	if gvrSrv, ok := server.(ServerGVRGetter); ok {
		return gvrSrv.GetGVR()
	}
	return schema.GroupVersionResource{}
}

func (req *WatchRequest) GetAuthOption() authorizer.AuthorizationOption {
	return authorizer.AuthorizationOption{
		Namespace: req.Account,
		Verb:      authorizer.RequestWatch,
	}
}

func (req *WatchRequest) GetGVR(server ServerGVRGetter) schema.GroupVersionResource {
	if gvrSrv, ok := server.(ServerGVRGetter); ok {
		return gvrSrv.GetGVR()
	}
	return schema.GroupVersionResource{}
}

func (req *InstanceCreateRequest) GetAuthOption() authorizer.AuthorizationOption {
	return authorizer.AuthorizationOption{
		Namespace: req.Account,
		Verb:      authorizer.RequestCreate,
	}
}

func (req *InstanceCreateRequest) GetGVR(server interface{}) schema.GroupVersionResource {
	return GetOfferingGVR(req)
}

func (req *InstanceDeleteRequest) GetAuthOption() authorizer.AuthorizationOption {
	return authorizer.AuthorizationOption{
		Namespace: req.Account,
		Name:      req.Name,
		Verb:      authorizer.RequestDelete,
	}
}
func (req *InstanceDeleteRequest) GetGVR(server interface{}) schema.GroupVersionResource {
	return GetOfferingGVR(req)
}

func (req *InstanceGetRequest) GetAuthOption() authorizer.AuthorizationOption {
	return authorizer.AuthorizationOption{
		Namespace: req.Account,
		Name:      req.Name,
		Verb:      authorizer.RequestGet,
	}
}
func (req *InstanceGetRequest) GetGVR(server interface{}) schema.GroupVersionResource {
	return GetOfferingGVR(req)
}

func (req *InstanceListRequest) GetAuthOption() authorizer.AuthorizationOption {
	return authorizer.AuthorizationOption{
		Namespace: req.Account,
		Verb:      authorizer.RequestList,
	}
}

func (req *InstanceListRequest) GetGVR(server interface{}) schema.GroupVersionResource {
	return GetOfferingGVR(req)
}