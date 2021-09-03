<template>
  <el-container :style="backgroundColor">
    <el-main>
      <Terminal ref='term' v-bind:connectURL="wsURL" v-on:ws-data="onWsData"></Terminal>
    </el-main>
    <el-aside width="60px" center>
      <el-menu :collapse="true" :background-color="themeBackground" text-color="#ffffff">
        <el-menu-item @click="dialogVisible=!dialogVisible" index="0">
          <i class="el-icon-setting"></i>
          <span slot="title">主题设置</span>
        </el-menu-item>
        <el-menu-item @click="shareDialogVisible=!shareDialogVisible" v-if="enableShare" index="1">
          <i class="el-icon-share"></i>
          <span slot="title">分享</span>
        </el-menu-item>
        <el-submenu index="2">
          <template slot="title">
            <i class="el-icon-s-custom"></i>
            <span slot="title">在线人员</span>
          </template>
          <el-menu-item-group>
            <span slot="title">人员</span>
            <el-menu-item index="2-1">选项1</el-menu-item>
            <el-menu-item index="2-2">选项2</el-menu-item>
          </el-menu-item-group>
        </el-submenu>

      </el-menu>
    </el-aside>
    <ThemeConfig :visible.sync="dialogVisible" @setTheme="handleChangeTheme"></ThemeConfig>

    <el-dialog
        :title="shareTitle"
        :visible.sync="shareDialogVisible"
        width="30%"
        :close-on-press-escape="false"
        :close-on-click-modal="false"
        @close="shareDialogClosed"
        center>
      <el-form v-if="!shareId" v-loading="loading">
        <el-form-item label="有效期">
          <el-select v-model="expiredTime" placeholder="请选择">
            <el-option
                v-for="item in expiredOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value">
            </el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <el-result v-if="shareId" icon="success" title="创建成功">
        <template slot="extra">
        </template>
      </el-result>
      <el-form  v-if="shareId" >
        <el-form-item label="地址">
          <el-input  readonly  :value="shareURL"/>
        </el-form-item>
        <el-form-item label="验证码">
          <el-input readonly  :value="shareCode"/>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
    <el-button type="primary" v-if="!shareId" @click="handleShareURlCreated">创建共享连接</el-button>
    <el-button type="primary" v-if="shareId" @click="copyShareURL">复制地址及密码</el-button>
  </span>
    </el-dialog>
  </el-container>
</template>

<script>
import Terminal from '@/components/Terminal';
import ThemeConfig from "@/components/ThemeConfig";
import {BASE_URL, BASE_WS_URL, CopyTextToClipboard} from "@/utils/common";

export default {
  components: {
    Terminal,
    ThemeConfig,
  },
  name: "Connection",
  data() {
    return {
      sessionId: '',
      enableShare: false,
      dialogVisible: false,
      themeBackground: "#1f1b1b",
      shareDialogVisible: false,
      expiredTime: 10,
      expiredOptions: [
        {label: "10m", value: 10},
        {label: "20m", value: 20},
        {label: "60m", value: 60},
      ],
      shareId: null,
      loading: false,
      shareCode: null,
      shareCode2: '#ffffff',
    }
  },
  computed: {
    wsURL() {
      return this.getConnectURL()
    },
    backgroundColor() {
      return {
        background: this.themeBackground
      }
    },
    shareTitle() {
      return this.shareId ? '分享' : '创建共享链接'
    },
    shareURL() {
      return this.shareId ? this.generateShareURL() : '无地址'
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
        default: {
          const urlParams = new URLSearchParams(window.location.search.slice(1));
          connectURL = `${BASE_WS_URL}/koko/ws/terminal/?${urlParams.toString()}`;
        }
      }
      return connectURL
    },

    generateShareURL() {
      return `${BASE_URL}/koko/share/${this.shareId}/`
    },

    copyShareURL() {
      if (!this.enableShare) {
        return
      }
      if (!this.shareId) {
        return;
      }
      const shareURL = this.generateShareURL();
      this.$log.debug("share URL: " + shareURL)
      CopyTextToClipboard(shareURL)
      this.$message(this.$t("Terminal.CopyShareURLSuccess"))
    },

    onWsData(msgType, msg) {
      switch (msg.type) {
        case "TERMINAL_SESSION": {
          const data = msg.data;
          this.sessionId = data.id;
          this.enableShare = data.enable_share;
          this.enableShare = true;
          break
        }
      }
      this.$log.debug("on ws data: ", msg)

    },
    handleChangeTheme(val) {
      if (this.$refs.term.term) {
        this.$refs.term.term.setOption("theme", val);
      }
      this.$log.debug(val);
      this.themeBackground = val.background;
    },
    handleShareURlCreated() {
      this.loading = true
      setTimeout(() => {
        this.loading = false
        this.shareId = '0000-00-00-0000'
        this.shareCode = '1234'
      }, 3000)
    },
    shareDialogClosed() {
      this.$log.debug("sharedialodClosed")
      this.loading = false;
      this.shareId = null;
      this.shareCode = null;
    }
  },
}
</script>

<style scoped>
</style>