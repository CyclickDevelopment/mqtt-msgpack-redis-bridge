# MQTT to Redis MSGPack Unpacker
This code example takes messages from an MQTT topic which are encoded using `msgpack` and
inserts them into a redis list.

The topic is set in `fsdca-push-client.go` or by passing the `-topic` argument.

## Usage
```
./prepare.sh
docker-compose up
```
