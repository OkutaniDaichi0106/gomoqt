package moqtransport

// import "github.com/quic-go/webtransport-go"

// //type EncodedDataStream <-chan []byte
// //type DecodedDataStream <-chan Messager

// type TrackStream struct {
// }

// type PeepStream struct{} // TODO: need?

// type Streamer interface {
// 	Encode(StreamHeaderPeep, <-chan []byte) <-chan []byte
// }

// func (PeepStream) Encode(header StreamHeaderPeep, payloadCh <-chan []byte) <-chan []byte {
// 	dataCh := make(chan []byte, 1<<8)
// 	go func() {
// 		defer close(dataCh)

// 		objectIDCounter := 0

// 		// Add the Header
// 		dataCh <- header.serialize()

// 		// Add the Object Chunk
// 		var chunk ObjectChunk
// 		for payload := range payloadCh {
// 			chunk = ObjectChunk{
// 				ObjectID: ObjectID(objectIDCounter),
// 				Payload:  payload,
// 			}
// 			dataCh <- chunk.serialize()

// 			objectIDCounter++
// 		}

// 		// Send an Object with Object Status Code
// 		chunk = ObjectChunk{
// 			ObjectID:   ObjectID(objectIDCounter),
// 			Payload:    []byte{},
// 			StatusCode: END_OF_PEEP,
// 		}
// 		dataCh <- chunk.serialize()
// 	}()

// 	return dataCh
// }

// func (PeepStream) Decode(stream webtransport.ReceiveStream) <-chan []Messager {
// 	messageCh := make(chan Messager, 1<<8)
// 	go func() {
// 		defer close(messageCh)

// 		objectIDCounter := 0

// 		// Add the Header
// 		messageCh <- deserializeHeader()

// 		// Add the Object Chunk
// 		var chunk ObjectChunk
// 		for payload := range payloadCh {
// 			chunk = ObjectChunk{
// 				ObjectID: ObjectID(objectIDCounter),
// 				Payload:  payload,
// 			}
// 			dataCh <- chunk.serialize()

// 			objectIDCounter++
// 		}

// 		// Send an Object with Object Status Code
// 		chunk = ObjectChunk{
// 			ObjectID:   ObjectID(objectIDCounter),
// 			Payload:    []byte{},
// 			StatusCode: END_OF_PEEP,
// 		}
// 		dataCh <- chunk.serialize()
// 	}()

// 	return messageCh
// }
