\

# Go Quorum Key-Value Store

A **distributed key-value store** in Go using a **quorum-based replication model**. Supports concurrent reads and writes across multiple nodes with configurable read (`R`) and write (`W`) quorum sizes.

---

## Features

- **Quorum Writes & Reads** – Ensures strong consistency using W/N and R/N quorum settings.
- **Vector Timestamps** – Uses strictly increasing timestamps for last-write-wins conflict resolution.
- **HTTP API** – Simple REST endpoints for external reads/writes and internal node replication.
- **Concurrent Safe** – Thread-safe in-memory store with Go `sync.RWMutex`.
- **Peer-to-Peer Replication** – Writes are propagated to all configured peers.

---

## Endpoints

| Method | Path                       | Description                                                           |
| ------ | -------------------------- | --------------------------------------------------------------------- |
| `POST` | `/put`                     | Write a key-value pair. JSON body: `{ "key": "foo", "value": "bar" }` |
| `GET`  | `/get?key=<key>`           | Read a key’s value from the cluster                                   |
| `POST` | `/internal/write`          | Internal endpoint used for replication between nodes                  |
| `GET`  | `/internal/read?key=<key>` | Internal endpoint to fetch a value from a node                        |

---

## Installation & Run

1. Clone repository:

```bash
git clone https://github.com/jsndz/KV-store.git
cd KV-store
```

2. Build:

```bash
go build -o kvstore main.go
```

3. Run nodes (example with 3 nodes):

```bash
# Node 1
./kvstore -id=node1 -addr=localhost:8001 -peers=localhost:8001,localhost:8002,localhost:8003

# Node 2
./kvstore -id=node2 -addr=localhost:8002 -peers=localhost:8001,localhost:8002,localhost:8003

# Node 3
./kvstore -id=node3 -addr=localhost:8003 -peers=localhost:8001,localhost:8002,localhost:8003
```

---

## Example Usage

Write a key:

```bash
curl -X POST http://localhost:8001/put -d '{"key":"foo","value":"bar"}' -H "Content-Type: application/json"
```

Read a key:

```bash
curl http://localhost:8002/get?key=foo
```

---

## Notes

- Currently **in-memory** only; restart will lose data.
- Configurable quorum based on the number of peers: `W = N/2 + 1`, `R = N/2 + 1`.
- Can handle **partial node failures** as long as quorum requirements are met.

---

## License

MIT License.

---
