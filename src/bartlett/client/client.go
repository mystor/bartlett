package client

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type CachedFile struct {
	ModTime time.Time
	Hash    string
}

// A cache of files which are on the computer
var cache map[string]CachedFile

// TODO(michael): These are copied DIRECTLY from server.
// move them into a common module between the client and the server
// so that they never get out of sync
type File struct {
	Data string // file is stored as string, not bytes
	Hash string
}

type SyncRequest struct {
	Added   map[string]File
	Changed map[string]File
	Unmod   map[string]File
}

type SyncResponse struct {
	Update map[string]File
	Delete []string
}

func shouldAcceptFile(file string) bool {
	cmd := exec.Command("git", "check-ignore", file)
	_, err := cmd.CombinedOutput()
	// The program will exit with status 0 if the file should be ignored
	// The program will exit with status 1 if the file should not be ignored
	// other exit codes are errors.
	if err != nil {
		// Non-zero exit code
		sys := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitStatus := sys.ExitStatus()
		if exitStatus == 1 {
			return true
		} else {
			log.Fatal("git check-ignore returned unexpected status code:",
				exitStatus)
		}
	}

	// Exit 0!
	return false
}

func getFileData(file string) File {
	fhandle, err := os.Open(file)

	if err != nil {
		log.Fatal("Error getting file data for file", err)
	}

	data, err := ioutil.ReadAll(fhandle)
	if err != nil {
		log.Fatal("Error reading file", err)
	}

	hash := sha1.Sum(data)

	return File{Hash: string(hash[:]), Data: string(data)}
}

func buildSyncRequest(basepath string) SyncRequest {
	// Build the sync request
	request := SyncRequest{
		Added:   make(map[string]File),
		Changed: make(map[string]File),
		Unmod:   make(map[string]File),
	}

	cb := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file should be gitignored
		if !shouldAcceptFile(path) {
			return nil
		}

		// Ignore directories
		if info.IsDir() {
			return nil
		}

		// Get the directory relative to the basepath
		relPath, err := filepath.Rel(basepath, path)
		if err != nil {
			return err
		}

		mtime := info.ModTime()

		cachedFile, pres := cache[relPath]
		if pres {
			// It's in the cache
			if cachedFile.ModTime == mtime {
				request.Unmod[relPath] = File{Hash: cachedFile.Hash}
				return nil
			} else {
				file := getFileData(path)
				if cachedFile.Hash == file.Hash {
					request.Unmod[relPath] = File{Hash: file.Hash}
				} else {
					cachedFile.Hash = file.Hash

					log.Println("Changed", relPath)
					request.Changed[relPath] = file
				}
			}
		} else {
			file := getFileData(path)

			// Add it to the cache
			cache[relPath] = CachedFile{
				Hash:    file.Hash,
				ModTime: mtime,
			}

			log.Println("Added", relPath)

			request.Added[relPath] = file
		}

		return nil
	}

	err := filepath.Walk(basepath, cb)
	if err != nil {
		log.Fatal("Error while walking directory:", err)
	}

	return request
}

func handleSyncResponse(resp SyncResponse, basepath string) {
	for fname, file := range resp.Update {
		// Ensure that the directory is created
		fullpath := path.Join(basepath, fname)
		// TODO(michael): Check if 0755 is a good set of permission bits
		err := os.MkdirAll(filepath.Dir(fullpath), 0755)
		if err != nil {
			log.Fatal("Error creating path for file", fname, ":", err)
		}

		// Write the file out
		// TODO(michael): Are these permissions right?
		fhandle, err := os.Create(fullpath)
		if err != nil {
			log.Fatal("Error creating file", fname, ":", err)
		}

		fhandle.Write([]byte(file.Data))
		fhandle.Close()

		// Store the file in the cache
		fdata, err := fhandle.Stat()
		if err != nil {
			log.Fatal("Error statting file", fname, ":", err)
		}

		cache[fname] = CachedFile{
			Hash:    file.Hash,
			ModTime: fdata.ModTime(),
		}
	}

	for _, fname := range resp.Delete {
		os.Remove(path.Join(basepath, fname))
	}
}

func pulse(url string, basepath string) {
	request := buildSyncRequest(basepath)

	js, err := json.Marshal(request)
	if err != nil {
		log.Fatal("Error marshaling JSON data for SyncRequest:", err)
	}

	res, err := http.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		log.Fatal("Error communicating with server:", err)
	}

	log.Println("Response:", res)

	syncResp := SyncResponse{}
	err = json.NewDecoder(res.Body).Decode(syncResp)
	res.Body.Close()
	if err != nil {
		log.Fatal("Error decoding response JSON from server:", err)
	}

	handleSyncResponse(syncResp, basepath)
}

func Run(wg *sync.WaitGroup, url string, localPort int) {
	defer wg.Done()

	log.Println("Started client, connecting to", url, "( localPort =", localPort, ")")

	// TODO(michael): Set up the localPort HTTP server

	// TODO(michael): Thw directory to connect in should probably be customizable
	// with a command line option, and passed into Run as an argument
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get working directory:", err)
	}

	i := 0
	for {
		i++
		log.Println("Pulsing", i)
		pulse(url, wd)

		// 0.5 seconds
		time.Sleep(time.Second / 20)
	}
}
