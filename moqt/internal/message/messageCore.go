package message

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

func ReadControlMessage(r quicvarint.Reader) (quicvarint.Reader, error) {
	// Get a payload reader
	num, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}

	reader := io.LimitReader(r, int64(num))

	return quicvarint.NewReader(reader), nil
}

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

func ReadTrackNamespace(r quicvarint.Reader) (TrackNamespace, error) {
	var tns TrackNamespace

	// Get the number of the elements of the track namespace
	num, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1<<6)

	for i := uint64(0); i < num; i++ {
		// Get a length of a string in the Track Namespace
		num, err = quicvarint.Read(r)
		if err != nil {
			return nil, err
		}

		_, err = r.Read(buf[:num])
		if err == nil {
			tns = append(tns, string(buf[:num]))
			continue
		} else {
			if err == io.EOF {
				tns = append(tns, string(buf[:num]))
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
	num, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}

	tnsp := make(TrackPrefix, num)

	buf := make([]byte, 1<<6)

	for i := uint64(0); i < num; i++ {
		// Get a length of a string in the Track Namespace
		num, err = quicvarint.Read(r)
		if err != nil {
			return nil, err
		}

		_, err = r.Read(buf[:num])
		if err == nil {
			tnsp = append(tnsp, string(buf[:num]))
			continue
		} else {
			if err == io.EOF {
				tnsp = append(tnsp, string(buf[:num]))
				return tnsp, nil
			}
			return nil, err
		}
	}

	return tnsp, nil
}
