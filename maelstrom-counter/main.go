package main

import (
	"context"
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)

	key := n.ID()
	ctx := context.Background()

	n.Handle("init", func(msg maelstrom.Message) error {
		if err := kv.Write(ctx, key, 0); err != nil {
			log.Fatal(err)
		}

		return nil
	})

	n.Handle("add", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		delta := int(body["delta"].(float64))
		value, err := kv.ReadInt(ctx, key)

		if err != nil {
			return err
		}

		log.Printf("add %d to %d\n", delta, value)
		err = kv.Write(ctx, key, value+delta)

		if err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{
			"type": "add_ok",
		})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		tot_value := 0

		for _, node := range n.NodeIDs() {
			if node == n.ID() {
				value, err := kv.ReadInt(ctx, key)

				if err != nil {
					return err
				}

				tot_value += value
			} else {
				msg, err := n.SyncRPC(ctx, node, map[string]any{
					"type": "local_read",
				})

				if err != nil {
					return err
				}

				var body map[string]any
				if err := json.Unmarshal(msg.Body, &body); err != nil {
					return err
				}

				value := int(body["value"].(float64))
				tot_value += value
			}
		}

		return n.Reply(msg, map[string]any{
			"type":  "read_ok",
			"value": tot_value,
		})
	})

	n.Handle("local_read", func(msg maelstrom.Message) error {
		value, err := kv.ReadInt(ctx, key)

		if err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{
			"type":  "local_read_ok",
			"value": value,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
