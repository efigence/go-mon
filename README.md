# go-mon

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/efigence/go-mon)

Go monitoring toolbox

example usage:

```go       
 // register GC stats into mon.GlobalRegistry (created by default)
mon.RegisterGcStats()
// register some of our own metrics
mon.GlobalRegistry.Register(`web.request_rate`, mon.NewEWMARate(time.Minute))
mon.GlobalRegistry.Register(`web.request_count`, mon.NewCounter())

// there is also MustRegister() that will panic if metric exists,
// and RegisterOrGet() that will just read existing one if it was already created

// read them back from registry somewhere else in code
rate, _ := mon.GlobalRegistry.GetMetric(`web.request_rate`)
count, _ := mon.GlobalRegistry.GetMetric(`web.request_count`)

// and update
rate.Update(1) // note that all *Rate metrics ignore update value ; each Update() is always "one request" for rate calculation purpose
count.Update(1)


// then expose it under an url via standard HTTP interface helper
func (r *Renderer) HandleMetrics( w http.ResponseWriter, req *http.Request) {
     mon.HandleMetrics(w, req)
}