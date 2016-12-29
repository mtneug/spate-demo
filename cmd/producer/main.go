// Copyright (C) 2016 Matthias Neugebauer <mtneug@mailbox.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/mtneug/spate-demo/cmd/producer/static"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/websocket"
)

const consumerSrvName = "consumer"

var (
	store int
	mutex *sync.Mutex
	cond  *sync.Cond

	docker          *client.Client
	desiredReplicas uint64
	actualReplicas  uint64

	amount    = 5
	variation = 1

	metric prometheus.Gauge
)

func init() {
	var err error
	docker, err = client.NewEnvClient()
	if err != nil {
		log.Fatal("Connection to Docker failed")
	}

	store = 40
	mutex = &sync.Mutex{}
	cond = sync.NewCond(mutex)

	metric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "spate_demo_store",
		Help: "A demo metric",
	})
	metric.Set(float64(store))
	prometheus.Register(metric)
}

func main() {
	go producer()

	http.Handle("/", http.HandlerFunc(indexHandler))
	http.Handle("/smoothie.js", http.HandlerFunc(smoothieJSHandler))
	http.Handle("/config", http.HandlerFunc(configHandler))
	http.Handle("/consume", http.HandlerFunc(consumeHandler))
	http.Handle("/stats", websocket.Handler(statsHandler))
	http.Handle("/metrics", prometheus.Handler())

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func producer() {
	for {
		mutex.Lock()
		store = store + amount + rand.Intn(variation+1)
		metric.Set(float64(store))
		cond.Broadcast()
		mutex.Unlock()

		<-time.After(time.Second)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, static.IndexPage)
}

func smoothieJSHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, static.SmoothieJS)
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	aStr := r.URL.Query().Get("amount")
	a, err := strconv.ParseInt(aStr, 10, 32)
	if err == nil {
		amount = int(a)
	}

	vStr := r.URL.Query().Get("variation")
	v, err := strconv.ParseInt(vStr, 10, 32)
	if err == nil {
		variation = int(v)
	}
}

func consumeHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	for store == 0 {
		cond.Wait()
	}

	store = store - int(math.Min(float64(rand.Intn(3)), float64(store)))
	metric.Set(float64(store))

	w.WriteHeader(http.StatusNoContent)
}

func statsHandler(ws *websocket.Conn) {
	for {
		err := updateDesiredReplicas()
		if err != nil {
			log.Print(err.Error())
		}

		err = updateActualReplicas()
		if err != nil {
			log.Print(err.Error())
		}

		mutex.Lock()
		websocket.JSON.Send(ws, map[string]string{
			"store":           strconv.Itoa(store),
			"desiredReplicas": strconv.FormatUint(desiredReplicas, 10),
			"actualReplicas":  strconv.FormatUint(actualReplicas, 10),
		})
		mutex.Unlock()

		<-time.After(2 * time.Second)
	}
}

func updateDesiredReplicas() error {
	srv, _, err := docker.ServiceInspectWithRaw(context.TODO(), consumerSrvName)
	if err != nil {
		return errors.New("Counting desiredReplicas failed")
	}

	srvMode := srv.Spec.Mode
	if srvMode.Replicated == nil || srvMode.Replicated.Replicas == nil {
		return errors.New("Not a replicated service")
	}
	desiredReplicas = *srv.Spec.Mode.Replicated.Replicas
	return nil
}

func updateActualReplicas() error {
	args := filters.NewArgs()
	args.Add("service", consumerSrvName)
	args.Add("desired-state", "running")

	tasks, err := docker.TaskList(context.TODO(), types.TaskListOptions{Filter: args})
	if err != nil {
		return errors.New("Counting actualReplicas failed")
	}

	actualReplicas = 0
	for _, t := range tasks {
		if t.Status.State == swarm.TaskStateRunning {
			actualReplicas++
		}
	}

	return nil
}
