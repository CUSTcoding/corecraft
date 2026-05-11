package zmqsub

import (
	"encoding/hex"
	"log"
	"time"

	zmq "github.com/pebbe/zmq4"

	"bitcoin-monitor-t2/internal/store"
)


func Start(zmqURL string, s *store.EventStore) {
	if zmqURL == "" {
		zmqURL = "tcp://127.0.0.1:28332"
	}

	log.Printf("[ZMQ] Conectando em %s", zmqURL)

	for {
		if err := run(zmqURL, s); err != nil {
			log.Printf("[ZMQ] Erro: %v — reconectando em 5s...", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func run(zmqURL string, s *store.EventStore) error {
	sock, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		return err
	}
	defer sock.Close()

	
	sock.SetSubscribe("hashblock")
	sock.SetSubscribe("hashtx")

	if err := sock.Connect(zmqURL); err != nil {
		return err
	}

	log.Println("[ZMQ] Subscrito em hashblock e hashtx")

	for {
		
		msg, err := sock.RecvMessageBytes(0)
		if err != nil {
			return err
		}

		if len(msg) < 2 {
			continue
		}

		topic := string(msg[0])
		data := msg[1]

		
		hash := hex.EncodeToString(reverseBytes(data))

		switch topic {
		case "hashblock":
			log.Printf("[ZMQ] 🟧 Bloco: %s", hash)
			s.AddBlock(hash)

		case "hashtx":
			log.Printf("[ZMQ] 💸 Tx: %s", hash)
			s.AddTx(hash)
		}
	}
}


func reverseBytes(b []byte) []byte {
	out := make([]byte, len(b))
	for i, v := range b {
		out[len(b)-1-i] = v
	}
	return out
}