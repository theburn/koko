import Vue from 'vue'
import VueRouter from 'vue-router'
import VueLogger from 'vuejs-logger'
import App from './App.vue'
import router from './router'
import loggerOptions from './plugins/logger'
import ElementUI from 'element-ui'
import 'element-ui/lib/theme-chalk/index.css';
import contextmenu from "v-contextmenu";
import "v-contextmenu/dist/index.css";
import VueCookies from 'vue-cookies'
Vue.use(VueCookies);
Vue.config.productionTip = false
Vue.use(VueRouter)
Vue.use(VueLogger, loggerOptions)
Vue.use(ElementUI)
Vue.use(contextmenu);
new Vue({
  router,
  render: h => h(App),
}).$mount('#app')
