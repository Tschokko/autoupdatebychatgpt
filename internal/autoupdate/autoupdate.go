package autoupdate

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strings"
)

// autoUpdateServer holds the server's name, a map of serial numbers to update packet paths, and a default update packet path.
type autoUpdateServer struct {
	name                string
	updatePackets       map[string]string
	defaultUpdatePacket string
}

// autoUpdateHandler manages multiple autoUpdateServers.
type autoUpdateHandler struct {
	servers map[string]autoUpdateServer
}

// ServeHTTP handles incoming HTTP requests to the auto update server.
func (h *autoUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodHead, http.MethodGet:
		for _, server := range h.servers {
			if strings.HasPrefix(r.URL.Path, server.name) {
				// Extract the requested file's basename.
				requestedFile := path.Base(r.URL.Path)

				// Handle list.txt requests.
				if requestedFile == "list.txt" {
					h.serveUpdateList(w, r, server)
					return
				}

				// Handle update packet requests.
				h.serveUpdatePacket(w, r, server, requestedFile)
				return
			}
		}
		http.NotFound(w, r)

	default:
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
	}
}

// serveUpdateList generates and serves the update list, ensuring the list is sorted.
func (h *autoUpdateHandler) serveUpdateList(w http.ResponseWriter, r *http.Request, server autoUpdateServer) {
	var body string
	sortedKeys := sortUpdatePackets(server.updatePackets)
	for _, serial := range sortedKeys {
		packet := server.updatePackets[serial]
		body += fmt.Sprintf("%s;%s\n", serial, packet)
	}
	if server.defaultUpdatePacket != "" {
		body += fmt.Sprintf("*;%s\n", server.defaultUpdatePacket)
	}

	if r.Method == http.MethodGet {
		_, _ = w.Write([]byte(body))
	}

	// Set the ETag header.
	etag := fmt.Sprintf("\"%x\"", md5.Sum([]byte(body)))
	w.Header().Set("ETag", etag)
}

// serveUpdatePacket serves the requested update packet if it exists.
func (h *autoUpdateHandler) serveUpdatePacket(w http.ResponseWriter, r *http.Request, server autoUpdateServer, requestedFile string) {
	for _, packet := range server.updatePackets {
		if path.Base(packet) == requestedFile {
			http.ServeFile(w, r, packet)
			return
		}
	}
	if path.Base(server.defaultUpdatePacket) == requestedFile {
		http.ServeFile(w, r, server.defaultUpdatePacket)
		return
	}
	http.NotFound(w, r)
}

// sortUpdatePackets returns a sorted slice of keys from the update packets map.
func sortUpdatePackets(packets map[string]string) []string {
	keys := make([]string, 0, len(packets))
	for k := range packets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
