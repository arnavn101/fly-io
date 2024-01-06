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
	sent_msg := map[string]map[int]bool{}

	ticker := time.NewTicker(time.Second * 2)

	lock_stored_msg := sync.Mutex{}
	lock_sent_msg := sync.Mutex{}

	getAllKeys := func() []int {
		lock_stored_msg.Lock()
		keys := maps.Keys(stored_msg)
		lock_stored_msg.Unlock()
		return keys
	}

	keysDifference := func(node_id string) []int {
		all_keys := getAllKeys()

		lock_sent_msg.Lock()
		sent_keys := maps.Keys(sent_msg[node_id])
		lock_sent_msg.Unlock()

		diff := make(map[int]bool, len(all_keys))
		for _, v1 := range all_keys {
			diff[v1] = true
		}

		for _, v2 := range sent_keys {
			if _, ok := diff[v2]; ok {
				diff[v2] = false
			}
		}

		return maps.Keys(diff)
	}

	go func() {
		for range ticker.C {
			for _, node := range n.NodeIDs() {
				if node != n.ID() {
					go n.Send(node, map[string]any{
						"type":     "broadcast-internal",
						"messages": keysDifference(node),
					})
				}
			}
		}
	}()

	n.Handle("broadcast-internal_ok", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		list_values := body["messages"].([]interface{})

		for _, interface_val := range list_values {
			value := int(interface_val.(float64))
			node_id := msg.Src

			lock_sent_msg.Lock()

			if _, ok := sent_msg[node_id]; !ok {
				sent_msg[node_id] = make(map[int]bool)
			}

			if _, ok := sent_msg[node_id][value]; !ok {
				sent_msg[node_id][value] = true
			}
			lock_sent_msg.Unlock()
		}

		return nil
	})

	n.Handle("broadcast-internal", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		list_values := body["messages"].([]interface{})

		for _, interface_val := range list_values {
			value := int(interface_val.(float64))
			lock_stored_msg.Lock()
			if _, ok := stored_msg[value]; !ok {
				stored_msg[value] = true
			}
			lock_stored_msg.Unlock()
		}

		body["type"] = "broadcast-internal_ok"
		return n.Reply(msg, body)
	})

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		value := int(body["message"].(float64))

		lock_stored_msg.Lock()
		stored_msg[value] = true
		lock_stored_msg.Unlock()

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
