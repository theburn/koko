<template>
  <el-container>
    <Terminal v-if="!codeDialog" ref='term' v-bind:connectURL="wsURL" v-bind:shareCode="shareCode" v-on:ws-data="onWsData"></Terminal>
    <el-dialog
        title="提示"
        :visible.sync="codeDialog"
        :close-on-press-escape="false"
        :close-on-click-modal="false"
        :show-close="false"
        width="30%">
      <el-form ref="form" label-width="80px" @submit.native.prevent>
        <el-form-item label="验证码">
          <el-input v-model="code"></el-input>
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button type="primary" @click="submitCode">确定</el-button>
      </div>
    </el-dialog>
  </el-container>
</template>

<script>
import Terminal from '@/components/Terminal'
import {BASE_WS_URL} from "@/utils/common";

export default {
  components: {
    Terminal,
  },
  name: "ShareTerminal",
  data() {
    return {
      code: '',
      codeDialog: true,
    }
  },
  computed: {
    wsURL() {
      return this.getConnectURL()
    },
    shareCode() {
      return this.code
    }
  },
  methods: {
    getConnectURL() {
      const params = this.$route.params
      const requireParams = new URLSearchParams();
      requireParams.append('type', "share");
      requireParams.append('target_id', params.id);
      return BASE_WS_URL + "/koko/ws/terminal/?" + requireParams.toString()
    },
    onWsData(msgType, msg) {
      switch (msgType) {
        case "TERMINAL_SHARE_ONLINE": {
          this.onlineShareInfo = msg.data;
          break
        }
      }
      this.$log.debug("on ws data: ", msg)
    },
    submitCode() {
      if (this.code === '') {
        this.$message("请输入验证码")
        return
      }
      this.$log.debug("code:", this.code)
      this.codeDialog = false
    }
  },

}
</script>

<style scoped>

</style>