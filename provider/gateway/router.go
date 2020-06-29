package gateway

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/gorilla/mux"

	"github.com/ovrclk/akash/provider"
	"github.com/ovrclk/akash/provider/cluster"
	"github.com/ovrclk/akash/provider/manifest"
	mtypes "github.com/ovrclk/akash/x/market/types"
)

const (
	contentTypeJSON = "application/json; charset=UTF-8"

	// Time allowed to write the file to the client.
	pingWait = 15 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 15 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = 10 * time.Second
)

const (
	// as per RFC https://www.iana.org/assignments/websocket/websocket.xhtml#close-code-number
	// errors from private use staring
	websocketInternalServerErrorCode = 4000
)

type wsLogsConfig struct {
	lid       mtypes.LeaseID
	service   string
	follow    bool
	tailLines *int64
	log       log.Logger
	client    cluster.ReadClient
}

func newRouter(log log.Logger, pclient provider.Client) *mux.Router {
	router := mux.NewRouter()

	// GET /status
	router.HandleFunc("/status",
		createStatusHandler(log, pclient)).
		Methods("GET")

	// PUT /deployment/<deployment-id>/manifest
	drouter := router.PathPrefix(deploymentPathPrefix).Subrouter()
	drouter.Use(requireDeploymentID(log))
	drouter.HandleFunc("/manifest",
		createManifestHandler(log, pclient.Manifest())).
		Methods("PUT")

	lrouter := router.PathPrefix(leasePathPrefix).Subrouter()
	lrouter.Use(requireLeaseID(log))

	// GET /lease/<lease-id>/status
	lrouter.HandleFunc("/status",
		leaseStatusHandler(log, pclient.Cluster())).
		Methods("GET")

	// GET /lease/<lease-id>/service/<service-name>/status
	lrouter.HandleFunc("/service/{serviceName}/status",
		leaseServiceStatusHandler(log, pclient.Cluster())).
		Methods("GET")

	logRouter := lrouter.PathPrefix("/service/{serviceName}/logs").Subrouter()
	logRouter.Use(requestLogParams())

	// GET /lease/<lease-id>/service/<service-name>/logs
	logRouter.HandleFunc("",
		leaseServiceLogsHandler(log, pclient.Cluster())).
		Methods("GET")

	return router
}

func createStatusHandler(log log.Logger, sclient provider.StatusClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		status, err := sclient.Status(req.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(log, w, status)
	}
}

func createManifestHandler(_ log.Logger, mclient manifest.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var mreq manifest.SubmitRequest

		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()

		if err := decoder.Decode(&mreq); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		if !requestDeploymentID(req).Equals(mreq.Deployment) {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		if err := mclient.Submit(req.Context(), &mreq); err != nil {
			// TODO: surface unauthorized, etc...
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func leaseStatusHandler(log log.Logger, cclient cluster.ReadClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		status, err := cclient.LeaseStatus(req.Context(), requestLeaseID(req))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(log, w, status)
	}
}

func leaseServiceStatusHandler(log log.Logger, cclient cluster.ReadClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		service := mux.Vars(req)["serviceName"]
		if service == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		status, err := cclient.ServiceStatus(req.Context(), requestLeaseID(req), service)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(log, w, status)
	}
}

func leaseServiceLogsHandler(log log.Logger, cclient cluster.ReadClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		service := mux.Vars(r)["serviceName"]
		if service == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		go wsLogWriter(ws, wsLogsConfig{
			lid:       requestLeaseID(r),
			service:   service,
			follow:    requestLogFollow(r),
			tailLines: requestLogTailLines(r),
			log:       log,
			client:    cclient,
		})
	}
}

func wsLogWriter(ws *websocket.Conn, cfg wsLogsConfig) {
	pingTicker := time.NewTicker(pingPeriod)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		pingTicker.Stop()
		cancel()
		_ = ws.Close()
	}()

	logs, err := cfg.client.ServiceLogs(ctx, cfg.lid, cfg.service, cfg.follow, cfg.tailLines)
	if err != nil {
		cfg.log.Error("couldn't fetch logs: %s", err.Error())
		err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocketInternalServerErrorCode, err.Error()))
		if err != nil {
			cfg.log.Error("couldn't push control message through websocket: %s", err.Error())
		}
		return
	}

	if len(logs) == 0 {
		_ = ws.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocketInternalServerErrorCode, "service "+cfg.service+" does not have running pods"))
		return
	}

	_ = ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		_ = ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	go func() {
		for {
			if _, _, e := ws.ReadMessage(); e != nil {
				if _, ok := e.(*websocket.CloseError); !ok {
					_ = ws.Close()
				}
				break
			}
		}
	}()

	var scanners sync.WaitGroup
	var scannersDone sync.WaitGroup
	var scannersStarted sync.WaitGroup

	logCh := make(chan cluster.ServiceLogMessage)

	scanners.Add(len(logs))
	scannersStarted.Add(len(logs))

	for _, lg := range logs {
		go func(name string, scan *bufio.Scanner) {
			defer scanners.Done()
			scannersStarted.Done()

			for scan.Scan() {
				logCh <- cluster.ServiceLogMessage{
					Name:    name,
					Message: scan.Text(),
				}
			}
		}(lg.Name, lg.Scanner)
	}

	scannersStarted.Wait()

	scannersDone.Add(1)
	go func() {
		defer scannersDone.Done()
		scanners.Wait()
		_ = ws.Close()
	}()

	alive := true

	for alive {
		select {
		case line := <-logCh:
			stream, err := json.Marshal(line)
			if err != nil {
				alive = false
				break
			}

			if err = ws.WriteMessage(websocket.TextMessage, stream); err != nil {
				alive = false
				break
			}
		case <-pingTicker.C:
			if err = ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				alive = false
				break
			}
		}
	}

	for i := range logs {
		_ = logs[i].Stream.Close()
	}

	scannersDone.Wait()

	// drain logs channel in separate goroutine to unblock seeders waiting for write space
	scannersStarted.Add(1)
	go func() {
		defer scannersStarted.Done()
		for range logCh {
		}
	}()

	close(logCh)
	scannersStarted.Wait()
}

func writeJSON(log log.Logger, w http.ResponseWriter, obj interface{}) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)

	_, err = w.Write(bytes)
	if err != nil {
		log.Error("error writing response", "err", err)
		return
	}
}
