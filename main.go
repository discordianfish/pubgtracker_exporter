package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	pubg "github.com/albshin/go-pubg"
	"github.com/fatih/camelcase"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// If you fork this, please create a new api key
	apiKey = "3eacefb9-dde6-4c7a-a3ef-ddeffffb035f"

	namespace = "pubgtracker"
	subsystem = "stats"

	metricsPath = "/metrics/"
)

var (
	exportedFields = []string{
		"TimeSurvived",
		"RoundsPlayed",
		"Wins",
		"Top10s",
		"Losses",
		"Rating",
		"BestRating",
		"BestRank",
		"Kills",
		"Assists",
		"Suicides",
		"TeamKills",
		"HeadshotKills",
		"VehicleDestroys",
		"RoadKills",
		"DailyKills",
		"WeeklyKills",
		"RoundMostKills",
		"MaxKillStreaks",
		"WeaponAcquired",
		"Days",
		"LongestTimeSurvived",
		"MostSurvivalTime",
		"AvgSurvivalTime",
		"WinPoints",
		"WalkDistance",
		"RideDistance",
		"MoveDistance",
		"AvgWalkDistance",
		"AvgRideDistance",
		"LongestKill",
		"Heals",
		"Revives",
		"Boosts",
		"DamageDealt",
		"DBNOs",
	}
)

func handler(w http.ResponseWriter, r *http.Request, client *pubg.API) {
	var (
		start    = time.Now()
		registry = prometheus.NewRegistry()

		durationGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "duration_seconds",
			Help:      "Duration of exporting pubgtracker stats",
		})
		errorCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Total number of errrors encounterd while exporting pubgtracker stats",
		})
	)
	player := r.URL.Path[len(metricsPath):]
	if player == "" {
		http.Error(w, "No player given", http.StatusBadRequest)
		return
	}
	info, err := client.GetPlayer(player)
	if err != nil {
		errorCounter.Inc()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	registry.MustRegister(durationGauge)
	registry.MustRegister(errorCounter)
	registry.MustRegister(&statsCollector{info})

	durationGauge.Set(time.Since(start).Seconds())
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

type statsCollector struct {
	*pubg.Player
}

func formatField(f string) string {
	return strings.ToLower(strings.Join(camelcase.Split(f), "_"))
}

func (c *statsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, region := range c.Player.Stats {
		for _, stats := range region.Stats {
			if !in(stats.Field, exportedFields) {
				continue
			}
			ch <- prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem, formatField(stats.Field)),
				stats.Label,
				[]string{"region", "match", "season"}, nil,
			)
		}
	}
}

func (c *statsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, region := range c.Player.Stats {
		if region.Region == "agg" {
			continue
		}
		for _, stats := range region.Stats {
			if !in(stats.Field, exportedFields) {
				continue
			}
			desc := prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem, formatField(stats.Field)),
				stats.Label,
				[]string{"region", "match", "season"}, nil,
			)
			ch <- prometheus.MustNewConstMetric(
				desc,
				prometheus.UntypedValue,
				stats.ValueDec,
				region.Region,
				region.Match,
				region.Season)
		}
	}
}

func in(search string, list []string) bool {
	for _, e := range list {
		if e == search {
			return true
		}
	}
	return false
}

func main() {
	var (
		listenAddress = flag.String("l", ":8080", "Address to listen on")
	)
	flag.Parse()
	client, err := pubg.New(apiKey)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc(metricsPath, func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, client)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
    			<head><title>pubgtracker exporter</title></head>
    			<body>
    				<h1>pubtracker exporter</h1>
    				<p>Use <i>/metrics/username</i> to get stats for <i>username</i>.
    				<p><a href="/metrics/discordianfish">Example stats for discordianfish</a></p>
    			</body>
    			</html>`))
	})

	log.Println("Listening on", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %s", err)
	}
}
