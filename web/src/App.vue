<template>
  <div class="container" style="width:100%;margin:10px;">
    <h1>{{ message }}</h1>
    <div>Total updates since <em>{{statsDate}}</em>: <strong>{{totals}}</strong></div>
    <div>
      <h2>Top 15 Updates Wikis</h2>
      <div class="chart" ref="chartdiv" style="min-height: 600px; min-width: 800px"></div>
      <h2>Top 15 Updates Wiktionaries</h2>
      <div class="chart" ref="piechartdiv" style="min-height: 600px; min-width: 800px"></div>
    </div>
  </div>
</template>

<script>
import * as am4core from "@amcharts/amcharts4/core";
import * as am4charts from "@amcharts/amcharts4/charts";
import am4themes_animated from "@amcharts/amcharts4/themes/animated";

am4core.useTheme(am4themes_animated);

export default {
  name: 'App',
  mounted() {
    let wikichart = am4core.create(this.$refs.chartdiv, am4charts.XYChart3D);
    let piechart = am4core.create(this.$refs.piechartdiv, am4charts.PieChart);

    fetch('http://localhost:8080/api/stats', {mode: 'cors'})
      .then(res => { return res.json()})
      .then(statsJSON => {
        let timestamp = statsJSON.Since
        let statsTimestamp = new Date(timestamp * 1000).toUTCString()
        this.statsDate = statsTimestamp
        let wikiCounters = statsJSON.Counters.filter(x => {
          return x.Name.startsWith("pleiades_wiki") && x.Name != "pleiades_wiki_wikidatawiki" && x.Name.endsWith("wiki")
        }).sort((x, y) => {
          if (x.Value < y.Value) return 1;
          if (y.Value < x.Value) return -1;
          return 0;
        }).slice(0,14);
        let wiktionaryCounters = statsJSON.Counters.filter(x => {
          return x.Name.startsWith("pleiades_wiki") && x.Name != "pleiades_wiki_wikidatawiki" && x.Name.endsWith("wiktionary")
        }).sort((x, y) => {
          if (x.Value < y.Value) return 1;
          if (y.Value < x.Value) return -1;
          return 0;
        }).slice(0,14);

        let describedWikiCounters = wikiCounters.map(x => {
          x.Description = x.Name.replace("pleiades_wiki_", "").replace("wiki", "");
          return x;
        });
        let describedWiktionaryCounters = wiktionaryCounters.map(x => {
          x.Description = x.Name.replace("pleiades_wiki_", "").replace("wiktionary", "");
          return x;
        });

        wikichart.data = describedWikiCounters;
        piechart.data = describedWiktionaryCounters;
        let totalCounter = statsJSON.Counters.find(x => x.Name == "pleiades_total")
        this.totals = totalCounter.Value.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
      });

    wikichart.paddingRight = 20;
    piechart.paddingRight = 20;

    wikichart.responsive.enabled = true;
    piechart.responsive.enabled = true;

    let pieSeries = piechart.series.push(new am4charts.PieSeries());
    pieSeries.dataFields.value = "Value";
    pieSeries.dataFields.category = "Description";

    let categoryAxis = wikichart.xAxes.push(new am4charts.CategoryAxis());
    categoryAxis.dataFields.category = "Description";
//    categoryAxis.title.text = "Wiki Property";

    let valueAxis = wikichart.yAxes.push(new am4charts.ValueAxis());
    valueAxis.text = "Counter"
    valueAxis.renderer.minWidth = 15;

    let series = wikichart.series.push(new am4charts.ColumnSeries3D());
    series.dataFields.valueY = "Value";
    series.dataFields.categoryX = "Description";
    series.name = "Update Events";
    series.tooltipText = "{name}: [bold]{valueY}[/]";
//    chart.legend = new am4charts.Legend();
    wikichart.cursor = new am4charts.XYCursor();
    this.wikichart = wikichart;
    this.piechart= piechart;
  },
  beforeDestroy() {
    if (this.chart) {
      this.chart.dispose();
    }
  },
  data() {
    return {
      message: 'Pleiades Stats for Great Profit!',
      totals: NaN,
      statsDate: NaN,
    };
  },
};
</script>

<style scoped>
  .container {
    width: 600px;
    margin: 50px auto;
    text-align: center;
  }
</style>
