package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type KeyValue struct {
	mtx  sync.RWMutex
	dict map[string]string
}

type KeyValueOperation struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Operation string `json:"operation"`
}

var _ raft.FSM = &KeyValue{}

func NewKeyValue() *KeyValue {
	return &KeyValue{
		dict: make(map[string]string),
	}
}

func (f *KeyValue) Apply(l *raft.Log) interface{} {
	var kvo KeyValueOperation
	err := json.Unmarshal(l.Data, &kvo)
	if err != nil {
		panic(err)
	}

	f.mtx.Lock()
	defer f.mtx.Unlock()
	switch kvo.Operation {
	case "set":
		f.dict[kvo.Key] = kvo.Value
	case "delete":
		delete(f.dict, kvo.Key)
	}

	return nil
}

func (f *KeyValue) Snapshot() (raft.FSMSnapshot, error) {
	// Make sure that any future calls to f.Apply() don't change the snapshot.
	snap := &KeyValueSnapshot{dict: make(map[string]string)}
	for k, v := range f.dict {
		snap.dict[k] = v
	}
	return snap, nil
}

func (f *KeyValue) Restore(r io.ReadCloser) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &f.dict)
	if err != nil {
		panic(err)
	}

	return nil
}

func (f *KeyValue) Get(key string) (string, error) {
	f.mtx.RLock()
	defer f.mtx.RUnlock()
	if v, ok := f.dict[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("key not found")
}

type KeyValueSnapshot struct {
	dict map[string]string
}

func (s *KeyValueSnapshot) Persist(sink raft.SnapshotSink) error {
	ba, _ := json.Marshal(s.dict)
	_, err := sink.Write(ba)
	if err != nil {
		err2 := sink.Cancel()
		if err2 != nil {
			panic(err2)
		}

		return fmt.Errorf("sink.Write(): %v", err)
	}
	return sink.Close()
}

func (s *KeyValueSnapshot) Release() {
}
