package mock

type Connection struct {
}

func (c *Connection) Close() error {
	return nil
}
