package status_updater

import (
	"time"
	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"fmt"
)

type KeyUpdaterParameters struct {
	Key        string
	KeyTTL     time.Duration
	RetryFreq  time.Duration
	UpdateFreq time.Duration
}

type KeyUpdater struct {
	params     *KeyUpdaterParameters
	etcdKApi   client.KeysAPI
	polTimer   *time.Timer
	lastStatus string
	DataChan   chan string
	ErrorChan  chan error
	stopChan   chan bool
}

func NewKeyUpdater(params *KeyUpdaterParameters, etcdKApi client.KeysAPI, dataChan chan string, errorChan chan error) (*KeyUpdater) {
	return &KeyUpdater{params, etcdKApi, nil, "", dataChan, errorChan, make(chan bool, 1)}
}

func (u *KeyUpdater) updateStatus() (time.Duration, error) {
	if u.lastStatus == "" {
		return 0, nil
	}

	start := time.Now()

	setOptions := client.SetOptions{TTL: u.params.KeyTTL}
	_, err := u.etcdKApi.Set(context.Background(), u.params.Key, u.lastStatus, &setOptions)

	return time.Since(start), err
}

func (u *KeyUpdater) updateLoop() {
	for {
		updateDuration, err := u.updateStatus()

		duration := u.params.UpdateFreq
		if err != nil {
			u.ErrorChan <- err
			// in case of error retry
			duration = u.params.RetryFreq
		}

		duration -= updateDuration

		if u.polTimer != nil {
			u.polTimer.Reset(duration)
		} else {
			u.polTimer = time.NewTimer(duration)
		}

		select {
		case u.lastStatus = <-u.DataChan:
			// new data arrived
			u.polTimer.Stop()
		case <-u.stopChan:
			// Got stop signal
			return
		case <-u.polTimer.C:
			// timeout reached
		}
	}
}

func (u *KeyUpdater) Start() {
	// try to recover old status in case of crash
	getResponse, err := u.etcdKApi.Get(context.Background(), u.params.Key, &client.GetOptions{})
	if err == nil {
		u.lastStatus = getResponse.Node.Value;
	}

	go u.updateLoop()
}

func (u *KeyUpdater) Stop() error {
	if u.polTimer != nil {
		u.polTimer.Stop()
	}

	u.stopChan <- true

	var err error = nil
	if u.lastStatus != "" {
		_, err = u.etcdKApi.Delete(context.Background(), u.params.Key, &client.DeleteOptions{})
	}

	return err
}

