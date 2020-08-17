<template>
    <div class="container" style="width:100%;margin:10px;">
    <h1>Big Update Statistics</h1>
    <DaysDropdown></DaysDropdown>
    <div>Total updates since <em>{{timestamp}}</em>: <strong>{{total}}</strong></div>
    <div>
      <table class="stats">
        <tr>
          <th>Description</th><th>Count</th>
        </tr>
        <tr v-for="item in counters" v-bind:key="item.Name">
          <td>{{ item.Description }}</td><td>{{ item.Value }}</td>
        </tr>
      </table>
    </div>
  </div>
</template>

<style scoped>
  .container {
    width: 600px;
    margin: 50px auto;
    text-align: left;
  }
  table {
    border-spacing: 5px;
  }
  table, th, td {
    border: 1px solid black;
    border-collapse: collapse;
  }
  th, td {
    padding: 5px;
  }
</style>

<script>
import { mapState } from "vuex";

import DaysDropdown from './DaysDropdown.vue';

export default {
  name: 'BignumCharts',
  components: {
    DaysDropdown,
  },
  created() {
      this.$store.dispatch("refresh");
      this.unwatch= this.$store.watch(
          (state, getters) => getters.bigStats,
          (oldValue, newValue) => {
              this.counters = this.$store.getters.bigStats;
          });
  },
  mounted() {

  },
  beforeDestroy() {
    this.unwatch();
  },
  data() {
      return {
        counters: [],
      }
  },
  computed: mapState({
      total: state => state.total,
      timestamp: state => state.timestamp,
  }),
};
</script>
