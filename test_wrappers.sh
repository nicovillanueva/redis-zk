for _ in $(seq 1 20); do export PORT0=$RANDOM ; python3 wrapper.py & ; done
