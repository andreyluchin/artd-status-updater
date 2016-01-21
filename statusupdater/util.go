package status_updater

import (
	"os"
	"time"
	"strings"
	"net/http"
	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/pkg/transport"
	"net/url"
)

type EtcdConnectionParams struct {
	Endpoint          string

	CertFile          string
	KeyFile           string
	CaFile            string

	Username          string
	Password          string

	ConnectionTimeout time.Duration
	RequestTimeout    time.Duration
}

func getPeersFlagValue(params *EtcdConnectionParams) []string {
	peerstr := params.Endpoint

	if peerstr == "" {
		peerstr = os.Getenv("ARTD_ST_ETCD_ENDPOINT")
	}

	// If we still don't have peers, use a default
	if peerstr == "" {
		peerstr = "http://127.0.0.1:4001,http://127.0.0.1:2379"
	}

	return strings.Split(peerstr, ",")
}

func getEndpoints(params *EtcdConnectionParams) ([]string, error) {
	eps := getPeersFlagValue(params)

	for i, ep := range eps {
		u, err := url.Parse(ep)
		if err != nil {
			return nil, err
		}

		if u.Scheme == "" {
			u.Scheme = "http"
		}

		eps[i] = u.String()
	}

	return eps, nil
}

func getTransport(params *EtcdConnectionParams) (*http.Transport, error) {
	tls := transport.TLSInfo{
		CAFile:   params.CaFile,
		CertFile: params.CertFile,
		KeyFile:  params.KeyFile,
	}

	return transport.NewTransport(tls, params.ConnectionTimeout)
}

func MakeNewEtcdKApi(params *EtcdConnectionParams) (client.KeysAPI, error) {
	eps, err := getEndpoints(params)
	if err != nil {
		return nil, err
	}

	tr, err := getTransport(params)
	if err != nil {
		return nil, err
	}

	cfg := client.Config{
		Transport:               tr,
		Endpoints:               eps,
		HeaderTimeoutPerRequest: params.RequestTimeout,
	}

	if params.Username != "" {
		cfg.Username = params.Username
		cfg.Password = params.Password
	}

	etcdClient, err := client.New(cfg)
	if err != nil {
		return nil, err;
	}

	return client.NewKeysAPI(etcdClient), nil
}
