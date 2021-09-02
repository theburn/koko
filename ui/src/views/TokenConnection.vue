<template>
  <Terminal v-bind:connectURL="wsURL"></Terminal>
</template>

<script>
import Terminal from "@/components/Terminal";
import {BASE_WS_URL} from "@/utils/common";

export default {
  components:{
    Terminal,
  },
  name: "TokenConnection",
  computed: {
    wsURL(){
      return this.getConnectURL()
    }
  },
  methods:{
    getConnectURL() {
      const params = this.$route.params
      const requireParams = new URLSearchParams();
      requireParams.append('type', "token");
      requireParams.append('target_id', params.id);
      return BASE_WS_URL + "/koko/ws/token/?" + requireParams.toString()
    },
  }
}
</script>

<style scoped>

</style>