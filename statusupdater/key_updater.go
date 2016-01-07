package status_updater

import (
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"time"
)

type KeyUpdaterParameters struct {
	Key        string
	KeyTTL     time.Duration
	UpdateFreq time.Duration
}

type KeyUpdater struct {
	params     *KeyUpdaterParameters
	etcdKApi   client.KeysAPI
	polTimer   *time.Timer
	lastStatus string
	DataChan   chan string
	ErrorChan  chan error
}

func NewKeyUpdater(params *KeyUpdaterParameters, etcdKApi client.KeysAPI, dataChan chan string) (*KeyUpdater) {
	return &KeyUpdater{params, etcdKApi, nil, "", dataChan, make(chan error)}
}

func (u *KeyUpdater) updateStatus() {
	if u.lastStatus == "" {
		u.waitForDataOrTimeout(0)
	}

	start := time.Now()

	setOptions := client.SetOptions{TTL: u.params.KeyTTL}
	_, err := u.etcdKApi.Set(context.Background(), u.params.Key, u.lastStatus, &setOptions)

	if err != nil {
		u.ErrorChan <- err
		return
	}

	u.waitForDataOrTimeout(time.Since(start))
}

func (u *KeyUpdater) waitForDataOrTimeout(timeOffset time.Duration) {
	duration := u.params.UpdateFreq - timeOffset
	if u.polTimer != nil {
		u.polTimer.Reset(duration)
	} else {
		u.polTimer = time.NewTimer(duration)
	}

	select {
	case <-u.polTimer.C:
		u.updateStatus()
	case u.lastStatus = <-u.DataChan:
		u.polTimer.Stop()
		u.updateStatus()
	}
}

func (u *KeyUpdater) Start() {
	// todo read prev value
	u.waitForDataOrTimeout(0)
}

func (u *KeyUpdater) Stop() error {
	if u.polTimer != nil {
		u.polTimer.Stop()
	}

	deleteOptions := client.DeleteOptions{}
	_, err := u.etcdKApi.Delete(context.Background(), u.params.Key, &deleteOptions)

	return err
}

