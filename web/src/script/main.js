  
import Vue from 'vue'
import { BootstrapVue, IconsPlugin } from 'bootstrap-vue'
import VueRouter from 'vue-router'

import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'

// Install BootstrapVue
Vue.use(BootstrapVue)
// Optionally install the BootstrapVue icon components plugin
Vue.use(IconsPlugin)
// Use Router
Vue.use(VueRouter)

import store from "./store";

import App from '../components/App'
import Home from '../components/Home.vue'
import BignumCharts from '../components/BigNumCharts.vue'
import WikipediaCharts from '../components/WikipediaCharts.vue'
import WiktionaryCharts from '../components/WiktionaryCharts.vue'

const routes = [
  { path: '/', component: Home },
  { path: '/bignum', component: BignumCharts},
  { path: '/wikipedias', component: WikipediaCharts},
  { path: '/wiktionaries', component: WiktionaryCharts},
  //  { path: '/profile', component: Profile }
]

const router = new VueRouter({
  routes // short for `routes: routes`
});

const app = new Vue({
  store,
  router,
  el: '#app',
  template: '<App/>',
  components: {
    App,
    VueRouter
  }
})
