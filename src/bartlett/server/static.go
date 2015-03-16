package server

import (
	"bartlett/shared"
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

func Sync(w http.ResponseWriter, r *http.Request) {
	syncReq := shared.NewSyncRequest()
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

func SyncPoll(syncReq *shared.SyncRequest) shared.SyncResponse {
	update := make(map[string]shared.File)
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

	return shared.SyncResponse{Update: update, Delete: del}
}

// TODO(michael): Should this ReadRequest be in shared.go?
type ReadRequest struct {
	Key    string
	Target shared.File
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

func ReadPoll(readReq ReadRequest) shared.File {
	var target shared.File
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
