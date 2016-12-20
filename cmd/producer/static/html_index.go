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

package static

const IndexPage = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>spate demo</title>

  <script type="text/javascript" src="/smoothie.js"></script>
  <script type="text/javascript">
    var store = new TimeSeries();
    var desiredReplicas = new TimeSeries();
    var actualReplicas = new TimeSeries();

    var ws = new WebSocket("ws://" + location.host + "/stats");
    ws.onmessage = function(e) {
      var data = JSON.parse(e.data);
      var time = new Date().getTime()
      store.append(time, data.store);
      desiredReplicas.append(time, data.desiredReplicas);
      actualReplicas.append(time, data.actualReplicas);
    };

    function init() {
      var storeChart = new SmoothieChart({
        minValue: 0,
        maxValue: 400
      });
      storeChart.addTimeSeries(store, {
        strokeStyle: 'rgba(255, 0, 0, 1)',
        fillStyle: 'rgba(255, 0, 0, 0.2)',
        lineWidth: 4
      });
      storeChart.streamTo(document.getElementById("storeChart"), 1000);

      var replicaChart = new SmoothieChart({
        minValue: 0,
        maxValue: 25
      });
      replicaChart.addTimeSeries(desiredReplicas, {
        strokeStyle: 'rgba(0, 0, 255, 1)',
        fillStyle: 'rgba(0, 0, 255, 0.2)',
        lineWidth: 4
      });
      replicaChart.addTimeSeries(actualReplicas, {
        strokeStyle: 'rgba(0, 255, 0, 1)',
        fillStyle: 'rgba(0, 255, 0, 0.2)',
        lineWidth: 4
      });
      replicaChart.streamTo(document.getElementById("replicaChart"), 1000);
    }

    function configure() {
      var amount    = document.getElementById("amount").value,
          variation = document.getElementById("variation").value;

      fetch(
        "/config?amount="+amount+"&variation="+variation
      ).then(function(resp) {
        console.log(resp);
      }).catch(function(err) {
        console.log(err);
      });
    }
  </script>
</head>
<body onload="init()">
  <h2>Workload</h2>

  <p>
    <label for="amount">amount: </label>
    <input type="text" id="amount" name="amount" value="5">

    <label for="variation">variation: </label>
    <input type="text" id="variation" name="variation" value="1">

    <button onclick="configure()">configure</button>
  </p>

  <canvas id="storeChart" width="1200" height="400"></canvas>
  <h2>Number of worker</h2>
  <canvas id="replicaChart" width="1200" height="400"></canvas>
</body>
</html>
`
