package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Value struct{
	Value string `json:"value"`
	TS int64 `json:"timestamp"`
}


type Store struct{
	mu sync.RWMutex
	m map[string]Value
}

func NewStore() *Store {
	return &Store{
		m:make(map[string]Value),
	}
}


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

func (n *Node) nextTs()int64{
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.ts+1 > time.Now().UnixNano() {
		return n.ts+1 
	}
	return time.Now().UnixNano()
}


func headerJSON(w http.ResponseWriter, code int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(v)
}

func (n *Node) HandlePut(w http.ResponseWriter, r *http.Request)  {
	//extract and decode the method
	var req struct {
		Key string 		`json:"key"`
		Value string  	`json:"value"`
	}

	if r.Method != http.MethodPost{
		headerJSON(w,http.StatusMethodNotAllowed, map[string]string{"error": "POST only"})
		return
	}
	if err:= json.NewDecoder(r.Body).Decode(&req); err!=nil{
		headerJSON(w,http.StatusBadRequest, map[string]string{"error": "Unable to decode body"})
		return
	}
	//create new ts 
	ts:=n.nextTs()
	v:=Value{Value: req.Value,TS: ts}
	ctx,cancel := context.WithTimeout(r.Context(),time.Second *10)
	defer cancel()
	
	//send them to all users concurrently
	ack := make(chan error,n.N)

	for _,peer := range n.peers{
		p := peer 
		go func(peer string) {
			data,err := json.Marshal(map[string]string{"key":req.Key,"value":v.Value,"ts":strconv.FormatInt(v.TS, 10)})
			if err!=nil{
				return
			}
			req, err := http.NewRequestWithContext(ctx, "POST",  fmt.Sprintf("http://%s/internal/write", peer), bytes.NewBuffer(data))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")

			if n.client == nil {
				n.client = http.DefaultClient
			}
			resp,err :=n.client.Do(req)
			if err != nil {
				ack <- err
				return
			}

			io.Copy(io.Discard, resp.Body)// read the body until its empty since the http is stream you cant close anytime
			resp.Body.Close()
            if resp.StatusCode >= 200 && resp.StatusCode < 300 {
                ack <- nil
            } else {
                ack <- fmt.Errorf("status %d", resp.StatusCode)
            }
		}(p)
	}
	success :=0
	failures := 0
	needed :=n.W	
	for i:=0;i<n.N;i++{
		select {
		case err := <-ack:
			if err==nil{
				success++
			} else {
				failures++
			}
		case <-ctx.Done():
			failures +=(n.N - success - failures)
			headerJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "timeout"})
			return
		}
		
		if success>needed{
			headerJSON(w, http.StatusOK, map[string]interface{}{"result": "ok", "acks": success})
			return
		}
		if failures>needed - success{
			headerJSON(w, http.StatusInternalServerError, map[string]interface{}{"error": "quorum not reached", "acks": success})
			return
		}
	}
	headerJSON(w, http.StatusInternalServerError, map[string]interface{}{"error": "quorum not reached", "acks": success})

}

func (n *Node) HandleGet(w http.ResponseWriter, r *http.Request)  {
	
}

func (n *Node) HandleInternalWrite(w http.ResponseWriter, r *http.Request){

}