package mcast

const (
	MULTICAST = 1
	RECEIVE   = 2
)

type Command struct {
	Ctype uint8
}

func NewCommand(ctype uint8) *Command {
	return &Command{Ctype: ctype}
}
