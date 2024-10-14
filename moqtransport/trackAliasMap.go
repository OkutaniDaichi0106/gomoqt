package moqtransport

import (
	"strings"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/moqtmessage"
)

type trackAliasMap struct {
	mu         sync.RWMutex
	nameIndex  map[string]map[string]moqtmessage.TrackAlias
	aliasIndex map[moqtmessage.TrackAlias]struct {
		trackNamespace moqtmessage.TrackNamespace
		trackName      string
	}
}

func (tamap *trackAliasMap) getAlias(tns moqtmessage.TrackNamespace, tn string) moqtmessage.TrackAlias {
	/*
	 * If an Track Alias exists, return the existing Track Alias
	 */
	tamap.mu.RLock()
	defer tamap.mu.RUnlock()
	existingAlias, ok := tamap.nameIndex[strings.Join(tns, "")][tn]
	if ok {
		return existingAlias
	}

	/*
	 * If no Track Alias was found, get new Track Alias and register the Track Namespace and Track Name with it
	 */
	tamap.mu.Lock()

	newAlias := moqtmessage.TrackAlias(len(tamap.aliasIndex))

	tamap.nameIndex[strings.Join(tns, "")][tn] = newAlias

	tamap.aliasIndex[newAlias] = struct {
		trackNamespace moqtmessage.TrackNamespace
		trackName      string
	}{
		trackNamespace: tns,
		trackName:      tn,
	}

	tamap.mu.Lock()

	return newAlias
}

func (tamap *trackAliasMap) getName(ta moqtmessage.TrackAlias) (moqtmessage.TrackNamespace, string, bool) {
	data, ok := tamap.aliasIndex[ta]
	if !ok {
		return nil, "", false
	}

	return data.trackNamespace, data.trackName, true
}

// type contentStatus struct {
// 	/*
// 	 * Current content status
// 	 */
// 	contentExists bool
// 	// finalGroupID  moqtmessage.GroupID
// 	// finalObjectID moqtmessage.ObjectID

// 	/*
// 	 * The Largest Group ID and Largest Object ID of the contents received so far
// 	 */
// 	largestGroupID  moqtmessage.GroupID
// 	largestObjectID moqtmessage.ObjectID
// }
