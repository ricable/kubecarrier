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

package controllers

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestCustomResourceDiscoveryReconciler(t *testing.T) {
	const (
		serviceClusterName = "eu-west-1"
	)
	crDiscovery := &corev1alpha1.CustomResourceDiscovery{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis.cloud",
			Namespace: "extreme-cloud",
		},
		Spec: corev1alpha1.CustomResourceDiscoverySpec{
			CRD:            corev1alpha1.ObjectReference{Name: "redis.cloud"},
			ServiceCluster: corev1alpha1.ObjectReference{Name: serviceClusterName},
		},
	}
	crDiscoveryNN := types.NamespacedName{
		Namespace: crDiscovery.Namespace,
		Name:      crDiscovery.Name,
	}

	crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "redis.cloud",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "cloud",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   "redis",
				Singular: "redis",
				Kind:     "Redis",
				ListKind: "RedisList",
			},
			Scope: "Namespaced",
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{Name: "corev1alpha1"},
			},
		},
	}
	r := &CustomResourceDiscoveryReconciler{
		Log:    testutil.NewLogger(t),
		Client: fakeclient.NewFakeClientWithScheme(testScheme, crDiscovery),
		Scheme: testScheme,
	}
	ctx := context.Background()
	reconcileLoop := func() {
		for i := 0; i < 3; i++ {
			_, err := r.Reconcile(ctrl.Request{
				NamespacedName: crDiscoveryNN,
			})
			require.NoError(t, err)
			require.NoError(t, r.Client.Get(ctx, crDiscoveryNN, crDiscovery))
		}
	}

	reconcileLoop() // should not panic on undiscovered instances

	crDiscovery.Status.CRD = crd
	crDiscovery.Status.SetCondition(corev1alpha1.CustomResourceDiscoveryCondition{
		Type:   corev1alpha1.CustomResourceDiscoveryDiscovered,
		Status: corev1alpha1.ConditionTrue,
	})
	require.NoError(t, r.Client.Status().Update(ctx, crDiscovery))

	reconcileLoop() // creates the CRD in the master cluster

	establishedCondition, ok := crDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDiscoveryEstablished)
	if assert.True(t, ok) {
		assert.Equal(t, corev1alpha1.ConditionFalse, establishedCondition.Status)
		assert.Equal(t, "Establishing", establishedCondition.Reason)
	}

	internalCRD := &apiextensionsv1.CustomResourceDefinition{}
	require.NoError(t, r.Client.Get(ctx, types.NamespacedName{
		Name: strings.Join([]string{"redis", serviceClusterName, crDiscovery.Namespace}, "."),
	}, internalCRD))

	internalCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:   apiextensionsv1.Established,
			Status: apiextensionsv1.ConditionTrue,
		},
	}
	require.NoError(t, r.Client.Status().Update(ctx, internalCRD))

	reconcileLoop() // updates the status to established and launches Catapult

	establishedCondition, ok = crDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDiscoveryEstablished)
	if assert.True(t, ok) {
		assert.Equal(t, corev1alpha1.ConditionTrue, establishedCondition.Status)
		assert.Equal(t, "Established", establishedCondition.Reason)
	}

	catapult := &operatorv1alpha1.Catapult{}
	require.NoError(t, r.Client.Get(ctx, types.NamespacedName{
		Name:      crDiscovery.Name,
		Namespace: crDiscovery.Namespace,
	}, catapult))
	catapult.Status.Conditions = []operatorv1alpha1.CatapultCondition{
		{
			Type:   operatorv1alpha1.CatapultReady,
			Status: operatorv1alpha1.ConditionTrue,
		},
	}
	require.NoError(t, r.Client.Status().Update(ctx, catapult))

	reconcileLoop() // updates status to ready

	readyCondition, ok := crDiscovery.Status.GetCondition(corev1alpha1.CustomResourceDiscoveryReady)
	if assert.True(t, ok) {
		assert.Equal(t, corev1alpha1.ConditionTrue, readyCondition.Status)
		assert.Equal(t, "ComponentsReady", readyCondition.Reason)
	}
}