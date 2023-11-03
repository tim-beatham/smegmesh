package conn

import (
	"errors"
	"slices"

	"github.com/tim-beatham/wgmesh/pkg/lib"
)

// ConnectionWindow maintains a sliding window of connections between users
type ConnectionWindow interface {
	// GetWindow is a list of connections to choose from
	GetWindow() []string
	// SlideConnection removes a node from the window and adds a random node
	// not already in the window. connList represents the list of possible
	// connections to choose from
	SlideConnection(connList []string) error
	// PushConneciton is used when connection list less than window size.
	PutConnection(conn []string) error
	// IsFull returns true if the window is full. In which case we must slide the window
	IsFull() bool
}

type ConnectionWindowImpl struct {
	window     []string
	windowSize int
}

// GetWindow gets the current list of active connections in
// the window
func (c *ConnectionWindowImpl) GetWindow() []string {
	return c.window
}

// SlideConnection slides the connection window by one shuffling items
// in the windows
func (c *ConnectionWindowImpl) SlideConnection(connList []string) error {
	// If the number of peer connections is less than the length of the window
	// then exit early. Can't slide the window it should contain all nodes!
	if len(c.window) < c.windowSize {
		return nil
	}

	filter := func(node string) bool {
		return !slices.Contains(c.window, node)
	}

	pool := lib.Filter(connList, filter)
	newNode := lib.RandomSubsetOfLength(pool, 1)

	if len(newNode) == 0 {
		return errors.New("could not slide window")
	}

	for i := len(c.window) - 1; i >= 1; i-- {
		c.window[i] = c.window[i-1]
	}

	c.window[0] = newNode[0]
	return nil
}

// PutConnection put random connections in the connection
func (c *ConnectionWindowImpl) PutConnection(connList []string) error {
	if len(c.window) >= c.windowSize {
		return errors.New("cannot place connection. Window full need to slide")
	}

	c.window = lib.RandomSubsetOfLength(connList, c.windowSize)
	return nil
}

func (c *ConnectionWindowImpl) IsFull() bool {
	return len(c.window) >= c.windowSize
}

func NewConnectionWindow(windowLength int) ConnectionWindow {
	window := &ConnectionWindowImpl{
		window:     make([]string, 0),
		windowSize: windowLength,
	}

	return window
}
