<template>
  <div class="container" style="width:100%;margin:10px;">
    <h1>Wiktionary Update Statistics</h1>
    <DaysDropdown></DaysDropdown>
    <div>Total updates since <em>{{timestamp}}</em>: <strong>{{total}}</strong></div>
    <WikiCharts category="wiktionary"></WikiCharts>
  </div>
</template>

<script>
import { mapState } from "vuex";

import DaysDropdown from './DaysDropdown.vue';
import WikiCharts from './GraphsComponent.vue';

export default {
  name: 'WiktionaryCharts',
  components: {
    DaysDropdown,
    WikiCharts,
  },
  created() {},
  mounted() {},
  beforeDestroy() {},
  computed: mapState({
      total: state => state.total,
      timestamp: state => state.timestamp,
  }),
};
</script>

<style scoped>
  .container {
    margin: 50px auto;
    text-align: center;
  }
</style>
