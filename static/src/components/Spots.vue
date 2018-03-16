<template>
<div>
  <div id='map'></div>
  <div id="legend">
    <div>
      <svg width="40" height="20">
        <rect width="40" height="20" style="fill:#00ff00;stroke-width:3;stroke:rgb(0,0,0)" />
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
var flatbuffers = require("../../node_modules/flatbuffers").flatbuffers;
var melbourne = require("../melbourne/schema_generated.js").melbourne;

export default {
  data() {
    return {
      map: null,

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
                "#00ff00",
                "#000000"
              ]
            }
          });
        })
        .catch(err => {
          console.error(err);
        });

      var geocoder = new MapboxGeocoder({
        accessToken: mapboxgl.accessToken,
        bbox: [144.932102, -37.834726, 144.987248, -37.792422]
      });

      this.map.addControl(geocoder, "top-right");
      // Add zoom and rotation controls to the map.
      this.map.addControl(new mapboxgl.NavigationControl());
    });
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
