import Vue from 'vue'
import Vuex from 'vuex'
import 'bulma/css/bulma.css'
import App from './App.vue'
import store from './store/store.js'

new Vue({
  el: '#app',
  store,
  render: h => h(App)
})
