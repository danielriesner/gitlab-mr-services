package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	//"net/url"
	"os"

	gitlab "github.com/xanzy/go-gitlab"
	//"strings"
)

func main() {
	var baseURL = flag.String("url", "", "URL (e.g. http://gitlab.com)")
	var privateTokenGlobal = flag.String("private_token", "", "Authorization Token (e.g. XXxXXx0xxxXXXxXxXxxX)")
	var port = flag.Int("port", 8181, "Port")

	flag.Parse()
	if *baseURL == "" {
		printUsageAndExit("Error: --url is required")
	}
	if *privateTokenGlobal == "" {
		printWarning("Warning: --private_token is not set")
	}

	http.HandleFunc("/router", func(w http.ResponseWriter, r *http.Request) {
		queryPrivateToken := r.URL.Query().Get("private_token")
		var privateToken *string
		if queryPrivateToken != "" {
			privateToken = &queryPrivateToken
		} else {
			privateToken = privateTokenGlobal
		}
		if *privateToken == "" {
			fmt.Fprintf(os.Stderr, "Error: private_token is required\n")
		}
		var requestBody = &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			log.Printf("WARN: Failed to deserialize request body (%s)", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		requestBodyAsByteArray, _ := json.Marshal(requestBody)
		log.Printf("INFO: Received %s", string(requestBodyAsByteArray))

		git := gitlab.NewClient(nil, string([]byte(*privateToken)))
		fmt.Printf("BASE URL: %s", *baseURL)
		git.SetBaseURL(*baseURL)

		// in case we opened a new merge request
		if requestBody.ObjectAttributes.Action == "open" {

			// label it with unique label
			HandleLabel(*requestBody, git)
			log.Printf("call bomr after creation")

			// in case we merging a mergerequest
		} else if requestBody.ObjectAttributes.Action == "merge" {

			// merge linked mergerequests
			HandleMerge(*requestBody, git)

			// in case the merge request was updated
		} else if requestBody.ObjectAttributes.Action == "" {

		}
	})

	log.Printf(fmt.Sprintf("INFO: Listening on port %d", *port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
