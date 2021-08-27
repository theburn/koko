import Vue from 'vue'
import VueRouter from 'vue-router'
import VueLogger from 'vuejs-logger'
import App from './App.vue'
import router from './router'
import loggerOptions from './plugins/logger'


Vue.config.productionTip = false
Vue.use(VueRouter)
Vue.use(VueLogger, loggerOptions)

new Vue({
  router,
  render: h => h(App),
}).$mount('#app')
