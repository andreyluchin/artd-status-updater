package status_updater

import (
	//"log"
	"net"
	//"io/ioutil"
	"bytes"
)

type DataListener struct {
	listener net.Listener
	dataChan chan string
}

func NewDataListener(socketPath string, dataChan chan string) (*DataListener, error) {
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	return &DataListener{ln, dataChan}, nil
}

func (l *DataListener) handleConnection(conn net.Conn) {
	// data, err := ioutil.ReadAll(conn);

	buf := bytes.NewBuffer(make([]byte, 0, 512))

	_, err := buf.ReadFrom(conn)
	// return buf.Bytes(), err

	if err != nil {
		return
	}

	conn.Close()

	// l.dataChan <- string(data[:])
	l.dataChan <- buf.String()
}

func (l *DataListener) Start() error {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			return err
			// log.Fatal("get client connection error: ", err)
		}

		go l.handleConnection(conn)
	}
}

func (l *DataListener) Stop() {
	l.listener.Close()
}
