<template>
  <el-container>
    <el-main>
      <Terminal v-bind:connectURL="wsURL" v-on:session-data="onSessionData"></Terminal>
    </el-main>
    <el-aside width="60px" center v-if="enableShare">
      <el-menu background-color="#1f1b1b" text-color="#ffffff">
        <el-menu-item @click="copyShareURL">
          <i class="el-icon-share"></i>
        </el-menu-item>
      </el-menu>
    </el-aside>
  </el-container>
</template>

<script>
import Terminal from '@/components/Terminal';
import {BASE_URL, BASE_WS_URL, CopyTextToClipboard} from "@/utils/common";

export default {
  components: {
    Terminal,
  },
  name: "Connection",
  data() {
    return {
      sessionId: '',
      enableShare: false
    }
  },
  computed: {
    wsURL() {
      return this.getConnectURL()
    }
  },
  methods: {
    getConnectURL() {
      let connectURL = '';
      const routeName = this.$route.name
      switch (routeName) {
        case "Token": {
          const params = this.$route.params
          const requireParams = new URLSearchParams();
          requireParams.append('type', "token");
          requireParams.append('target_id', params.id);
          connectURL = BASE_WS_URL + "/koko/ws/token/?" + requireParams.toString()
          break
        }
        default:{
          const urlParams = new URLSearchParams(window.location.search.slice(1));
          connectURL = `${BASE_WS_URL}/koko/ws/terminal/?${urlParams.toString()}`;}
      }
      return connectURL
    },

    generateShareURL() {
      return `${BASE_URL}/koko/share/${this.sessionId}/`
    },

    copyShareURL() {
      if (!this.enableShare) {
        return
      }
      const shareURL = this.generateShareURL();
      this.$log.debug("share URL: " + shareURL)
      CopyTextToClipboard(shareURL)
      this.$message(this.$t("Terminal.CopyShareURLSuccess"))
    },

    onSessionData(session) {
      this.$log.debug("onSessionData: " , session)
      this.sessionId = session.id;
      this.enableShare = session.enable_share;
    }
  },
}
</script>

<style scoped>
</style>