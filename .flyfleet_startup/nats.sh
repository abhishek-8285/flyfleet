#!/bin/bash
echo -e "\033]0;NATS Jetstream\007"
echo "Starting NATS Broker natively..."
cd nats-server-v2.10.11-linux-amd64 && ./nats-server -js -m 8222
exec bash
