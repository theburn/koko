<template>
  <div id="terminal"></div>
</template>

<script>
import 'xterm/css/xterm.css'
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import ZmodemBrowser from 'nora-zmodemjs/src/zmodem_browser';

export default {
  name: "Terminal",
  data() {
    return {
      wsURL: '',
      params: '',
      term: null,
      ws: null,
      disable_rz_sz: false
    }
  },
  mounted: function () {
    const wsURL = this.getConnectURL();
    this.$log.debug(wsURL);
    this.connect(wsURL)
  },
  methods: {
    createTerminal(elementId) {
      const termRef = document.getElementById(elementId)
      const ua = navigator.userAgent.toLowerCase();
      let lineHeight = 1;
      if (ua.indexOf('windows') !== -1) {
        lineHeight = 1.2;
      }
      let fontSize = this.getFontSize();
      let term = new Terminal({
        fontFamily: 'monaco, Consolas, "Lucida Console", monospace',
        lineHeight: lineHeight,
        fontSize: fontSize,
        rightClickSelectsWord: true,
        theme: {
          background: '#1f1b1b'
        }
      });
      term.open(termRef);
      term.focus();
      const fitAddon = new FitAddon();
      term.loadAddon(fitAddon);
      fitAddon.fit();
      window.onresize =()=> {
        this.$log.debug(window.innerHeight, window.innerWidth)
        fitAddon.fit();
      }
      term.attachCustomKeyEventHandler(function (e) {
        if (e.ctrlKey && e.key === 'c' && term.hasSelection()) {
          return false;
        }
        return !(e.ctrlKey && e.key === 'v');
      });
      return term
    },
    getFontSize() {
      let fontSize = 14
      // localStorage.getItem default null
      let localSettings = localStorage.getItem('LunaSetting')
      if (localSettings !== null) {
        let settings = JSON.parse(localSettings)
        fontSize = settings['fontSize']
      }
      if (!fontSize || fontSize < 5 || fontSize > 50) {
        fontSize = 13;
      }
      return fontSize
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
      // let ws = new WebSocket(wsURL, ["JMS-KOKO"]);
      this.$log.debug(wsURL)
      this.term = this.createTerminal("terminal");
      this.term.onData(data => {
        this.$log.debug(data);
      })

      window.addEventListener('jmsFocus', evt => {
        this.$log.debug(evt);
        this.term.focus()
      })
      this.$log.debug(ZmodemBrowser);
    },
    resizeTerminal() {
      // 延迟调整窗口大小
      if (this.resizeTimer != null) {
        clearTimeout(this.resizeTimer);
      }
      if (this.ws == null) {
        return;
      }
      this.resizeTimer = setTimeout(function () {
        const termRef = document.getElementById('terminal')
        termRef.style.height = (window.innerHeight - 16) + 'px';
        this.term.fit();
        this.term.focus();

      }, 500);
    }
  }
}
</script>

<style scoped>


#terminal {
  background-color: #1f1b1b;
  overflow: auto;
}
</style>