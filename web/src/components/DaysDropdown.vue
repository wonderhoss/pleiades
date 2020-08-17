<template>
    <div>
      <div>
        Displaying data for day {{ historic.id == 0 ? 'today' : historic.date}}
      </div>
      <b-dropdown class="days-dropdown" :text="historic.id == 0 ? 'current' : historic.date" id="days">
        <b-dropdown-item href="#" :id="0" @click="selectDay(0)">today</b-dropdown-item>
        <template v-for="day in days">
          <b-dropdown-item href="#" :id="day.id" v-bind:key="day.id" @click="selectDay(day.id)">{{day.date}}</b-dropdown-item>
        </template>
      </b-dropdown>
    </div>
</template>

<style scoped>
  .days-dropdown /deep/ .dropdown-menu {
    max-height: 250px;
    overflow-y: auto;
  }
</style>

<script>
import { mapState } from "vuex";

export default {
  name: 'DaysDropdown',
  mounted() {
    this.$root.$on('bv::dropdown::show', bvEvent => {
      this.$store.dispatch("fetchDays");
    });
  },
  computed: mapState({
      days: state => state.days,
      historic: state => {
          let dob = new Date(state.historic * 86400 * 1000);
          return {id: state.historic, date: dob.toISOString().substring(0,10)};
      },
  }),
  methods: {
    selectDay: function (day) {
        this.$store.commit("updateHistoric", day);
        this.$store.dispatch("refresh");
    }
  }
};
</script>
