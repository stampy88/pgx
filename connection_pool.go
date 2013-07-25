package pgx

type ConnectionPoolOptions struct {
	MaxConnections int // max simultaneous connections to use (currently all are immediately connected)
	AfterConnect   func(*Connection) error
}

type ConnectionPool struct {
	connectionChannel chan *Connection
	parameters        ConnectionParameters // parameters used when establishing connection
	options           ConnectionPoolOptions
}

// NewConnectionPool creates a new ConnectionPool. parameters are passed through to
// Connect directly.
func NewConnectionPool(parameters ConnectionParameters, options ConnectionPoolOptions) (p *ConnectionPool, err error) {
	p = new(ConnectionPool)
	p.connectionChannel = make(chan *Connection, options.MaxConnections)

	p.parameters = parameters
	p.options = options

	for i := 0; i < p.options.MaxConnections; i++ {
		var c *Connection
		c, err = p.createConnection()
		if err != nil {
			return
		}
		p.connectionChannel <- c
	}

	return
}

// Acquire takes exclusive use of a connection until it is released.
func (p *ConnectionPool) Acquire() (c *Connection) {
	c = <-p.connectionChannel
	return
}

// Release gives up use of a connection.
func (p *ConnectionPool) Release(c *Connection) {
	if c.TxStatus != 'I' {
		c.Execute("rollback")
	}
	p.connectionChannel <- c
}

// Close ends the use of a connection by closing all underlying connections.
func (p *ConnectionPool) Close() {
	for i := 0; i < p.options.MaxConnections; i++ {
		c := <-p.connectionChannel
		_ = c.Close()
	}
}

func (p *ConnectionPool) createConnection() (c *Connection, err error) {
	c, err = Connect(p.parameters)
	if err != nil {
		return
	}
	if p.options.AfterConnect != nil {
		err = p.options.AfterConnect(c)
		if err != nil {
			return
		}
	}
	return
}

// SelectFunc acquires a connection, delegates the call to that connection, and releases the connection
func (p *ConnectionPool) SelectFunc(sql string, onDataRow func(*DataRowReader) error, arguments ...interface{}) (err error) {
	c := p.Acquire()
	defer p.Release(c)

	return c.SelectFunc(sql, onDataRow, arguments...)
}

// SelectRows acquires a connection, delegates the call to that connection, and releases the connection
func (p *ConnectionPool) SelectRows(sql string, arguments ...interface{}) (rows []map[string]interface{}, err error) {
	c := p.Acquire()
	defer p.Release(c)

	return c.SelectRows(sql, arguments...)
}

// SelectRow acquires a connection, delegates the call to that connection, and releases the connection
func (p *ConnectionPool) SelectRow(sql string, arguments ...interface{}) (row map[string]interface{}, err error) {
	c := p.Acquire()
	defer p.Release(c)

	return c.SelectRow(sql, arguments...)
}

// SelectValue acquires a connection, delegates the call to that connection, and releases the connection
func (p *ConnectionPool) SelectValue(sql string, arguments ...interface{}) (v interface{}, err error) {
	c := p.Acquire()
	defer p.Release(c)

	return c.SelectValue(sql, arguments...)
}

// SelectValues acquires a connection, delegates the call to that connection, and releases the connection
func (p *ConnectionPool) SelectValues(sql string, arguments ...interface{}) (values []interface{}, err error) {
	c := p.Acquire()
	defer p.Release(c)

	return c.SelectValues(sql, arguments...)
}

// Execute acquires a connection, delegates the call to that connection, and releases the connection
func (p *ConnectionPool) Execute(sql string, arguments ...interface{}) (commandTag string, err error) {
	c := p.Acquire()
	defer p.Release(c)

	return c.Execute(sql, arguments...)
}
