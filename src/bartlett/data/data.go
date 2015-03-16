package data

import ()

// TODO(michael): Change data to []byte and hash to [20]byte
// Basic data storage type
type File struct {
	Data []byte // file is stored as string, not bytes
	Hash [20]byte
}

// Static synchronization request types
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

// Live synchronization request types
type WatchRequest struct {
	Key    string
	Target File
}

type WatchResponse struct {
	Target File
	Locked bool
}

type PushRequest struct {
	Key     string
	Updated File
}

type UnlockRequest struct {
	Key string
}
