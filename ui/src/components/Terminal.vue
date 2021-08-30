<template>
  <div id="term"></div>
</template>

<script>
import 'xterm/css/xterm.css'
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import ZmodemBrowser from "nora-zmodemjs/src/zmodem_browser";

function decodeToStr(octets) {
  if (typeof TextEncoder == "function") {
    return new TextDecoder("utf-8").decode(new Uint8Array(octets))
  }
  return decodeURIComponent(escape(String.fromCharCode.apply(null, octets)));
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
      config:null,
    }
  },
  mounted: function () {
    const wsURL = this.getConnectURL();
    this.config = this.loadConfig();
    this.connect(wsURL)
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
      window.onresize = () => {
        this.fitAddon.fit();
        this.$log.debug("Windows resize event", term.cols, term.rows, term)
      }
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

    connect(wsURL) {
      const ws = new WebSocket(wsURL, ["JMS-KOKO"]);
      this.$log.debug(wsURL)
      this.term = this.createTerminal("terminal");
      window.addEventListener('jmsFocus', evt => {
        this.$log.debug(evt);
        this.term.focus()
        this.term.scrollToBottom()
      })
      this.$log.debug(ZmodemBrowser);
      const zsentry = new ZmodemBrowser.Sentry({
        to_terminal: (octets) => {
          if (!zsentry.get_confirmed_session()) {
            this.term.write(decodeToStr(octets));
          }
        },
        sender: (octets) => {
          this.ws.send(new Uint8Array(octets));
        },
        on_retract: () => {
          console.log('zmodem Retract')
        },
        on_detect: (detection) => {
          var promise;
          let file_input_el;
          var zsession = detection.confirm();
          this.term.write("\r\n")
          if (zsession.type === "send") {
            // 动态创建 input 标签，否则选择相同的文件，不会触发 onchang 事件
            file_input_el = document.createElement("input");
            file_input_el.type = "file";
            file_input_el.style.display = "none";//隐藏
            document.body.appendChild(file_input_el);
            document.body.onfocus = function () {
              document.body.onfocus = null;
              setTimeout(function () {
                // 如果未选择任何文件，则代表取消上传。主动取消
                if (file_input_el.files.length === 0) {
                  console.log("Cancel file clicked")
                  if (!zsession.aborted()) {
                    zsession.abort()
                  }
                }
              }, 1000);
            }
            promise = this._handle_send_session(file_input_el, zsession);
          } else {
            promise = this._handle_receive_session(zsession);
          }
          promise.catch(console.error.bind(console)).then(() => {
            console.log("zmodem Detect promise finished")
          }).finally(() => {
            if (file_input_el != null) {
              document.body.removeChild(file_input_el);
            }
          })

        }
      });

      this.term.onData(data => {
        if (this.initialed === null || this.ws === null) {
          return
        }
        this.lastSendTime = new Date();
        this.$log.debug("term on data event")
        this.ws.send(this.message(this.terminalId, 'TERMINAL_DATA', data));
      });

      this.term.onResize(({cols, rows}) => {
        if (this.initialed === null || this.ws === null) {
          return
        }
        this.$log.debug("send term resize ")
        this.ws.send(this.message(this.terminalId, 'TERMINAL_RESIZE', JSON.stringify({cols, rows})))
      })
      this.ws = ws;
      ws.binaryType = "arraybuffer";
      ws.onopen = () => {
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
            console.log("more than 30s do not receive data")
          }
          let pingTimeout = (currentDate - this.lastSendTime) - MaxTimeout
          if (pingTimeout < 0) {
            return;
          }
          this.ws.send(this.message(this.terminalId, 'PING', ""));
        }, 25 * 1000);
      }
      ws.onerror = (e) => {
        this.term.writeln("Connection websocket error");
        this.fireEvent(new Event("CLOSE", {}))
        this.handleError(e)
      }
      ws.onclose = (e) => {
        this.term.writeln("Connection websocket closed");
        this.fireEvent(new Event("CLOSE", {}))
        this.handleError(e)
      }
      ws.onmessage = (e) => {
        this.lastReceiveTime = new Date();
        if (typeof e.data === 'object') {
          zsentry.consume(e.data);
        } else {
          this.dispatch(this.term, e.data);
        }
      }
    },

    dispatch(term, data) {
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
          this.fireEvent(new Event("CLOSE", {}))
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

    fireEvent(e) {
      window.dispatchEvent(e)
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
    }
  }
}
</script>

<style scoped>
@import '../assets/styles/index.css';

#term {
  height: 100%;
  overflow: auto;
}
</style>