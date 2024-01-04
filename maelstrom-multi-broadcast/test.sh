rm -fr store
go build .
../maelstrom/maelstrom test -w broadcast --bin maelstrom-multi-broadcast --node-count 5 --time-limit 20 --rate 10