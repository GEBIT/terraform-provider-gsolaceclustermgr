/**
fakeserver to mock the solace api for testing.
Code based on / inspired by https://github.com/Mastercard/terraform-provider-restapi/blob/master/fakeserver

Note that we intentionallay do NOT use generator code here!
*/

package fakeserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

/* Fakeserver represents a HTTP server with objects to hold and return*/
type Fakeserver struct {
	server  *http.Server
	objects map[string]ServiceInfo
	debug   bool
	running bool
}

type ServiceInfo struct {
	Id      string
	Name    string
	State   string
	Created time.Time
	Updated time.Time
}

/* NewFakeServer creates a HTTP server used for tests and debugging*/
func NewFakeServer(iPort int, iObjects map[string]ServiceInfo, iStart bool, iDebug bool) *Fakeserver {
	serverMux := http.NewServeMux()

	svr := &Fakeserver{
		debug:   iDebug,
		objects: iObjects,
		running: false,
	}

	serverMux.HandleFunc("/api/v2/missionControl/", svr.handleBrokerServices)
	// subtrees are also handled
	// NOTE: the trailing slash will be added automatically to the URL even when not given
	apiObjectServer := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", iPort),
		Handler: serverMux,
	}

	svr.server = apiObjectServer

	if iStart {
		svr.StartInBackground()
	}
	if svr.debug {
		log.Printf("fakeserver.go: Set up fakeserver: port=%d, debug=%t\n", iPort, svr.debug)
	}
	log.Printf("fakeserver ready")

	return svr
}

/*StartInBackground starts the HTTP server in the background*/
func (svr *Fakeserver) StartInBackground() {
	go svr.server.ListenAndServe()

	/* Let the server start */
	time.Sleep(1 * time.Second)
	svr.running = true
}

/*Shutdown closes the server*/
func (svr *Fakeserver) Shutdown() {
	svr.server.Close()
	svr.running = false
}

/*Running returns whether the server is running*/
func (svr *Fakeserver) Running() bool {
	return svr.running
}

/*GetServer returns the server object itself*/
func (svr *Fakeserver) GetServer() *http.Server {
	return svr.server
}

func (svr *Fakeserver) handleBrokerServices(w http.ResponseWriter, r *http.Request) {
	var jObj map[string]interface{}
	var sInfo ServiceInfo
	var id string
	var ok bool

	/* Assume this will never fail */
	b, _ := io.ReadAll(r.Body)

	/** we dont handle bearer token right now */

	if svr.debug {
		log.Printf("fakeserver.go: Recieved request: %+v\n", r)
		log.Printf("fakeserver.go: Headers:\n")
		for name, headers := range r.Header {
			name = strings.ToLower(name)
			for _, h := range headers {
				log.Printf("fakeserver.go:  %v: %v", name, h)
			}
		}
		log.Printf("fakeserver.go: BODY: %s\n", string(b))
	}

	path := r.URL.EscapedPath()

	parts := strings.Split(path, "/") // note: the first part is empty
	if svr.debug {
		log.Printf("fakeserver.go: Request received: %s %s\n", r.Method, path)
		log.Printf("fakeserver.go: Split request up into %d parts: %v\n", len(parts), parts)
		if r.URL.RawQuery != "" {
			log.Printf("fakeserver.go: Query string: %s\n", r.URL.RawQuery)
		}
	}

	if (len(parts) == 5 || (len(parts) == 6 && parts[5] == "")) && r.Method == "POST" {
		/* handle creation */
		sid := uuid.New().String()

		err := json.Unmarshal(b, &jObj)
		if err != nil {
			/* Failure goes back to the user as a 500. Log data here for
			   debugging (which shouldn't ever fail!) */
			log.Fatalf("fakeserver.go: Unmarshal of request failed: %s\n", err)
			log.Fatalf("\nBEGIN passed data:\n%s\nEND passed data.", string(b))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// parse and store obj
		sInfo := ServiceInfo{
			Id:      sid,
			Name:    jObj["name"].(string),
			State:   "PENDING",
			Created: time.Now(),
		}
		svr.objects[sid] = sInfo
		if svr.debug {
			log.Printf("Created Info: %v)", sInfo)
		}
		// return created obj
		result := map[string]interface{}{
			"data": map[string]interface{}{
				"id":            "O" + sInfo.Id, // the id of the operation
				"resourceId":    sInfo.Id,       // the id of the service resource
				"name":          sInfo.Name,
				"createdTime":   sInfo.Created.Format(time.RFC3339),
				"creationState": sInfo.State,
			},
			"meta": map[string]interface{}{
				"additionalProp": map[string]interface{}{},
			},
		}
		b, _ := json.Marshal(result)
		w.Header().Add("Content-Type", "json")
		w.WriteHeader(202)
		w.Write(b)
		return
	} else if len(parts) == 6 {
		// an obj was specified.
		id = parts[5]
		sInfo, ok = svr.objects[id]
		if svr.debug {
			log.Printf("fakeserver.go: Detected ID %s (exists: %t, method: %s)", id, ok, r.Method)
		}
		if !ok {
			log.Printf("Object with ID %s not found", id)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if r.Method == "GET" {
			// complete creation on first GET after ten seconds
			if sInfo.State == "PENDING" && time.Since(sInfo.Created).Seconds() > 10.0 {
				sInfo.State = "COMPLETED"
				sInfo.Updated = time.Now()
			}
			sUpdated := ""
			if sInfo.Updated.IsZero() == false {
				sUpdated = sInfo.Updated.Format(time.RFC3339)
			}
			if svr.debug {
				log.Printf("fakeserver.go: GET service %v", sInfo)
			}

			result := map[string]interface{}{
				"data": map[string]interface{}{
					"id":            sInfo.Id,
					"name":          sInfo.Name,
					"createdTime":   sInfo.Created.Format(time.RFC3339),
					"updatedTime":   sUpdated,
					"creationState": sInfo.State,
				},
				"meta": map[string]interface{}{
					"additionalProp": map[string]interface{}{},
				},
			}
			b, _ := json.Marshal(result)
			w.Header().Add("Content-Type", "json")
			w.Write(b)
			return
		} else if r.Method == "PATCH" {
			// handle update - only supported when get returns actual completed service
			sInfo.Name = jObj["Name"].(string)
			sInfo.Updated = time.Now()

			if svr.debug {
				log.Printf("fakeserver.go: PATCH service %v", sInfo)
			}
			// return what?)
			result := map[string]interface{}{
				"data": map[string]interface{}{
					"id":            sInfo.Id,
					"name":          sInfo.Name,
					"createdTime":   sInfo.Created.Format(time.RFC3339),
					"updatedTime":   sInfo.Updated.Format(time.RFC3339),
					"creationState": sInfo.State,
				},
				"meta": map[string]interface{}{
					"additionalProp": map[string]interface{}{},
				},
			}
			b, _ := json.Marshal(result)
			w.Header().Add("Content-Type", "json")
			w.Write(b)
			return
		} else if r.Method == "DELETE" {
			if svr.debug {
				log.Printf("fakeserver.go: DELETE service %v", sInfo)
			}
			// handle delete
			delete(svr.objects, id)
			// return status DELETING
			result := map[string]interface{}{
				"data": map[string]interface{}{
					"id":          "O" + sInfo.Id,
					"resourceId":  sInfo.Id,
					"name":        sInfo.Name,
					"createdTime": sInfo.Created.Format(time.RFC3339),
					"status":      "PENDING",
				},
				"meta": map[string]interface{}{
					"additionalProp": map[string]interface{}{},
				},
			}
			b, _ := json.Marshal(result)
			w.Header().Add("Content-Type", "json")
			w.WriteHeader(202)
			w.Write(b)
			return
		}

	}
	// unexpected
	if svr.debug {
		log.Printf("fakeserver.go: Bad request!")
	}
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

}
