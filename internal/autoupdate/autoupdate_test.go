package autoupdate

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sortedMap(unsorted map[string]string) map[string]string {
	keys := make([]string, 0, len(unsorted))
	for k := range unsorted {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sorted := make(map[string]string)
	for _, k := range keys {
		sorted[k] = unsorted[k]
	}

	return sorted
}

func TestDownloadList(t *testing.T) {
	// arrange
	var serverName string = "/testing"

	var serialNumber string = "1"
	var updatePacket string = path.Join(serverName, "update.tar")
	var anotherSerialNumber string = "2"
	var anotherUpdatePacket = path.Join(serverName, "another_update.tar")
	var defaultUpdatePacket string = path.Join(serverName, "wildcard_update.tar")
	updatePackets := map[string]string{serialNumber: updatePacket, anotherSerialNumber: anotherUpdatePacket}

	server := autoUpdateServer{name: serverName, updatePackets: updatePackets, defaultUpdatePacket: defaultUpdatePacket}

	handlerToTest := autoUpdateHandler{servers: map[string]autoUpdateServer{serverName: server}}
	downloadListTarget := path.Join(handlerToTest.servers[serverName].name, "list.txt")

	t.Run("HeadDownloadList", func(t *testing.T) {
		// arrange
		req := httptest.NewRequest(http.MethodHead, downloadListTarget, nil)
		respRecorder := httptest.NewRecorder()

		// act
		handlerToTest.ServeHTTP(respRecorder, req)

		// assert
		var expectedBody string
		for s, u := range sortedMap(handlerToTest.servers[serverName].updatePackets) {
			expectedBody += fmt.Sprintf("%s;%s\n", s, u)
		}
		if handlerToTest.servers[serverName].defaultUpdatePacket != "" {
			expectedBody += fmt.Sprintf("*;%s\n", handlerToTest.servers[serverName].defaultUpdatePacket)
		}
		expectedETag := fmt.Sprintf("\"%x\"", md5.Sum([]byte(expectedBody)))

		assert.Equal(t, http.StatusOK, respRecorder.Code)
		assert.Empty(t, respRecorder.Body.Bytes())
		assert.Equal(t, expectedETag, respRecorder.Header().Get("ETag"))
	})

	t.Run("GetDownloadList", func(t *testing.T) {
		// arrange
		req := httptest.NewRequest(http.MethodGet, downloadListTarget, nil)
		respRecorder := httptest.NewRecorder()

		// act
		handlerToTest.ServeHTTP(respRecorder, req)

		// assert
		var expectedBody string
		for s, u := range sortedMap(handlerToTest.servers[serverName].updatePackets) {
			expectedBody += fmt.Sprintf("%s;%s\n", s, u)
		}
		if handlerToTest.servers[serverName].defaultUpdatePacket != "" {
			expectedBody += fmt.Sprintf("*;%s\n", handlerToTest.servers[serverName].defaultUpdatePacket)
		}
		expectedETag := fmt.Sprintf("\"%x\"", md5.Sum([]byte(expectedBody)))

		assert.Equal(t, http.StatusOK, respRecorder.Code)
		assert.Equal(t, expectedBody, respRecorder.Body.String())
		assert.Equal(t, expectedETag, respRecorder.Header().Get("ETag"))
	})

	t.Run("GetUpdatePacket", func(t *testing.T) {
		// arrange
		updatePacketTarget := path.Join(handlerToTest.servers[serverName].name, handlerToTest.servers[serverName].updatePackets[serialNumber])
		req := httptest.NewRequest(http.MethodGet, updatePacketTarget, nil)
		respRecorder := httptest.NewRecorder()

		// act
		handlerToTest.ServeHTTP(respRecorder, req)

		// assert
		assert.Equal(t, http.StatusOK, respRecorder.Code)
	})

	t.Run("GetAnotherUpdatePacket", func(t *testing.T) {
		// arrange
		updatePacketTarget := path.Join(handlerToTest.servers[serverName].name, handlerToTest.servers[serverName].updatePackets[anotherSerialNumber])
		req := httptest.NewRequest(http.MethodGet, updatePacketTarget, nil)
		respRecorder := httptest.NewRecorder()

		// act
		handlerToTest.ServeHTTP(respRecorder, req)

		// assert
		assert.Equal(t, http.StatusOK, respRecorder.Code)
	})

	t.Run("GetDefaultUpdatePacket", func(t *testing.T) {
		// arrange
		updatePacketTarget := path.Join(handlerToTest.servers[serverName].name, handlerToTest.servers[serverName].defaultUpdatePacket)
		req := httptest.NewRequest(http.MethodGet, updatePacketTarget, nil)
		respRecorder := httptest.NewRecorder()

		// act
		handlerToTest.ServeHTTP(respRecorder, req)

		// assert
		assert.Equal(t, http.StatusOK, respRecorder.Code)
	})
}
