import Vue from 'vue'
import Vuex from 'vuex'
import _ from 'lodash'

Vue.use(Vuex)

export default new Vuex.Store({
  state: {
    ws: null,
    wsConnected: false,
    wsProtocol: location.hostname === "localhost" ? "ws://" : "wss://",

    spots: [],
    count: {
      occupied: 0,
      unoccupied: 0
    },
    features: [], // [{ type: "Feature", geometry: geometry }]
    searchPoint: [], // [{ type: "Feature", geometry: geometry }]
    closestSpace: {}, // { type: "Feature", geometry: geometry }
    directions: [] // [{ type: "Feature", geometry: geometry }]
  },
  mutations: {
    startWS(state) {
      state.ws = new WebSocket(state.wsProtocol + location.host + "/ws")
      state.ws.binaryType = 'arraybuffer'
      state.ws.onopen = event => {
        state.wsConnected = true
      }
    },
    updateSpots(state, msg) {
      state.spots = []
      state.count = {
        occupied: 0,
        unoccupied: 0
      }

      for (let i = 0; i < msg.spotsLength(); i++) {
        if (msg.spots(i).status() === 'Present') {
          state.count.occupied++
        } else if (msg.spots(i).status() === 'Unoccupied') {
          state.count.unoccupied++
        }

        state.spots.push({
          bayId: msg.spots(i).bayId(),
          longitude: msg.spots(i).longitude(),
          latitude: msg.spots(i).latitude(),
          stMarkerId: msg.spots(i).stMarkerId(),
          status: msg.spots(i).status()
        })
      }
    },
    updateFeatures(state) {
      state.features = []
      for (let i = 0; i < state.spots.length; i++) {
        state.features.push({
          type: "Feature",
          geometry: {
            type: "Point",
            coordinates: [parseFloat(state.spots[i].longitude), parseFloat(state.spots[i].latitude)]
          },
          properties: {
            status: state.spots[i].status
          }
        })
      }
    },
    updateSearchPoint(state, geometry) {
      state.searchPoint = []
      state.searchPoint.push({
        type: "Feature",
        geometry: geometry
      })
    },
    setClosestSpace(state) {
      state.closestSpace = {}
      var options = {
        units: "kilometers"
      }
      var i
      var d
      for (i = 0; i < state.features.length; i++) {
        if (state.features[i].properties.status === "Unoccupied" && _.hasIn(state.closestSpace, "geometry") === false) {
          state.closestSpace = state.features[i]
        } else if (state.features[i].properties.status === "Unoccupied" && _.hasIn(state.closestSpace, "geometry") === true) {
          state.closestSpace = (turf.distance(turf.point(state.searchPoint[0].geometry.coordinates), turf.point(state.features[i].geometry.coordinates), options)
            < turf.distance(turf.point(state.searchPoint[0].geometry.coordinates), turf.point(state.closestSpace.geometry.coordinates), options)) ? state.features[i] : state.closestSpace
        }
      }
    },
    setDirections(state, geometry) {
      state.directions = []
      state.directions.push({
        type: "Feature",
        geometry: geometry
      })
    }
  }
})
