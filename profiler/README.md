# Profile


## Accessing the pprof endpoints
The debug endpoint can be accessed at `/debug/pprof/` and provides various profiling information such as CPU, memory, goroutine, and mutex profiles. 
These endpoints are useful for diagnosing performance issues and identifying bottlenecks in your application.

## Using the pprof visualizer
While all the data is exposed via the endpoints, the visualizer can be used to view the profiles in a more interactive friendly way.

### Explore the heap
```shell
go tool pprof -http=localhost:9090 http://localhost:8080/debug/pprof/heap
```

### Profile CPU performance
```shell
go tool pprof -http=localhost:9090 http://localhost:8080/debug/pprof/profile
```

### Run a Trace
```shell
curl -s http://localhost:8080/debug/pprof/trace > ./cpu-trace.out
go tool tace -http=localhost:9091 ./cpu-trace.out
```

> Note: tracing uses the chrome debugger for visualization the below link must be opened in the Chrome browser.
The trace will be served at http://[::]:9091
