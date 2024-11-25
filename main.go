package main

import ( 
  "fmt"
  "net/http"
  "sync/atomic"
)

type apiConfig struct {
  fileserverHits atomic.Int32
}

func (conf *apiConfig) hitsIncMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { 
    conf.fileserverHits.Add(1)   
    w.Header().Set("Cache-Control", "no-cache")
    next.ServeHTTP(w, r)
  }) 
}

func (conf *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(200)
  w.Write([]byte(fmt.Sprintf("%d", conf.fileserverHits.Load())))
}

func (conf *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
  conf.fileserverHits.CompareAndSwap(conf.fileserverHits.Load(), 0)

  w.WriteHeader(200)
} 

func main() {
    conf := apiConfig{}

    mux := http.NewServeMux()

    appFunc := http.StripPrefix("/app/", http.FileServer(http.Dir("./app")))

    mux.Handle("/app/", conf.hitsIncMiddleware(appFunc))

    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
      w.Header().Set("Content-Type", "text/plain; charset=utf-8")
      w.WriteHeader(200)
      w.Write([]byte("Ok"))
    })

    mux.HandleFunc("/metrics", conf.metricsHandler)
    mux.HandleFunc("/reset", conf.resetHandler)

    s := &http.Server{
      Addr: ":8080",
      Handler: mux,
    }

    s.ListenAndServe()
}

