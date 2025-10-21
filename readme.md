Building a Key value store based on quorum system

It should have a store where i can store key and values

## Build a Store object

```go

type Value struct{
	Value string `json:"value"`
	TS int64 `json:"timestamp"`// timestamp for Lamport-style
}


type Store struct{
	mu sync.RWMutex
	m map[string]Value
}

```

```go
func NewStore() *Store {
	return &Store{
		m:make(map[string]Value),
	}
}

// RLock is used to allow multiple goroutines to read from the store concurrently without blocking each other, while still preventing reads from happening during a write.

func (s *Store)Get(key string)(Value,bool){
	s.mu.RLock()
	defer s.mu.RUnlock()
	v,ok:= s.m[key]
	return v,ok
}


func (s *Store)Put(key string,value Value){
	s.mu.Lock()
	defer s.mu.Unlock()
	curr,ok := s.m[key]
	if !ok || curr.TS <= value.TS{
		s.m[key] = value
	}
}


```

```go

type Node struct{
	id string
	addr string
	httpAddr string
	store *Store
	client *http.Client
	peers []string

	ts int64
	mu sync.Mutex

	N int
	W int
	R int
}


func NewNode(id,addr string , peers []string) *Node {
	N := len(peers)
	q := N /2 +1
	return &Node{
		id:id,
		addr: addr,
		store: NewStore(),
		client: &http.Client{Timeout: 2*time.Second},
		peers: peers,

		ts: time.Now().UnixNano(),
		N: N,
		W:q,
		R:q,
	}
}


```

Why N/2 + 1?

In quorum-based systems (like Dynamo, Cassandra, or Raft-style reads/writes):

Write quorum (W): the minimum number of nodes that must acknowledge a write.

Read quorum (R): the minimum number of nodes that must respond to a read.

Setting W = R = N/2 + 1 ensures:

Majority intersection: Any read quorum and write quorum will always overlap on at least one node.

This guarantees strong consistency, because a read will always see the latest write.

Formally:

R+W>N
R+W>N

Here, R = W = N/2 + 1:

(N/2+1)+(N/2+1)=N+2>N
(N/2+1)+(N/2+1)=N+2>N

So reads and writes will always intersect at at least one node, ensuring the latest value is observed.
