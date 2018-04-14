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
    openSpaces: [], // [{ type: "Feature", geometry: geometry }]
    searchPoint: [], // [{ type: "Feature", geometry: geometry }]
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
      state.openSpaces = []
      for (let i = 0; i < state.spots.length; i++) {
        let f = {
          type: "Feature",
          geometry: {
            type: "Point",
            coordinates: [parseFloat(state.spots[i].longitude), parseFloat(state.spots[i].latitude)]
          },
          properties: {
            status: state.spots[i].status
          }
        }

        state.features.push(f)

        if (f.properties.status === "Unoccupied") {
          state.openSpaces.push({
            distance: 0,
            ...f
          })
        }
      }
    },
    updateSearchPoint(state, geometry) {
      state.searchPoint = []
      state.searchPoint.push({
        type: "Feature",
        geometry: geometry
      })
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