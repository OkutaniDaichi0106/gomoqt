package message

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type TrackNamespace []string

func AppendTrackNamespace(b []byte, tns TrackNamespace) []byte {
	// Append the number of the elements of the Track Namespace
	b = quicvarint.Append(b, uint64(len(tns)))

	for _, v := range tns {
		// Append the length of the data
		b = quicvarint.Append(b, uint64(len(v)))

		// Append the data
		b = append(b, []byte(v)...)
	}

	return b
}

func readTrackNamespace(r Reader) (TrackNamespace, error) {
	// Get the number of the elements of the track namespace
	l, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}

	tns := make(TrackNamespace, l)

	var num uint64
	for i := uint64(0); i < l; i++ {
		// Get a length of a string in the Track Namespace
		num, err = quicvarint.Read(r)
		if err != nil {
			return nil, err
		}

		buf := make([]byte, num)

		_, err = r.Read(buf)
		if err == nil {
			tns = append(tns, string(buf))
			continue
		} else {
			if err == io.EOF {
				tns = append(tns, string(buf))
				return tns, nil
			}
			return nil, err
		}
	}

	return tns, nil
}

type TrackPrefix []string

func AppendTrackNamespacePrefix(b []byte, tnsp TrackPrefix) []byte {
	// Append the number of the elements of the Track Namespace
	b = quicvarint.Append(b, uint64(len(tnsp)))

	for _, v := range tnsp {
		// Append the length of the data
		b = quicvarint.Append(b, uint64(len(v)))

		// Append the data
		b = append(b, []byte(v)...)
	}

	return b
}

func ReadTrackNamespacePrefix(r quicvarint.Reader) (TrackPrefix, error) {
	// Get the number of the elements of the track namespace
	l, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}

	tnsp := make(TrackPrefix, l)

	var num uint64
	for i := uint64(0); i < l; i++ {
		// Get a length of a string in the Track Namespace
		num, err = quicvarint.Read(r)
		if err != nil {
			return nil, err
		}

		buf := make([]byte, num)

		_, err = r.Read(buf)
		if err == nil {
			tnsp = append(tnsp, string(buf))
			continue
		} else {
			if err == io.EOF {
				tnsp = append(tnsp, string(buf))
				return tnsp, nil
			}
			return nil, err
		}
	}

	return tnsp, nil
}
