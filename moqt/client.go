package moqt

import (
	"log/slog"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

type Client struct {
	SupportedVersions []Version

	setupRW SetupRequestWriter

	Publisher  *Publisher
	Subscriber *Subscriber
}

func (c Client) Run(conn Connection) error {
	stream, err := conn.OpenStream()
	if err != nil {
		slog.Error("failed to open a bidirectional stream", slog.String("error", err.Error()))
		return err
	}

	_, err = stream.Write([]byte{byte(protocol.SESSION)})
	if err != nil {
		slog.Error("failed to send the Session Stream Type", slog.String("error", err.Error()))
		return err
	}

	if c.setupRW == nil {
		c.setupRW = defaultSetupRequestWriter{
			once:   new(sync.Once),
			stream: stream,
		}
	}
	err = c.setupRW.Setup(c.SupportedVersions)
	if err != nil {
		slog.Error("failed to set up", slog.String("error", err.Error()))
		return err
	}

	sess := Session{
		Connection: conn,
		stream:     stream,
	}

	if c.Publisher != nil {
		go c.Publisher.run(sess)
		slog.Info("run a publisher")
	}

	if c.Subscriber != nil {
		go c.Subscriber.run(sess)
		slog.Info("run a subscriber")
	}

	return nil
}
