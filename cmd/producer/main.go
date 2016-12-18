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
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/websocket"
)

var (
	store int
	mutex *sync.Mutex
	cond  *sync.Cond

	metric prometheus.Gauge
)

func init() {
	store = 40
	mutex = &sync.Mutex{}
	cond = sync.NewCond(mutex)

	metric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "spate_demo_store",
		Help: "A demo metric",
	})
	metric.Set(float64(store))
}

func main() {
	go producer()

	prometheus.Register(metric)

	http.Handle("/", http.HandlerFunc(indexHandler))
	http.Handle("/consume", http.HandlerFunc(consumeHandler))
	http.Handle("/stats", websocket.Handler(statsHandler))
	http.Handle("/metrics", prometheus.Handler())

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func producer() {
	for {
		mutex.Lock()
		store = store + rand.Intn(10)
		metric.Set(float64(store))
		cond.Broadcast()
		mutex.Unlock()

		<-time.After(time.Second)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, indexPage)
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
		mutex.Lock()
		websocket.JSON.Send(ws, map[string]string{
			"store":    strconv.Itoa(store),
			"replicas": strconv.Itoa(0),
		})
		mutex.Unlock()

		<-time.After(2 * time.Second)
	}
}

const indexPage = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>spate demo</title>

  <script type="text/javascript" src="http://smoothiecharts.org/smoothie.js"></script>
  <script type="text/javascript">
    var store = new TimeSeries();
    var replicas = new TimeSeries();

    var ws = new WebSocket("ws://localhost:5000/stats");
    ws.onmessage = function(e) {
      var data = JSON.parse(e.data);
      var time = new Date().getTime()
      store.append(time, data.store);
      replicas.append(time, data.replicas);
    };

    function init() {
      var storeChart = new SmoothieChart();
      storeChart.addTimeSeries(store, {
        strokeStyle: 'rgba(255, 0, 0, 1)',
        fillStyle: 'rgba(255, 0, 0, 0.2)',
        lineWidth: 4
      });
      storeChart.streamTo(document.getElementById("storeChart"), 2000);

      var replicaChart = new SmoothieChart();
      replicaChart.addTimeSeries(replicas, {
        strokeStyle: 'rgba(0, 255, 0, 1)',
        fillStyle: 'rgba(0, 255, 0, 0.2)',
        lineWidth: 4
      });
      replicaChart.streamTo(document.getElementById("replicaChart"), 2000);
    }
  </script>
</head>
<body onload="init()" style="margin: 0">
  <canvas id="storeChart" width="1200" height="400"></canvas>
  <canvas id="replicaChart" width="1200" height="400"></canvas>
</body>
</html>
`
