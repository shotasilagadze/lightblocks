package common

// define the command received from clients. We assume the format
// from the client is correct and both keys and values are non empty strings
type Command struct {
	Command string
	Values  []string
}
