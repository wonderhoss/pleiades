<template>
    <div>
      <h2>Top 15 Updated {{ this.category }} Edits</h2>
      <div class="chart" ref="chartdiv" style="min-height:500px;"></div>
      <div class="chart" ref="piechartdiv" style="min-height:500px;"></div>
    </div>
</template>

<script>
import * as am4core from "@amcharts/amcharts4/core";
import * as am4charts from "@amcharts/amcharts4/charts";
import am4themes_animated from "@amcharts/amcharts4/themes/animated";

am4core.useTheme(am4themes_animated);

export default {
  name: 'WikiCharts',
  props: ['category'],
  components: {
  },
  created() {
      this.$store.dispatch("refresh");
      switch(this.category) {
          case "wikipedia":
            this.unwatch= this.$store.watch(
                (state, getters) => getters.wikiyStats,
                (oldValue, newValue) => {
                    this.wikichart.data = this.$store.getters.wikiStats;
                    this.piechart.data = this.$store.getters.wikiStats;
                });
              break;
          case "wiktionary":
            this.unwatch= this.$store.watch(
                (state, getters) => getters.wiktionaryStats,
                (oldValue, newValue) => {
                    this.wikichart.data = this.$store.getters.wiktionaryStats;
                    this.piechart.data = this.$store.getters.wiktionaryStats;
                });
              break;
      }
  },
  mounted() {
    let wikichart = am4core.create(this.$refs.chartdiv, am4charts.XYChart3D);
    let piechart = am4core.create(this.$refs.piechartdiv, am4charts.PieChart);

    switch(this.category) {
        case "wikipedia":
            wikichart.data = this.$store.getters.wikiStats;
            piechart.data = this.$store.getters.wikiStats;
            break;
        case "wiktionary":
            wikichart.data = this.$store.getters.wiktionaryStats;
            piechart.data = this.$store.getters.wiktionaryStats;
            break;
    }

    wikichart.paddingRight = 20;
    piechart.paddingRight = 20;

    wikichart.responsive.enabled = true;
    piechart.responsive.enabled = true;

    let pieSeries = piechart.series.push(new am4charts.PieSeries());
    pieSeries.dataFields.value = "Value";
    pieSeries.dataFields.category = "Description";

    let categoryAxis = wikichart.xAxes.push(new am4charts.CategoryAxis());
    categoryAxis.dataFields.category = "Description";

    let valueAxis = wikichart.yAxes.push(new am4charts.ValueAxis());
    valueAxis.text = "Counter"
    valueAxis.renderer.minWidth = 15;

    let series = wikichart.series.push(new am4charts.ColumnSeries3D());
    series.dataFields.valueY = "Value";
    series.dataFields.categoryX = "Description";
    series.name = "Update Events";
    series.tooltipText = "{name}: [bold]{valueY}[/]";
    wikichart.cursor = new am4charts.XYCursor();
    this.wikichart = wikichart;
    this.piechart= piechart;
  },
  beforeDestroy() {
    this.unwatch();
    if (this.piechart) {
      this.piechart.dispose();
    }
    if (this.wikichart) {
      this.wikichart.dispose();
    }
  },
};
</script>

<style scoped>
</style>
