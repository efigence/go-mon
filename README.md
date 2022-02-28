# go-mon

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/efigence/go-mon)

Go monitoring toolbox. Basic usage:

```go
// register internal stats
mon.RegisterGcStats()

//create calculated rate via EWMA average
requestRate, _ := mon.GlobalRegistry.RegisterOrGet(`web.request_rate`, mon.NewEWMARate(time.Minute))
//create counter
requestCount, _ := mon.GlobalRegistry.RegisterOrGet(`web.request_count`, mon.NewRawCounter())

// create gauge
requestConcurrency, _ := mon.GlobalRegistry.RegisterOrGet(`web.concurrent_connections`, mon.NewRawGauge())
// second parameter for Metric type is unit
temp, _ := mon.GlobalRegistry.RegisterOrGet(`room`, mon.NewRawGauge("temperature"))

// update
// number is ignored for rate, it is treated always as single event
requestRate.Update(1)

requestCount.Update(100)
requestConcurrency.Update(20)
temp.Update(23.4)
mon.GlobalStatus.Update(mon.Ok,"all is fine")


// publish
http.Handle("/_status/health", mon.HandleHealthcheck)
http.Handle("/_status/metrics", mon.HandleMetrics)
```

If your app requires graceful stop, HAPRoxy's `http-check send-state` is also supported:

```go
// publish and handle haproxy's "X-Haproxy-Server-State"
healthcheckHandler, haproxyStatus := mon.HandleHealthchecksHaproxy()
http.Handle("/_status/health", healthcheckHandler)
// prepare to stop
mon.GlobalStatus.Update(mon.Warning,"shutting down")
// then check whether you can shutdown server gracefully (i.e. haproxy marked it as down and we have no ongoing connections)
for i := 1; i <= 10; i++ {
    if haproxyStatus.SafeToStop() {
        os.Exit(0)
    } else {
        // waiting for connections to finish
        time.Sleep(time.Second * 3)
    }
}
// and exit if it takes more than 30s as we do not expect normal requests to take that long.
os.Exit(0)

```

Returning JSON for metrics:

```json
{
  "metrics": {
    "gc.count": {
      "type": "c",
      "value": 0
    },
    "gc.cpu": {
      "type": "G",
      "unit": "percent",
      "value": 0
    },
    "gc.free": {
      "type": "c",
      "value": 636
    },
    "gc.heap_alloc": {
      "type": "G",
      "unit": "bytes",
      "value": 1561000
    },
    "gc.heap_idle": {
      "type": "G",
      "unit": "bytes",
      "value": 63897600
    },
    "gc.heap_inuse": {
      "type": "G",
      "unit": "bytes",
      "value": 2752512
    },
    "gc.heap_obj": {
      "type": "G",
      "value": 9459
    },
    "gc.malloc": {
      "type": "c",
      "value": 10095
    },
    "gc.mcache_inuse": {
      "type": "G",
      "unit": "bytes",
      "value": 13824
    },
    "gc.mspan_inuse": {
      "type": "G",
      "unit": "bytes",
      "value": 37544
    },
    "gc.pause": {
      "type": "C",
      "unit": "ns",
      "value": 0
    },
    "gc.stack_inuse": {
      "type": "G",
      "unit": "bytes",
      "value": 458752
    },
    "room": {
      "type": "G",
      "unit": "temperature",
      "value": 23.4
    },
    "web.concurrent_connections": {
      "type": "G",
      "value": 20
    },
    "web.request_count": {
      "type": "c",
      "value": 100
    },
    "web.request_rate": {
      "type": "G",
      "value": 9.962166085834647e-11
    }
  },
  "instance": "main",
  "interval": 10,
  "fqdn": "test.example.com",
  "ts": "2020-01-31T18:09:18.700941258+01:00"
```

Status:

```json
{
  "state": 1,
  "name": "foobar",
  "fqdn": "test.example.com",
  "display_name": "foobar",
  "description": "do foo to bar",
  "msg": "started",
  "ok": true,
  "ts": "2020-01-31T18:19:18.347185356+01:00",
  "components": {}
}
```



## Metrics

module by default creates global container for the metrics under `mon.GlobalRegistry` or you can create your own if you need to compartmentalize. 


### Example usage

Register GC stats
```go       
mon.RegisterGcStats()
```

Register some of our own metrics
```
mon.GlobalRegistry.Register(`web.request_rate`, mon.NewEWMARate(time.Minute))
mon.GlobalRegistry.Register(`web.request_count`, mon.NewCounter())
```

there is also MustRegister() that will panic if metric exists,
and RegisterOrGet() that will just read existing one if it was already created

read them back from registry somewhere else in code
```go
rate, _ := mon.GlobalRegistry.GetMetric(`web.request_rate`)
count, _ := mon.GlobalRegistry.GetMetric(`web.request_count`)
```
// and update
```go
rate.Update(1) 
count.Update(1)
```
note that all *Rate metrics ignore update value ; each Update() is always "one request" for rate calculation purpose

Now we can expose them under url via standard HTTP interface helper
```go
http.Handle("/_status/health", mon.HandleMetrics)
```

## Status

### How it works

Status object contains app's name, fqdn and human readable description, state, and 0 or more components (each being its own status object).

If there are no subcomponents status returns its own state.

If there is more than zero it uses merge function to return state of its components. Merge function by default chooses the worst state out of all.

States are typical nagios/Icinga states shifted by 1 to avoid 0 being interpreted as OK

* `0`- Invalid 
* `1` - OK
* `2` - Warning 
* `3` - Critical
* `4` - Unknown 

### Usage

Set app's name and description.

```go
mon.GlobalStatus.Name = "my-app"
mon.GlobalStatus.DisplayName = "My Application"
mon.GlobalStatus.Description = "My great appserver"
```

for simple apps just use global object:

```go
mon.GlobalStatus.Update(mon.Ok,"all is fine")
```

for more complex ones, add components (note: do **not** update parent object, it will get updated with worst status of the children)

```go
dbState := mon.GlobalStatus.MustNewComponent("db")
cacheState := mon.GlobalStatus.MustNewComponent("cache")
// once they started
go func () {
	for {
           if stateGood() {
		    dbState.Update(mon.Ok, "running")
           } else {
              dbState.Update(mon.Critical, "error xyz")
           }
		time.Sleep(time.Second * 10)
	}
}()
...
```

then publish the results:

```go
http.Handle("/_status/health", mon.HandleHealthcheck)
```

or use any other router that supports `func(w http.ResponseWriter, r *http.Request)`

```go
r := gin.New()
appMetricsR := r.Group("/_status")
appMetricsR.GET("/health", gin.WrapF(mon.HandleHealthcheck))
```

   

