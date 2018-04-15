<template>
<div>
  <div id='map'></div>
  <div id="legend">
    <div>
      <svg width="40" height="20">
        <rect width="40" height="20" style="fill:#009900;stroke-width:3;stroke:rgb(0,0,0)" />
      </svg>
      Open Spaces -- {{ count.unoccupied }}
    </div>
    <div>
      <svg width="40" height="20">
        <rect width="40" height="20" style="fill:#ff0000;stroke-width:3;stroke:rgb(0,0,0)" />
      </svg>
      Occupied Spaces -- {{ count.occupied }}
    </div>
    <div>
      Total Spaces -- {{ count.occupied + count.unoccupied }}
    </div>
  </div>
</div>
</template>

<script>
import axios from "axios";
import xs from "xstream";
import _ from "lodash";
var flatbuffers = require("../../node_modules/flatbuffers").flatbuffers;
var melbourne = require("../melbourne/schema_generated.js").melbourne;

export default {
  data() {
    return {
      map: null,
      geocoder: null,

      producer: {
        start: listener => {
          this.$store.commit("startWS");
          this.$store.state.ws.onmessage = event => {
            let bytes = new Uint8Array(event.data);
            let buf = new flatbuffers.ByteBuffer(bytes);
            let msg = melbourne.Message.getRootAsMessage(buf);
            listener.next(msg);
          };
        },
        stop: () => {
          console.log("No longer listening to websocket.");
        }
      },
      updateListener: {
        next: msg => {
          this.$store.commit("updateSpots", msg);
          this.$store.commit("updateFeatures");
        },
        error: err => {
          console.error(err);
        },
        complete: () => {
          console.log("update stream complete.");
        }
      }
    };
  },
  computed: {
    count() {
      return this.$store.state.count;
    },
    features() {
      return this.$store.state.features;
    },
    openSpaces() {
      return this.$store.state.openSpaces;
    },
    searchPoint() {
      return this.$store.state.searchPoint;
    },
    directions() {
      return this.$store.state.directions;
    },
    main$() {
      return xs.createWithMemory(this.producer);
    },
    update$() {
      return xs.from(this.main$);
    }
  },
  watch: {
    features(ft) {
      this.map.getSource("spots").setData({
        type: "FeatureCollection",
        features: ft
      });
    },
    searchPoint(sp) {
      this.map.getSource("search-point").setData({
        type: "FeatureCollection",
        features: sp
      });
    },
    directions(dir) {
      this.map.getSource("directions").setData({
        type: "FeatureCollection",
        features: dir
      });
    }
  },
  mounted() {
    this.update$.addListener(this.updateListener);

    mapboxgl.accessToken =
      "pk.eyJ1IjoibWFyY2NiIiwiYSI6ImNqYTR1enN2dGE0bWEyd3BhcTd6cnBzc3MifQ.Z4zYRzVCXv5zCqqdpgKZ-w";
    this.map = new mapboxgl.Map({
      container: "map", // container id
      style: "mapbox://styles/mapbox/light-v9",
      center: [144.963056, -37.813611], // starting position [lng, lat]
      zoom: 13.75
    });

    this.map.on("load", () => {
      axios({
        url: "/api/spots",
        method: "get",
        responseType: "arraybuffer"
      })
        .then(response => {
          let bytes = new Uint8Array(response.data);
          let buf = new flatbuffers.ByteBuffer(bytes);
          let msg = melbourne.Message.getRootAsMessage(buf);
          this.$store.commit("updateSpots", msg);
          this.$store.commit("updateFeatures");

          this.map.addSource("spots", {
            type: "geojson",
            data: {
              type: "FeatureCollection",
              features: this.features
            }
          });

          this.map.addSource("search-point", {
            type: "geojson",
            data: {
              type: "FeatureCollection",
              features: this.searchPoint
            }
          });

          this.map.addSource("directions", {
            type: "geojson",
            data: {
              type: "FeatureCollection",
              features: this.directions
            }
          });

          this.map.addLayer({
            id: "spotsLayer",
            type: "circle",
            source: "spots",
            paint: {
              "circle-radius": {
                base: 1.75,
                stops: [[11, 1], [22, 35]]
              },
              "circle-color": [
                "match",
                ["get", "status"],
                "Present",
                "#ff0000",
                "Unoccupied",
                "#009900",
                "#000000"
              ]
            }
          });

          this.map.addLayer({
            id: "searchLayer",
            type: "circle",
            source: "search-point",
            paint: {
              "circle-radius": 10,
              "circle-color": "#007cbf",
              "circle-stroke-width": 3,
              "circle-stroke-color": "#fff"
            }
          });

          this.map.addLayer({
            id: "directionsLayer",
            type: "line",
            source: "directions",
            paint: {
              "line-width": 3,
              "line-color": "#007cbf",
              "line-width": 3
            }
          });
        })
        .catch(err => {
          console.error(err);
        });

      this.geocoder = new MapboxGeocoder({
        accessToken: mapboxgl.accessToken,
        bbox: [144.932102, -37.834726, 144.987248, -37.792422]
      });

      this.map.addControl(this.geocoder, "top-right");
      // Add zoom and rotation controls to the map.
      this.map.addControl(new mapboxgl.NavigationControl());

      this.geocoder.on("result", e => {
        this.$store.commit("updateSearchPoint", e.result.geometry);
        var closestSpaces = _.chain(this.openSpaces)
          .map(v => {
            v.distance = turf.distance(
              turf.point(e.result.geometry.coordinates),
              turf.point(v.geometry.coordinates),
              { units: "kilometers" }
            );
            return v;
          })
          .orderBy(["distance"], ["asc"])
          .take(15) // max requests for Mapbox route API is 60 per minute
          .value();

        axios.all(this.getConcurrentRequests(closestSpaces))
          .then(
            axios.spread((...responses) => {
              this.$store.commit(
                "setDirections",
                this.findClosestSpace(responses)
              );
            })
          )
          .catch(err => {
            console.error(err);
          });
      });
    });
  },
  methods: {
    getRouteRequest(coordinates) {
      // var start = this.$store.state.closestSpace.geometry.coordinates;
      var start = this.$store.state.searchPoint[0].geometry.coordinates;
      var end = coordinates;
      var reqURL =
        "https://api.mapbox.com/directions/v5/mapbox/walking/" +
        start[0] +
        "," +
        start[1] +
        ";" +
        end[0] +
        "," +
        end[1] +
        "?geometries=geojson&access_token=" +
        mapboxgl.accessToken;
      return axios({
        url: reqURL,
        method: "get"
      });
    },
    getConcurrentRequests(openSpaces) {
      var conReqs = [];
      for (let i = 0; i < openSpaces.length; i++) {
        conReqs.push(this.getRouteRequest(openSpaces[i].geometry.coordinates));
      }
      return conReqs;
    },
    findClosestSpace(routeResponses) {
      var closestRoute = routeResponses[0].data.routes[0].geometry;
      var d = routeResponses[0].data.routes[0].distance;

      for (let i = 1; i < routeResponses.length; i++) {
        if (routeResponses[i].data.routes[0].distance < d) {
          d = routeResponses[i].data.routes[0].distance;
          closestRoute = routeResponses[i].data.routes[0].geometry;
        }
      }

      return closestRoute;
    }
  }
};
</script>

<style>
body {
  margin: 0;
  padding: 0;
}

#legend {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 240px;
  height: 80px;
  background-color: white;
  color: black;
}

#map {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 100%;
}

.mapboxgl-ctrl-geocoder {
  border: 0;
  border-radius: 0;
  position: relative;
  top: 0;
  width: 800px;
  margin-top: 0;
}

.mapboxgl-ctrl-geocoder > div {
  min-width: 100%;
  margin-left: 0;
}
</style>
