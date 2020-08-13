import Vue from 'vue';
import Vuex from 'vuex';

import ISO6391 from 'iso-639-1';
import iso6392 from 'iso-639-2';
import iso6393 from 'iso-639-3';


Vue.use(Vuex);

export default new Vuex.Store({
    state: {
        counters: [],
        total: NaN,
        timestamp: NaN,
    },
    getters: {
        wikiStats: state => {
            let wikiCounters = state.counters.filter(x => {
                return x.Name.startsWith("pleiades_wiki") && x.Name != "pleiades_wiki_wikidatawiki" && x.Name.endsWith("wiki")
            }).sort((x, y) => {
              if (x.Value < y.Value) return 1;
              if (y.Value < x.Value) return -1;
              return 0;
            }).slice(0,14);

            let describedWikiCounters = wikiCounters.map(x => {
              let langCode = x.Name.replace("pleiades_wiki_", "").replace("wiki", "");
              if (langCode.length === 2) {
                let lang = ISO6391.getName(langCode);
                if (lang == "") {
                    x.Description = langCode;
                } else {
                    x.Description = lang;
                }
            } else if (langCode.length === 3) {
                let lang = iso6392.find(x => x.iso6392B == langCode)
                if (lang !== undefined){
                    x.Description = lang.name.split(";")[0];
                } else {
                    let lang2 = iso6393.find(x => x.iso6393 == langCode);
                    if (lang2 !== undefined) {
                        x.Description = lang2.name;
                    } else {
                        x.Description = langCode;
                    }
                }
            } else {
                x.Description = langCode;
            }
              return x;
            });
            return describedWikiCounters;
        },
        wiktionaryStats: state => {
            let wiktionaryCounters = state.counters.filter(x => {
                return x.Name.startsWith("pleiades_wiki") && x.Name != "pleiades_wiki_wikidatawiki" && x.Name.endsWith("wiktionary")
              }).sort((x, y) => {
                if (x.Value < y.Value) return 1;
                if (y.Value < x.Value) return -1;
                return 0;
              }).slice(0,14);
      
              let describedWiktionaryCounters = wiktionaryCounters.map(x => {
                let langCode = x.Name.replace("pleiades_wiki_", "").replace("wiktionary", "");
                if (langCode.length === 2) {
                    let lang = ISO6391.getName(langCode);
                    if (lang == "") {
                        x.Description = langCode;
                    } else {
                        x.Description = lang;
                    }
                } else if (langCode.length === 3) {
                    let lang = iso6392.find(x => x.iso6392B == langCode)
                    if (lang !== undefined){
                        x.Description = lang.name.split(";")[0];
                    } else {
                        let lang2 = iso6393.find(x => x.iso6393 == langCode);
                        if (lang2 !== undefined) {
                            x.Description = lang2.name;
                        } else {
                            x.Description = langCode;
                        }
                    }
                } else {
                    x.Description = langCode;
                }
                return x;
              });
            return describedWiktionaryCounters;
        },
        bigStats: state => {
            let Counters = state.counters.filter(x => {
                if (x.Name.startsWith("pleiades_") && !x.Name.startsWith("pleiades_wiki") && !x.Name.startsWith("pleiades_type")) {
                    return true;
                }
                return false;
              });
              Counters = Counters.map(c => {
                switch(c.Name) {
                    case "pleiades_length_inc":
                        c.Description = "Number of updates that increased entry length"
                        break;
                    case "pleiades_length_dec":
                        c.Description = "Number of updates that decreased entry length"
                        break;
                    case "pleiades_growth":
                        c.Description = "Total change in size of all Wiki* (MiB)"
                        c.Value = c.Value / 1048576;
                        break;
                    case "pleiades_bot":
                        c.Description = "Total change made by bot accounts"
                        break;
                    case "pleiades_minor":
                        c.Description = "Total changes marked as 'minor'"
                        break;
                    case "pleiades_total":
                        c.Description = "Grand total of all changes"
                        break;
                    default:
                        c.Description = c.Name;
                }
                return c;
              })
              return Counters.sort((x, y) => {
                if (x.Value < y.Value) return 1;
                if (y.Value < x.Value) return -1;
                return 0;
              });
        }
    },
    mutations: {
        updateCounters(state, counters) {
            state.counters = counters;
        },
        updateTimestamp(state, timestamp) {
            state.timestamp = timestamp;
        },
        updateTotal(state, total) {
            state.total = total;
        }
    },
    actions: {
        refresh({ commit }) {
//            fetch('http://localhost:8080/api/stats', {mode: 'cors'})
            fetch('/api/stats', {mode: 'cors'})
            .then(res => { return res.json()})
            .then(statsJSON => {
                let timestamp = statsJSON.Since;
                let statsTimestamp = new Date(timestamp * 1000);
                let formattedDate = statsTimestamp.toISOString().substring(0,19).replace("T", " ") + " UTC";

                let totalCounter = statsJSON.Counters.find(x => x.Name == "pleiades_total");
                commit("updateTotal", totalCounter.Value.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ","));
                commit("updateTimestamp", formattedDate);
                commit("updateCounters", statsJSON.Counters);
            });
        }
    },
});
