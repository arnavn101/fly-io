
MAELSTROM_BIN = ./maelstrom/maelstrom

clean:
	rm -fr test

dup:
	cp -fr $(OLD) $(NEW)
	sed -i 's/$(OLD)/$(NEW)/g' $(NEW)/go.mod
	go work use $(NEW)

echo: clean
	go build -C maelstrom-echo -o maelstrom-echo.out
	$(MAELSTROM_BIN) test -w echo --bin maelstrom-echo/maelstrom-echo.out --node-count 1 --time-limit 10

unique-ids: clean
	go build -C maelstrom-unique-ids -o maelstrom-unique-ids.out
	$(MAELSTROM_BIN) test -w unique-ids --bin maelstrom-unique-ids/maelstrom-unique-ids.out --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

broadcast: clean
	go build -C maelstrom-broadcast -o maelstrom-broadcast.out
	$(MAELSTROM_BIN) test -w broadcast --bin maelstrom-broadcast/maelstrom-broadcast.out --node-count 1 --time-limit 20 --rate 10

multi-broadcast: clean
	go build -C maelstrom-multi-broadcast -o maelstrom-multi-broadcast.out
	$(MAELSTROM_BIN) test -w broadcast --bin maelstrom-multi-broadcast/maelstrom-multi-broadcast.out --node-count 5 --time-limit 20 --rate 10

multi-broadcast-ft: clean
	go build -C maelstrom-multi-broadcast-ft -o maelstrom-multi-broadcast-ft.out
	$(MAELSTROM_BIN) test -w broadcast --bin maelstrom-multi-broadcast-ft/maelstrom-multi-broadcast-ft.out --node-count 5 --time-limit 20 --rate 10 --nemesis partition
