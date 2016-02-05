package main

import (
	"os"
	"os/signal"
	"syscall"
	"flag"
	"time"
	"github.com/code-tool/artd-status-updater/statusupdater"
	"runtime"
)

func parseArguments() (string, *status_updater.EtcdConnectionParams, *status_updater.KeyUpdaterParameters) {
	var unixSocketPath string
	etcdParams := status_updater.EtcdConnectionParams{}
	keyUpdaterParams := status_updater.KeyUpdaterParameters{}

	// etcd connection params
	flag.StringVar(&etcdParams.Endpoint, "etcd-endpoint", "http://127.0.0.1:4001,http://127.0.0.1:2379", "Coma separated etcd endpoints")
	//
	flag.StringVar(&etcdParams.CertFile, "etcd-cert-file", "", "identify HTTPS client using this SSL certificate file")
	flag.StringVar(&etcdParams.KeyFile, "etcd-key-file", "", "identify HTTPS client using this SSL key file")
	flag.StringVar(&etcdParams.CaFile, "etcd-ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")
	flag.StringVar(&etcdParams.Username, "etcd-user", "", "Etcd auth user")
	flag.StringVar(&etcdParams.Password, "etcd-password", "", "Etcd auth password")
	flag.DurationVar(&etcdParams.ConnectionTimeout, "etcd-timeout", time.Second, "etcd connection timeout per request")
	flag.DurationVar(&etcdParams.RequestTimeout, "etcd-request-timeout", 5 * time.Second, "etcd requst timeout")

	// Server socket path
	flag.StringVar(&unixSocketPath, "socket", "/tmp/artd-status-updater.sock", "Path to unix socket liten on")

	// Key updater parameters
	flag.StringVar(&keyUpdaterParams.Key, "key", "/artifact-downloader/status", "Key where to push status")
	flag.DurationVar(&keyUpdaterParams.KeyTTL, "key-ttl", 10 * time.Second, "TTL for status key")
	flag.DurationVar(&keyUpdaterParams.RetryFreq, "key-retry-freq", 1 * time.Second, "Key update retry update freq")
	flag.DurationVar(&keyUpdaterParams.UpdateFreq, "key-update-freq", 5 * time.Second, "Key update freq")

	flag.Parse()

	return unixSocketPath, &etcdParams, &keyUpdaterParams
}

func main() {
	runtime.GOMAXPROCS(1)
	unixSocketPath, etcdParams, keyUpdaterParams := parseArguments()

	etcdKApi, err := status_updater.MakeNewEtcdKApi(etcdParams)
	if err != nil {
		panic(err)
	}

	dataChan := make(chan string)
	errorChan := make(chan error)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	dataListener := status_updater.NewDataListener(unixSocketPath, dataChan, errorChan)
	// start server
	err = dataListener.Start()
	if err != nil {
		panic(err)
	}

	// start key updater
	keyUpdater := status_updater.NewKeyUpdater(keyUpdaterParams, etcdKApi, dataChan);
	keyUpdater.Start()

	select {
	case err = <-errorChan:
		panic(err)
	case <-signalChan:
		dataListener.Stop()
		keyUpdater.Stop()
	}
}
