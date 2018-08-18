package promcase

import (
	"github.com/fatih/structs"
	log "github.com/sirupsen/logrus"
)

// Queue is a buffered channel of incoming UDP messages.
var Queue chan Message

// InitQueue initializes the buffered channel.
func InitQueue(max int) {
	Queue = make(chan Message, max)
}

// ProcessQueue function starts a go routine which consumes messages from Queue.
func ProcessQueue() {
	go func() {
		for {
			select {
			case m := <-Queue:
				log.WithFields(log.Fields{"m": structs.Map(m)}).Debug("received udp message")
				err := m.Process()
				if err != nil {
					log.WithFields(log.Fields{"m": structs.Map(m)}).Error(err)
				}
			}
		}
	}()
}
