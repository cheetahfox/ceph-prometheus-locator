# ceph-prometheus-locator


We need to figure out which server is currently serving up this output.

For example this is what ceph is doing.

```json
[
  {
    "targets": [
      "192.168.66.69:9100"
    ],
    "labels": {
      "instance": "atlas"
    }
  },
  {
    "targets": [
      "192.168.66.71:9100"
    ],
    "labels": {
      "instance": "electra"
    }
  },
  {
    "targets": [
      "192.168.66.70:9100"
    ],
    "labels": {
      "instance": "merope"
    }
  },
  {
    "targets": [
      "192.168.66.72:9100"
    ],
    "labels": {
      "instance": "taygeta"
    }
  }
]
```

We need to figure out which server is currently active and then serve up a URL that redirects to the correct server.


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
            replacement: 8b768c1e-044d-11ef-94fd-0cc47a59362a
            action: replace
          http_sd_configs:
          - follow_redirects: true
            enable_http2: true
            refresh_interval: 1m
            url: http://192.168.66.70:8765/sd/prometheus/sd-config?service=node-exporter
```