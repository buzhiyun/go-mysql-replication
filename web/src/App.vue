<template>
  <div id="app">
    <!-- <img alt="Vue logo" src="./assets/logo.png"> -->
	<asd />
    <!-- <HelloWorld msg="Welcome to Your Vue.js App"/> -->
  </div>
</template>

<script>
// import HelloWorld from './components/HelloWorld.vue'
import asd from  './components/app.vue'

export default {
  name: 'app',
  components: {
    asd
  }
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
  /*background: #dedede;*/
}
html {
  /*background: #dedede;*/
}
</style>
