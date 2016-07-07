package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

var (
	httpPort     = flag.Uint("http-port", 8080, "HTTP 'admin' port to listen on")
	validGhEvent = regexp.MustCompile(`^[a-z_]{1,30}$`)
)

func init() {
	flag.Parse()
}

func main() {
	httpMux := http.NewServeMux()

	httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ghEvent := r.Header.Get("X-Github-Event")

		if !validGhEvent.MatchString(ghEvent) {
			http.Error(w, "Request requires valid X-Github-Event", http.StatusBadRequest)
			return
		}

		basicHook := &BasicHook{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(basicHook)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Println(err)
			return
		}
		if basicHook.Repository.Name == "" || basicHook.Repository.Owner.Login == "" {
			http.Error(w, "Failed parsing JSON HTTP Body", http.StatusBadRequest)
			return
		}

		x := HookExec{
			Owner: basicHook.Repository.Owner.Login,
			Repo:  basicHook.Repository.Name,
			Event: ghEvent,
		}

		err = x.Exec()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Println(err)
		}
	})

	err := http.ListenAndServe(":"+strconv.Itoa(int(*httpPort)), httpMux)
	if err != nil {
		log.Fatal(err)
	}
}

// BasicHook represents the minimum body we need to parse
type BasicHook struct {
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`
}