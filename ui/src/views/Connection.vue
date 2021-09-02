<template>
  <Terminal v-bind:connectURL="wsURL"></Terminal>
</template>

<script>
import Terminal  from '@/components/Terminal';
import {BASE_WS_URL} from "@/utils/common";
export default {
  components:{
    Terminal,
  },
  name: "Connection",
  computed: {
    wsURL(){
      return this.getConnectURL()
    }
  },
  methods:{
    getConnectURL() {
      const urlParams = new URLSearchParams(window.location.search.slice(1));
      return `${BASE_WS_URL}/koko/ws/terminal/?${urlParams.toString()}`
    },
  }
}
</script>

<style scoped>

</style>