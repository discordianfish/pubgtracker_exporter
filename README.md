# Pubgtracker Exporter
Prometheus Exporter exposing [pubtracker](https://pubgtracker.com) stats.

It exposes the raw stats from the pubgtracker API, therefor the metrics are
untyped and do not follow the Prometheus best practices.


## Usage
```
$ ./pubgtracker_exporter &
$ curl -s localhost:8080/metrics/discordianfish | grep -i pubgtracker_stats_longest_kill
# HELP pubgtracker_stats_longest_kill Longest Kill
# TYPE pubgtracker_stats_longest_kill untyped
pubgtracker_stats_longest_kill{match="duo",region="eu",season="2017-pre2"}
342.54
pubgtracker_stats_longest_kill{match="duo",region="eu",season="2017-pre4"}
184.39
pubgtracker_stats_longest_kill{match="solo",region="eu",season="2017-pre2"}
219.08
pubgtracker_stats_longest_kill{match="solo",region="eu",season="2017-pre4"}
190.7
pubgtracker_stats_longest_kill{match="squad",region="eu",season="2017-pre2"}
367.31
pubgtracker_stats_longest_kill{match="squad",region="eu",season="2017-pre4"}
1.41
```
