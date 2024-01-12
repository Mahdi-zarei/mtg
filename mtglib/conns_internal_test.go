package mtglib

import (
	"bytes"
	"context"
	"errors"
	"github.com/9seconds/mtg/v2/internal/testlib"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
)

type ConnRewindBaseConn struct {
	testlib.EssentialsConnMock

	readBuffer bytes.Buffer
}

func (c *ConnRewindBaseConn) Read(p []byte) (int, error) {
	c.Called(p)

	return c.readBuffer.Read(p) //nolint: wrapcheck
}

type ConnTrafficTestSuite struct {
	suite.Suite

	eventStreamMock *EventStreamMock
	connMock        *testlib.EssentialsConnMock
	conn            io.ReadWriter
}

func (suite *ConnTrafficTestSuite) SetupTest() {
	suite.eventStreamMock = &EventStreamMock{}
	suite.connMock = &testlib.EssentialsConnMock{}
	suite.conn = connTraffic{
		Conn: suite.connMock,
		ctx:  context.Background(),
	}
}

func (suite *ConnTrafficTestSuite) TearDownTest() {
	suite.eventStreamMock.AssertExpectations(suite.T())
	suite.connMock.AssertExpectations(suite.T())
}

func (suite *ConnTrafficTestSuite) TestWriteNothingOk() {
	suite.connMock.On("Write", mock.Anything).Once().Return(0, nil)

	n, err := suite.conn.Write(make([]byte, 10))
	suite.NoError(err)
	suite.Equal(0, n)
}

func (suite *ConnTrafficTestSuite) TestWriteNothingErr() {
	suite.connMock.On("Write", mock.Anything).Once().Return(0, io.EOF)

	n, err := suite.conn.Write(make([]byte, 10))
	suite.True(errors.Is(err, io.EOF))
	suite.Equal(0, n)
}

type ConnRewindTestSuite struct {
	suite.Suite

	connMock *ConnRewindBaseConn
	conn     *connRewind
}

func (suite *ConnRewindTestSuite) SetupTest() {
	suite.connMock = &ConnRewindBaseConn{}
	suite.conn = newConnRewind(suite.connMock)
}

func (suite *ConnRewindTestSuite) TearDownTest() {
	suite.connMock.AssertExpectations(suite.T())
}

func (suite *ConnRewindTestSuite) TestRead() {
	suite.connMock.On("Read", mock.Anything)
	suite.connMock.readBuffer.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	buf := make([]byte, 2)

	n, err := suite.conn.Read(buf)
	suite.NoError(err)
	suite.Equal(2, n)
	suite.Equal([]byte{1, 2}, buf)

	n, err = suite.conn.Read(buf)
	suite.NoError(err)
	suite.Equal(2, n)
	suite.Equal([]byte{3, 4}, buf)

	suite.conn.Rewind()

	data, err := io.ReadAll(suite.conn)
	suite.NoError(err)
	suite.Equal([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, data)
}

func TestConnTraffic(t *testing.T) {
	t.Parallel()
	suite.Run(t, &ConnTrafficTestSuite{})
}

func TestConnRewind(t *testing.T) {
	t.Parallel()
	suite.Run(t, &ConnRewindTestSuite{})
}
