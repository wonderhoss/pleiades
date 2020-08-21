<template>
  <div id="app">
    <b-navbar toggleable="lg" type="dark" variant="info" class="navbar-dark fixed-top bg-dark">
      <b-navbar-brand to="/">
      <b-img rounded v-bind:src="this.logo" width="30" height="30" alt="Pleiades Logo"></b-img>
      Pleiades Wiki Stats
      </b-navbar-brand>

      <b-navbar-toggle target="nav-collapse"></b-navbar-toggle>

      <b-collapse id="nav-collapse" is-nav>
        <b-navbar-nav>
          <b-nav-item to="/bignum">Big Numbers</b-nav-item>
          <b-nav-item to="/wikipedias">Wikipedias</b-nav-item>
          <b-nav-item to="/wiktionaries">Wiktionaries</b-nav-item>
        </b-navbar-nav>

        <!-- Right aligned nav items -->
        <b-navbar-nav class="ml-auto">
        </b-navbar-nav>
      </b-collapse>
    </b-navbar>

   <router-view></router-view>

  </div>
</template>

<script>
  import logo from '../static/pleiades_logo.png';
  export default {
    name: 'App',
    data() {
      return { logo: logo };
    }
  }
</script>