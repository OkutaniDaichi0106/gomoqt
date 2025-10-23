// Package moqt implements the MOQ Transfork protocol, providing a multiplexer for track routing.
// It allows handling of track subscriptions, announcements, and track serving.
//
// client example:
/*
	client := moqt.Client{
		Logger: slog.Default(),
	}

	sess, err := client.Dial(context.Background(), "https://localhost:4469/broadcast", nil)
	if err != nil {
    	// handle error
	}

	annRecv, err := sess.AcceptAnnounce("/")
	if err != nil {
		// handle error
	}
	defer annRecv.Close()

	ann, err := annRecv.ReceiveAnnouncement(context.Background())
	if err != nil {
		// handle error
	}

	go func(ann *moqt.Announcement) {
		if !ann.IsActive() {
			return
		}

		tr, err := sess.Subscribe(ann.BroadcastPath(), "index", nil)
		if err != nil {
			// handle error
		}
		defer tr.Close()

		gr, err := tr.AcceptGroup(context.Background())
		if err != nil {
			// handle error
		}

		go func(gr *moqt.GroupReader) {
			for {
				frame, err := gr.ReadFrame()
				if err != nil {
					// handle error
				}
			}
		}(gr)
	}(ann)
*/
package moqt
