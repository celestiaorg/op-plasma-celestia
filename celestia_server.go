package celestia

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	s3 "github.com/celestiaorg/op-plasma-celestia/s3"
	plasma "github.com/ethereum-optimism/optimism/op-plasma"
	"github.com/ethereum-optimism/optimism/op-service/rpc"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

type CelestiaServer struct {
	log        log.Logger
	endpoint   string
	store      *CelestiaStore
	s3Store    *s3.S3Store
	tls        *rpc.ServerTLSConfig
	httpServer *http.Server
	listener   net.Listener

	cache        bool
	fallback     bool
	cacheLock    sync.RWMutex
	fallbackLock sync.RWMutex
}

func NewCelestiaServer(host string, port int, store *CelestiaStore, s3Store *s3.S3Store, fallback bool, cache bool, log log.Logger) *CelestiaServer {
	endpoint := net.JoinHostPort(host, strconv.Itoa(port))
	return &CelestiaServer{
		log:      log,
		endpoint: endpoint,
		store:    store,
		s3Store:  s3Store,
		httpServer: &http.Server{
			Addr: endpoint,
		},
		fallback: fallback,
		cache:    cache,
	}
}

func (d *CelestiaServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/get/", d.HandleGet)
	mux.HandleFunc("/put/", d.HandlePut)

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

	// 1 read blob from cache if enabled
	if d.cache {
		d.log.Debug("Retrieving data from cached backends")
		input, err := d.multiSourceRead(r.Context(), comm, false)
		if err == nil {
			if _, err := w.Write(input); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		d.log.Warn("Failed to read from cache", "err", err)
	}
	// 2 read blob from Celestia
	input, err := d.store.Get(r.Context(), comm)
	if err != nil && errors.Is(err, plasma.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		if _, err := w.Write(input); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	//3 fallback
	if d.fallback {
		input, err = d.multiSourceRead(r.Context(), comm, true)
		if err != nil {
			d.log.Error("Failed to read from fallback", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		d.log.Error("Fallback disabled")
		w.WriteHeader(http.StatusInternalServerError)
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

// multiSourceRead ... reads from a set of backends and returns the first successfully read blob
func (b *CelestiaServer) multiSourceRead(ctx context.Context, commitment []byte, fallback bool) ([]byte, error) {

	if fallback {
		b.fallbackLock.RLock()
		defer b.fallbackLock.RUnlock()
	} else {
		b.cacheLock.RLock()
		defer b.cacheLock.RUnlock()
	}

	key := crypto.Keccak256(commitment)
	data, err := b.s3Store.Get(ctx, key)
	if err != nil {
		b.log.Warn("Failed to read from redundant target S3", "err", err)
		return nil, errors.New("no data found in any redundant backend")
	}
	// verify cert:data using EigenDA verification checks
	commit, err := b.store.CreateCommitment(data)
	if err != nil || !bytes.Equal(commit, data[9:]) {
		return nil, fmt.Errorf("celestia: invalid commitment: calldata=%x commit=%x err=%w", data, commit, err)
	}

	return data, nil
}
