
<template>
  <div>
    <b-jumbotron header="Welcome to Pleiades!" lead="Showing you Wikimedia Stats since the Pandemic!">
        <p>
            Pleiades is a stats aggregator that offers a glimpse into the editing activity of the world's editors of
            Wikipedia, Wiktionary and associated Wikis.
        </p>
        <p>
            Data is updated in near-real-time and available as daily aggregates since 2020-08-15.
        </p>
      <p><a class="btn btn-primary btn-lg" href="https://github.com/gargath/pleiades" role="button">Learn more &raquo;</a></p>
    </b-jumbotron>

  </div>
</template>

<style>
body {
  padding-top: 4.0rem;
}
</style>

<script>
export default {
    name: 'Home',
    mounted() {  },
};
</script>