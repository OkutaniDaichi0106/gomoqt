package moqt

// func readGroup(r io.Reader) (SubscribeID, GroupSequence, error) {
// 	// Read a GROUP message
// 	var gm message.GroupMessage
// 	_, err := gm.Decode(r)
// 	if err != nil {
// 		slog.Error("failed to read a GROUP message", slog.String("error", err.Error()))
// 		return 0, 0, err
// 	}

// 	//
// 	return SubscribeID(gm.SubscribeID), GroupSequence(gm.GroupSequence), nil
// }

// func writeGroup(w io.Writer, id SubscribeID, seq GroupSequence) error {
// 	gm := message.GroupMessage{
// 		SubscribeID:   message.SubscribeID(id),
// 		GroupSequence: message.GroupSequence(seq),
// 	}
// 	_, err := gm.Encode(w)
// 	if err != nil {
// 		slog.Error("failed to send a GROUP message", slog.String("error", err.Error()))
// 		return err
// 	}

// 	return nil
// }
