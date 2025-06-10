# ceph-prometheus-locator

This project is half programming practice and half solving a real problem. 

## What even is this thing?

This is an extremely opinionated way to solve the problem of using Ceph's built-in 
monitoring externally from the Ceph dashboards. This allows me scrape the data sources
that Ceph manages automatically and then store the metrics in an external Prometheus 
compatible TSDB (in my case Thanos). 

The central problem this simple service solves is locating the node that is currently
running the Ceph managed Prometheus HTTP Service Discovery. Ironically this service 
finds on which node the Prometheus service discovery is currently running. It provides 
a known and unchanging http end-point that can be configured in external systems so 
Prometheus can find and scrape the services configured by Ceph.   

## Running

This service is intended to be run in Kubernetes. It features health and readiness checks.
But this should be able to be run in any container runtime. 

## Configuration 

This service consumes a static config file that provides a list of possible locations 
where the Prometheus SD service is running. The repo includes an example which you can 
customize for your purposes. 

It also supports several environment variables to make changes to default runtime 
behavior.

| Variable          | Default | Options    |
| ----------------- | ------- | ---------- | 
| DEBUG             | False   | True/False |
| REFRESH_INTERVAL  | 60      | Seconds    |
| LOCATOR_CONFIG    | config.json | file   |


## Example Prometheus config

And the prom config looks like:

```yaml
        - job_name: ceph-node
          honor_timestamps: true
          scrape_interval: 10s
          scrape_timeout: 10s
          metrics_path: /metrics
          scheme: http
          follow_redirects: true
          enable_http2: true
          relabel_configs:
          - source_labels: [__address__]
            separator: ;
            regex: (.*)
            target_label: cluster
            replacement: 783974-BDA3422
            action: replace
          http_sd_configs:
          - follow_redirects: true
            enable_http2: true
            refresh_interval: 1m
            url: http://<DNS Or Ip Address>/sd/prometheus/sd-config?service=node-exporter
```