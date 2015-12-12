package external

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrUnsupportedRequest = errors.New("unsupported request")

type Availability int

const (
	AvailabilityGlobal Availability = iota
	AvailabilityLocal
)

type ExternalHandler interface {
	// InitRemote, Prepare, Store, Retrieve, CheckPresent, and Remove must be
	// implemented.
	//
	// GetCost, GetAvailability, and WhereIs may return ErrUnsupportedRequest if
	// the ExternalHandler does not want to implement them.

	InitRemote(e *External) error
	Prepare(e *External) error

	Store(e *External, key, file string) error
	Retrieve(e *External, key, file string) error
	Remove(e *External, key string) error

	// If CheckPresent returns a non-nil error, External will reply UNKNOWN to
	// git-annex.
	CheckPresent(e *External, key string) (found bool, err error)

	GetCost(e *External) (cost int, err error)
	GetAvailability(e *External) (Availability, error)

	// If WhereIs returns an empty string with a nil error, then External will
	// indicate to git-annex that no location is known for that key.
	WhereIs(e *External, key string) (string, error)
}

type External struct {
	in         *bufio.Reader
	out        io.Writer
	h          ExternalHandler
	hasErrored bool
}

func RunLoop(in io.Reader, out io.Writer, h ExternalHandler) error {
	e := &External{
		in:  bufio.NewReader(in),
		out: out,
		h:   h,
	}
	return e.loop()
}

func (e *External) GetConfig(name string) (string, error) {
	fmt.Fprintf(e.out, "GETCONFIG %s\n", filterNewlines(name))
	return e.readValue()
}

func (e *External) SetConfig(name, value string) error {
	fmt.Fprintf(e.out, "SETCONFIG %s %s\n", filterNewlines(name), filterNewlines(value))
	return nil
}

func (e *External) Progress(location int64) {
	fmt.Fprintf(e.out, "PROGRESS %d\n", location)
}

func (e *External) DirHash(key string) (string, error) {
	fmt.Fprintf(e.out, "DIRHASH %s\n", filterNewlines(key))
	return e.readValue()
}

func (e *External) GetUUID() (string, error) {
	fmt.Fprintf(e.out, "GETUUID\n")
	return e.readValue()
}

func (e *External) GetGitDir() (string, error) {
	fmt.Fprintf(e.out, "GETGITDIR\n")
	return e.readValue()
}

func (e *External) SetState(key, value string) error {
	fmt.Fprintf(e.out, "SETSTATE %s %s\n", filterNewlines(key), filterNewlines(value))
	return nil
}

func (e *External) GetState(key string) (string, error) {
	fmt.Fprintf(e.out, "GETSTATE %s\n", filterNewlines(key))
	return e.readValue()
}

func (e *External) SetURLPresent(key, url string) error {
	fmt.Fprintf(e.out, "SETURLPRESENT %s %s\n", filterNewlines(key), filterNewlines(url))
	return nil
}

func (e *External) SetURLMissing(key, url string) error {
	fmt.Fprintf(e.out, "SETURLMISSING %s %s\n", filterNewlines(key), filterNewlines(url))
	return nil
}

func (e *External) SetURIPresent(key, uri string) error {
	fmt.Fprintf(e.out, "SETURIPRESENT %s %s\n", filterNewlines(key), filterNewlines(uri))
	return nil
}

func (e *External) SetURIMissing(key, uri string) error {
	fmt.Fprintf(e.out, "SETURIMISSING %s %s\n", filterNewlines(key), filterNewlines(uri))
	return nil
}

func (e *External) GetURLs(key, prefix string) ([]string, error) {
	fmt.Fprintf(e.out, "GETURLS %s %s\n", filterNewlines(key), filterNewlines(prefix))

	out := make([]string, 0, 4)
	for {
		value, err := e.readValue()
		if err != nil {
			return nil, err
		}
		if value == "" {
			break
		}
		out = append(out, value)
	}

	return out, nil
}

func (e *External) Debug(message string) {
	fmt.Fprintf(e.out, "DEBUG %s\n", filterNewlines(message))
}

func (e *External) Error(message string) {
	fmt.Fprintf(e.out, "ERROR %s\n", filterNewlines(message))
	e.hasErrored = true
}

func (e *External) readValue() (string, error) {
	line, err := e.in.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(line, "\n")

	if strings.HasPrefix(line, "VALUE ") {
		return strings.TrimPrefix(line, "VALUE "), nil
	}

	return "", errors.New("protocol error: expected VALUE")
}
