import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

export default new Vuex.Store({
  state: {
    ws: null,
    wsConnected: false,

    spots: [],
    features: []
  },
  mutations: {
    startWS(state) {
      state.ws = new WebSocket("ws://" + location.host + "/ws")
      state.ws.binaryType = 'arraybuffer'
      state.ws.onopen = event => {
        state.wsConnected = true
      }
    },
    updateSpots(state, msg) {
      state.spots = []
      for (let i = 0; i < msg.spotsLength(); i++) {
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
    }
  }
})
