package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"golang.org/x/exp/maps"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	stored_msg := map[int]bool{}

	ticker := time.NewTicker(time.Second * 2)
	lock := sync.Mutex{}

	getAllKeys := func() []int {
		lock.Lock()
		keys := maps.Keys(stored_msg)
		lock.Unlock()
		return keys
	}

	go func() {
		for range ticker.C {
			for _, node := range n.NodeIDs() {
				if node != n.ID() {
					go n.Send(node, map[string]any{
						"type":     "broadcast-internal",
						"messages": getAllKeys(),
					})
				}
			}
		}
	}()

	n.Handle("broadcast-internal", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		list_values := body["messages"].([]interface{})

		for _, interface_val := range list_values {
			value := int(interface_val.(float64))
			lock.Lock()
			if _, ok := stored_msg[value]; !ok {
				stored_msg[value] = true
			}
			lock.Unlock()
		}

		return nil
	})

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		value := int(body["message"].(float64))

		lock.Lock()
		stored_msg[value] = true
		lock.Unlock()

		return n.Reply(msg, map[string]any{
			"type": "broadcast_ok",
		})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": getAllKeys(),
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
