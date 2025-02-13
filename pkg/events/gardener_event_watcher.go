// Copyright (c) 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events

import (
	v1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
)

type GardenerEventWatcherConfig struct {
	SeedEventWatcherConfig     EventWatcherConfig
	SeedKubeInformerFactories  []kubeinformers.SharedInformerFactory
	ShootEventWatcherConfig    EventWatcherConfig
	ShootKubeInformerFactories []kubeinformers.SharedInformerFactory
}

type GardenerEventWatcher struct {
	SeedKubeInformerFactories  []kubeinformers.SharedInformerFactory
	ShootKubeInformerFactories []kubeinformers.SharedInformerFactory
}

func (e *GardenerEventWatcherConfig) New() *GardenerEventWatcher {
	for indx, namespace := range e.SeedEventWatcherConfig.Namespaces {
		_ = e.SeedKubeInformerFactories[indx].InformerFor(&v1.Event{},
			NewEventInformerFuncForNamespace(
				"seed",
				namespace,
			),
		)
	}

	for indx, namespace := range e.ShootEventWatcherConfig.Namespaces {
		_ = e.ShootKubeInformerFactories[indx].InformerFor(&v1.Event{},
			NewEventInformerFuncForNamespace(
				"shoot",
				namespace,
			),
		)
	}

	return &GardenerEventWatcher{
		SeedKubeInformerFactories:  e.SeedKubeInformerFactories,
		ShootKubeInformerFactories: e.ShootKubeInformerFactories,
	}
}

func (e *GardenerEventWatcher) Run(stopCh <-chan struct{}) {
	for _, informerFactory := range e.SeedKubeInformerFactories {
		informerFactory.Start(stopCh)
	}

	for _, informerFactory := range e.ShootKubeInformerFactories {
		informerFactory.Start(stopCh)
	}
	<-stopCh
}
