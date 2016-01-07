package status_updater

import (
	"net"
	"bytes"
)

type DataListener struct {
	listener  net.Listener
	dataChan  chan string
	errorChan chan error
}

func NewDataListener(socketPath string, dataChan chan string, errorChan chan error) (*DataListener, error) {
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	return &DataListener{ln, dataChan, errorChan}, nil
}

func (l *DataListener) handleConnection(conn net.Conn) {
	buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))

	_, err := buf.ReadFrom(conn)
	conn.Close()

	if err != nil {
		l.errorChan <- err
		return
	}

	l.dataChan <- buf.String()
}

func (l *DataListener) Start() {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			l.errorChan <- err
			return
		}

		go l.handleConnection(conn)
	}
}

func (l *DataListener) Stop() {
	l.listener.Close()
}
