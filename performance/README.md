# Performance Stress Test

For convenience, the `test.sh` script in conjunction with the `docker-compose.yml`
file in this directory, allows running Vault with this plugin with telemetry
configured such that metrics are scraped by prometheus.

## Running

The script uses the docker-compose tool to create a few service containers.

```
./test.sh
```

After the script launches all of the necessary containers, it begins a loop
of **NUM_ITERS** iterations (defaults to 1,000).  Each iteration, a request
to the `auth/circleci/login` endpoint is made.

Once the loop completes, the containers are kept running so that `prometheus`
can continue to scrape metrics as the system winds down.

## Metrics

The metrics can be viewed at `http://localhost:9090/graph`.