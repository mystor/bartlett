package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	StaticPollNum  = 3
	StaticPollTime = 500
)

var static State

type SyncRequest struct {
	Added   map[string]File
	Changed map[string]File
	Unmod   map[string]File
}

func NewSyncRequest() *SyncRequest {
	added := make(map[string]File)
	changed := make(map[string]File)
	unmod := make(map[string]File)
	return &SyncRequest{Changed: changed, Unmod: unmod, Added: added}
}

type SyncResponse struct {
	Update map[string]File
	Delete []string
}

func Sync(w http.ResponseWriter, r *http.Request) {
	syncReq := NewSyncRequest()
	err := json.NewDecoder(r.Body).Decode(&syncReq)
	if err != nil {
		log.Fatal(err)
	}

	resp := SyncPoll(syncReq)

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func SyncPoll(syncReq *SyncRequest) SyncResponse {
	update := make(map[string]File)
	del := make([]string, 0)

	for key, clientFile := range syncReq.Added {
		serverFile, prs := static.Files[key]
		if !prs {
			static.Files[key] = clientFile
		} else {
			update[key] = serverFile
		}
	}

	for key, clientFile := range syncReq.Changed {
		static.Files[key] = clientFile
	}

	for i := 0; i < StaticPollNum; i++ {
		for key, clientFile := range syncReq.Unmod {
			serverFile, prs := static.Files[key]
			if prs && serverFile.Hash != clientFile.Hash {
				update[key] = serverFile
			} else if !prs {
				del = append(del, key)
			}
		}

		for key, serverFile := range static.Files {
			_, prs1 := syncReq.Added[key]
			_, prs2 := syncReq.Changed[key]
			_, prs3 := syncReq.Unmod[key]
			if !(prs1 || prs2 || prs3) {
				update[key] = serverFile
			}
		}

		if len(update) > 0 || len(del) > 0 {
			break
		} else {
			time.Sleep(StaticPollTime * time.Millisecond)
		}
	}

	return SyncResponse{Update: update, Delete: del}
}

type ReadRequest struct {
	Key    string
	Target File
}

func Read(w http.ResponseWriter, r *http.Request) {
	var readReq ReadRequest
	err := json.NewDecoder(r.Body).Decode(&readReq)
	if err != nil {
		fmt.Println("error in read")
		log.Fatal(err)
	}

	resp := ReadPoll(readReq)

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func ReadPoll(readReq ReadRequest) File {
	var target File
	for i := 0; i < StaticPollNum; i++ {
		liveFile, livePrs := live.Files[readReq.Key]
		if livePrs {
			target = liveFile
			break
		}
		serverFile, prs := static.Files[readReq.Key]
		target = serverFile
		if prs && serverFile.Hash != readReq.Target.Hash {
			break
		}
		time.Sleep(StaticPollTime * time.Millisecond)
	}
	return target
}
