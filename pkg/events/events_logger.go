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
	"encoding/json"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	kubeinformersinterfaces "k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func NewEventInformerFuncForNamespace(origin, namespace string) kubeinformersinterfaces.NewInformerFunc {
	return func(clientset kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		watchlist := cache.NewListWatchFromClient(
			clientset.CoreV1().RESTClient(),
			"events",
			namespace,
			fields.Everything(),
		)
		informer := cache.NewSharedIndexInformer(
			watchlist,
			&v1.Event{},
			resyncPeriod,
			cache.Indexers{},
		)
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if isV1Event(obj) {
					v1Event := obj.(*v1.Event)
					printV1Event(v1Event, origin)
				}
			},
			UpdateFunc: func(oldObj interface{}, newObject interface{}) {
				if isV1Event(newObject) {
					v1Event := newObject.(*v1.Event)
					printV1Event(v1Event, origin)
				}
			},
		})
		return informer
	}
}

func isV1Event(obj interface{}) bool {
	_, ok := obj.(*v1.Event)
	return ok
}

func getEventFromV1Event(v1Event *v1.Event, origin string) *event {
	involvedObject := v1Event.InvolvedObject.Name
	if v1Event.InvolvedObject.Kind != "" {
		involvedObject = v1Event.InvolvedObject.Kind + "/" + involvedObject
	}

	return &event{
		Origin:         origin,
		Namespace:      v1Event.Namespace,
		Type:           v1Event.Type,
		Count:          v1Event.Count,
		FirstTimestamp: v1Event.FirstTimestamp,
		LastTimestamp:  v1Event.LastTimestamp,
		Reason:         v1Event.Reason,
		Object:         involvedObject,
		Message:        v1Event.Message,
		Source:         v1Event.Source.Component,
		SourceHost:     v1Event.Source.Host,
	}
}

func isOlderThan(event *v1.Event, than time.Duration) bool {
	return time.Since(event.CreationTimestamp.Time) > than
}

func printV1Event(v1Event *v1.Event, origin string) {
	if isOlderThan(v1Event, time.Second*5) {
		return
	}
	j, err := json.Marshal(getEventFromV1Event(v1Event, origin))
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	fmt.Printf("%s\n", string(j))
}
