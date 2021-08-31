<template>
  <div>
    <el-row>
      <el-col :span="23">
        <div id="term"></div>
      </el-col>
      <el-col :span="1">
        <div v-contextmenu:contextmenu style="color: white">右键点击此区域</div>
      </el-col>
    </el-row>
    <el-dialog
        title="上传文件"
        :visible.sync="zmodeDialog"
        @close="dialogCloseCallback"
        center>
      <el-row>
        <el-col :span="8" :offset="4">
              <el-upload drag action="#" :auto-upload="false" :multiple="false" ref="upload"
                         :on-change="handleFileChange">
                <i class="el-icon-upload"></i>
                <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
              </el-upload>

        </el-col>
      </el-row>
      <div slot="footer" >
        <el-button @click="closeZmodemDialog">取 消</el-button>
        <el-button type="primary" @click="uploadSubmit">上传</el-button>
      </div>
    </el-dialog>
    <v-contextmenu ref="contextmenu">
      <v-contextmenu-item>菜单1</v-contextmenu-item>
      <v-contextmenu-item>菜单2</v-contextmenu-item>
      <v-contextmenu-item>菜单3</v-contextmenu-item>
    </v-contextmenu>
  </div>

</template>

<script>
import 'xterm/css/xterm.css'
import {Terminal} from 'xterm';
import {FitAddon} from 'xterm-addon-fit';
import ZmodemBrowser from "nora-zmodemjs/src/zmodem_browser";

function decodeToStr(octets) {
  if (typeof TextEncoder == "function") {
    return new TextDecoder("utf-8").decode(new Uint8Array(octets))
  }
  return decodeURIComponent(escape(String.fromCharCode.apply(null, octets)));
}

function fireEvent(e) {
  window.dispatchEvent(e)
}

function bytesHuman(bytes, precision) {
  if (!/^([-+])?|(\.\d+)(\d+(\.\d+)?|(\d+\.)|Infinity)$/.test(bytes)) {
    return '-'
  }
  if (bytes === 0) return '0';
  if (typeof precision === 'undefined') precision = 1;
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB', 'BB'];
  const num = Math.floor(Math.log(bytes) / Math.log(1024));
  const value = (bytes / Math.pow(1024, Math.floor(num))).toFixed(precision);
  return `${value} ${units[num]}`
}

const MaxTimeout = 30 * 1000
export default {
  name: "Terminal",
  data() {
    return {
      wsURL: '',
      params: '',
      term: null,
      fitAddon: null,
      ws: null,
      disable_rz_sz: false,
      pingInterval: null,
      lastReceiveTime: null,
      lastSendTime: null,
      config: null,
      zmodeDialog: false,
      zmodeSession: null,
      fileList: [],
      initialed: false,
      uploading:false,
    }
  },
  mounted: function () {
    const wsURL = this.getConnectURL();
    this.connect(wsURL)
    this.registerJMSEvent()
  },
  methods: {
    createTerminal() {
      let lineHeight = this.config.lineHeight;
      let fontSize = this.config.fontSize;
      const term = new Terminal({
        fontFamily: 'monaco, Consolas, "Lucida Console", monospace',
        lineHeight: lineHeight,
        fontSize: fontSize,
        rightClickSelectsWord: true,
        theme: {
          background: '#1f1b1b'
        }
      });
      const fitAddon = new FitAddon();
      term.loadAddon(fitAddon);
      const termRef = document.getElementById("term")
      term.open(termRef);
      fitAddon.fit();
      term.focus();
      this.fitAddon = fitAddon;
      termRef.addEventListener('resize', () => {
        this.fitAddon.fit();
        this.$log.debug("Windows resize event", term.cols, term.rows, term)
      })
      termRef.addEventListener('mouseenter', () => {
        term.focus();
        term.scrollToBottom()
      })
      return term
    },

    getConnectURL() {
      let urlParams = new URLSearchParams(window.location.search.slice(1));
      let scheme = document.location.protocol === "https:" ? "wss" : "ws";
      let port = document.location.port ? ":" + document.location.port : "";
      let baseWsUrl = scheme + "://" + document.location.hostname + port;
      let wsURL = baseWsUrl + '/koko/ws/terminal/?' + urlParams.toString();
      switch (urlParams.get("type")) {
        case 'token':
          wsURL = baseWsUrl + "/koko/ws/token/?" + urlParams.toString();
          break
        case 'shareroom':
          this.disable_rz_sz = true
          break
        default:
      }
      return wsURL
    },

    registerJMSEvent() {
      window.addEventListener('jmsFocus', evt => {
        this.$log.debug("jmsFocus ", evt);
        if (this.term) {
          this.term.focus()
          this.term.scrollToBottom()
        }
      })
    },

    connect(wsURL) {
      this.$log.debug(wsURL)
      const ws = new WebSocket(wsURL, ["JMS-KOKO"]);
      this.config = this.loadConfig();
      this.term = this.createTerminal();
      this.$log.debug(ZmodemBrowser);
      this.zsentry = new ZmodemBrowser.Sentry({
        to_terminal: (octets) => {
          if (this.zsentry && !this.zsentry.get_confirmed_session()) {
            this.term.write(decodeToStr(octets));
          }
        },
        sender: (octets) => {
          this.lastSendTime = new Date();
          this.ws.send(new Uint8Array(octets));
        },
        on_retract: () => {
          console.log('zmodem Retract')
        },
        on_detect: (detection) => {
          const zsession = detection.confirm();
          this.term.write("\r\n")
          if (zsession.type === "send") {
            this.handleSendSession(zsession);
          } else {
            this.handleReceiveSession(zsession);
          }
        }
      });

      this.term.onData(data => {
        if (!this.initialed || this.ws === null) {
          return
        }
        this.lastSendTime = new Date();
        this.$log.debug("term on data event")
        this.ws.send(this.message(this.terminalId, 'TERMINAL_DATA', data));
      });

      this.term.onResize(({cols, rows}) => {
        if (!this.initialed || this.ws === null) {
          return
        }
        this.$log.debug("send term resize ")
        this.ws.send(this.message(this.terminalId, 'TERMINAL_RESIZE', JSON.stringify({cols, rows})))
      })
      this.ws = ws;
      ws.binaryType = "arraybuffer";
      ws.onopen = this.onWebsocketOpen;
      ws.onerror = this.onWebsocketErr;
      ws.onclose = this.onWebsocketClose;
      ws.onmessage = this.onWebsocketMessage;
    },

    onWebsocketMessage(e) {
      this.lastReceiveTime = new Date();
      if (typeof e.data === 'object') {
        this.zsentry.consume(e.data);
      } else {
        this.dispatch(e.data);
      }
    },

    onWebsocketOpen() {
      if (this.pingInterval !== null) {
        clearInterval(this.pingInterval);
      }
      this.lastReceiveTime = new Date();
      this.pingInterval = setInterval(() => {
        if (this.ws.readyState === WebSocket.CLOSING ||
            this.ws.readyState === WebSocket.CLOSED) {
          clearInterval(this.pingInterval)
          return
        }
        let currentDate = new Date();
        if ((this.lastReceiveTime - currentDate) > MaxTimeout) {
          this.$log.debug("more than 30s do not receive data")
        }
        let pingTimeout = (currentDate - this.lastSendTime) - MaxTimeout
        if (pingTimeout < 0) {
          return;
        }
        this.ws.send(this.message(this.terminalId, 'PING', ""));
      }, 25 * 1000);
    },

    onWebsocketErr(e){
      this.term.writeln("Connection websocket error");
      fireEvent(new Event("CLOSE", {}))
      this.handleError(e)
    },

    onWebsocketClose(e){
      this.term.writeln("Connection websocket closed");
      fireEvent(new Event("CLOSE", {}))
      this.handleError(e)
    },

    dispatch(data) {
      if (data === undefined) {
        return
      }
      let msg = JSON.parse(data)
      switch (msg.type) {
        case 'CONNECT': {
          this.terminalId = msg.id;
          this.fitAddon.fit();
          const cols = this.term.cols;
          const rows = this.term.rows;
          this.ws.send(this.message(this.terminalId, 'TERMINAL_INIT',
              JSON.stringify({cols, rows})));
          this.initialed = true;
          break
        }
        case "CLOSE":
          this.term.writeln("Receive Connection closed");
          fireEvent(new Event("CLOSE", {}))
          break
        case "PING":
          break
        default:
          console.log(data)
      }
    },

    message(id, type, data) {
      return JSON.stringify({
        id,
        type,
        data,
      });
    },

    handleError(e) {
      console.log(e)
    },

    loadLunaConfig() {
      let config = {};
      let fontSize = 14;
      let quickPaste = "0"
      // localStorage.getItem default null
      let localSettings = localStorage.getItem('LunaSetting')
      if (localSettings !== null) {
        let settings = JSON.parse(localSettings)
        fontSize = settings['fontSize']
        quickPaste = settings['quickPaste']
      }
      if (!fontSize || fontSize < 5 || fontSize > 50) {
        fontSize = 13;
      }
      config['fontSize'] = fontSize;
      config['quickPaste'] = quickPaste;
      return config
    },

    loadConfig() {
      const config = this.loadLunaConfig();
      const ua = navigator.userAgent.toLowerCase();
      let lineHeight = 1;
      if (ua.indexOf('windows') !== -1) {
        lineHeight = 1.2;
      }
      config['lineHeight'] = lineHeight
      return config
    },

    handleFileChange(file, fileList) {
      if (fileList.length > 1) {
        fileList.shift()
      }
      const filesObj = fileList.map(el => el.raw);
      this.$log.debug(filesObj);

      this.$log.debug(file, fileList)
      this.fileList = fileList
    },

    handleReceiveSession(zsession) {
      zsession.on('offer', xfer => {
        const on_form_submit = () => {
          // 开始下载
          const buffer = [];
          xfer.on('input', payload => {
            // 下载中
            this.updateReceiveProgress(xfer);
            buffer.push(new Uint8Array(payload));
          });
          xfer.accept().then(() => {
            this.saveToDisk(xfer, buffer);
          }, console.error.bind(console));
        };

        on_form_submit();
      });
      zsession.on('session_end', () => {
        this.term.write('\r\n')
      });
      zsession.start();
    },

    saveToDisk(xfer, buffer) {
      ZmodemBrowser.Browser.save_to_disk(buffer, xfer.get_details().name);
    },
    updateReceiveProgress(xfer) {
      let detail = xfer.get_details();
      let name = detail.name;
      let total = detail.size;
      let offset = xfer.get_offset();
      let percent;
      if (total === 0 || total === offset) {
        percent = 100
      } else {
        percent = Math.round(offset / total * 100);
      }
      let msg = 'download ' + name + ": " + bytesHuman(total) + " " + percent + "%"
      this.term.write("\r" + msg);
    },
    updateSendProgress(xfer, percent) {
      let detail = xfer.get_details();
      let name = detail.name;
      let total = detail.size;
       percent = Math.round(percent);
      let msg = 'upload ' +  name + ": " + bytesHuman(total) + " " + percent + "%"
      this.term.write("\r" + msg);
    },
    handleSendSession(zsession) {
      this.zmodeSession = zsession;
      this.openZmodemDialog()

      zsession.on('session_end', () => {
        this.zmodeSession = null;
        this.fileList = [];
        this.term.write('\r\n')
      });
    },

    uploadSubmit() {
      this.uploading = true;
      this.closeZmodemDialog()
      if (!this.zmodeSession) {
        return
      }
      const filesObj = this.fileList.map(el => el.raw);
      this.$log.debug("Zomdem submit file: ", filesObj)
      ZmodemBrowser.Browser.send_files(this.zmodeSession, filesObj,
          {
            on_offer_response: (obj, xfer) => {
              if (xfer) {
                xfer.on('send_progress', (percent) => {
                  this.updateSendProgress(xfer, percent)
                });
              }
            },
            on_file_complete(obj) {
              console.log("COMPLETE", obj);
            },
          }
      ).then(
          this.zmodeSession.close.bind(this.zmodeSession),
          console.error.bind(console)
      ).catch(err => {
        console.log(err)
      });
    },

    dialogCloseCallback() {
      if (this.zmodeSession && !this.uploading) {
        this.$log.debug("dialog close callback zmodeSession abort")
        this.zmodeSession.abort();
      }
      this.$refs.upload.clearFiles();
      this.uploading = false;
    },
    openZmodemDialog() {
      this.zmodeDialog = true;
    },
    closeZmodemDialog() {
      this.zmodeDialog = false;
    }
  }
}
</script>

<style scoped>
@import '../assets/styles/index.css';

div {
  height: 100%;
}

#term {
  height: 100%;
  overflow: auto;
}
</style>