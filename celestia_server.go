package celestia

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"strconv"
	"time"

	plasma "github.com/ethereum-optimism/optimism/op-plasma"
	"github.com/ethereum-optimism/optimism/op-service/rpc"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
)

type CelestiaServer struct {
	log        log.Logger
	endpoint   string
	store      *CelestiaStore
	tls        *rpc.ServerTLSConfig
	httpServer *http.Server
	listener   net.Listener
}

func NewCelestiaServer(host string, port int, store *CelestiaStore, log log.Logger) *CelestiaServer {
	endpoint := net.JoinHostPort(host, strconv.Itoa(port))
	return &CelestiaServer{
		log:      log,
		endpoint: endpoint,
		store:    store,
		httpServer: &http.Server{
			Addr: endpoint,
		},
	}
}

func (d *CelestiaServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/get/", d.HandleGet)
	mux.HandleFunc("/put/", d.HandlePut)
	mux.HandleFunc("/put", d.HandlePut)

	d.httpServer.Handler = mux

	listener, err := net.Listen("tcp", d.endpoint)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	d.listener = listener

	d.endpoint = listener.Addr().String()
	errCh := make(chan error, 1)
	go func() {
		if d.tls != nil {
			if err := d.httpServer.ServeTLS(d.listener, "", ""); err != nil {
				errCh <- err
			}
		} else {
			if err := d.httpServer.Serve(d.listener); err != nil {
				errCh <- err
			}
		}
	}()

	// verify that the server comes up
	tick := time.NewTimer(10 * time.Millisecond)
	defer tick.Stop()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server failed: %w", err)
	case <-tick.C:
		return nil
	}
}

func (d *CelestiaServer) HandleGet(w http.ResponseWriter, r *http.Request) {
	d.log.Debug("GET", "url", r.URL)

	route := path.Dir(r.URL.Path)
	if route != "/get" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key := path.Base(r.URL.Path)
	comm, err := hexutil.Decode(key)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input, err := d.store.Get(r.Context(), comm)
	if err != nil && errors.Is(err, plasma.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(input); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (d *CelestiaServer) HandlePut(w http.ResponseWriter, r *http.Request) {
	d.log.Debug("PUT", "url", r.URL)

	route := path.Base(r.URL.Path)
	if route != "put" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	comm, err := d.store.Put(r.Context(), input)
	if err != nil {
		key := hexutil.Encode(comm)
		d.log.Info("Failed to store commitment to the DA server", "err", err, "key", key)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(comm); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (b *CelestiaServer) Endpoint() string {
	return b.listener.Addr().String()
}

func (b *CelestiaServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = b.httpServer.Shutdown(ctx)
	return nil
}
