package external

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

func (e *External) initRemote() {
	err := e.h.InitRemote(e)
	if err != nil {
		fmt.Fprintf(e.out, "INITREMOTE-FAILURE %s\n", filterNewlines(err.Error()))
		return
	}

	fmt.Fprintf(e.out, "INITREMOTE-SUCCESS\n")
}

func (e *External) prepare() {
	err := e.h.Prepare(e)
	if err != nil {
		fmt.Fprintf(e.out, "PREPARE-FAILURE %v\n", filterNewlines(err.Error()))
		return
	}

	fmt.Fprintf(e.out, "PREPARE-SUCCESS\n")
}

func (e *External) store(key, file string) {
	err := e.h.Store(e, key, file)
	if err != nil {
		fmt.Fprintf(e.out, "TRANSFER-FAILURE STORE %s %s\n", key, filterNewlines(err.Error()))
		return
	}

	fmt.Fprintf(e.out, "TRANSFER-SUCCESS STORE %s\n", key)
}

func (e *External) retrieve(key, file string) {
	err := e.h.Retrieve(e, key, file)
	if err != nil {
		fmt.Fprintf(e.out, "TRANSFER-FAILURE RETRIEVE %s %s\n", key, filterNewlines(err.Error()))
		return
	}

	fmt.Fprintf(e.out, "TRANSFER-SUCCESS RETRIEVE %s\n", key)
}

func (e *External) checkPresent(key string) {
	found, err := e.h.CheckPresent(e, key)
	if err != nil {
		fmt.Fprintf(e.out, "CHECKPRESENT-UNKNOWN %s %s\n", key, filterNewlines(err.Error()))
		return
	}

	if found {
		fmt.Fprintf(e.out, "CHECKPRESENT-SUCCESS %s\n", key)
	} else {
		fmt.Fprintf(e.out, "CHECKPRESENT-FAILURE %s\n", key)
	}
}

func (e *External) remove(key string) {
	err := e.h.Remove(e, key)
	if err != nil {
		fmt.Fprintf(e.out, "REMOVE-FAILURE %s %s\n", key, filterNewlines(err.Error()))
		return
	}

	fmt.Fprintf(e.out, "REMOVE-SUCCESS %s\n", key)
}

func (e *External) getCost() {
	cost, err := e.h.GetCost(e)
	if err == ErrUnsupportedRequest {
		fmt.Fprintf(e.out, "UNSUPPORTED-REQUEST\n")
		return
	}
	if err != nil {
		e.Error(err.Error())
		return
	}

	fmt.Fprintf(e.out, "COST %d\n", cost)
}

func (e *External) getAvailability() {
	avail, err := e.h.GetAvailability(e)
	if err == ErrUnsupportedRequest {
		fmt.Fprintf(e.out, "UNSUPPORTED-REQUEST\n")
		return
	}
	if err != nil {
		e.Error(err.Error())
		return
	}

	switch avail {
	case AvailabilityGlobal:
		fmt.Fprintf(e.out, "AVAILABILITY GLOBAL\n")
	case AvailabilityLocal:
		fmt.Fprintf(e.out, "AVAILABILITY LOCAL\n")
	default:
		e.Error("GetAvailability returned an invalid value")
	}
}

func (e *External) whereIs(key string) {
	where, err := e.h.WhereIs(e, key)
	if err == ErrUnsupportedRequest {
		fmt.Fprintf(e.out, "UNSUPPORTED-REQUEST\n")
		return
	}
	if err != nil {
		e.Error(err.Error())
		return
	}

	if where == "" {
		fmt.Fprintf(e.out, "WHEREIS-FAILURE\n")
	} else {
		fmt.Fprintf(e.out, "WHEREIS-SUCCESS %s\n", filterNewlines(where))
	}
}

func (e *External) loop() (err error) {
	defer func() {
		if err != nil && !e.hasErrored {
			e.Error(err.Error())
		}
	}()

	fmt.Fprintf(e.out, "VERSION 1\n")

	for {
		if e.hasErrored {
			return nil
		}

		line, err := e.in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		line = strings.TrimRight(line, "\r\n")

		fields := strings.Split(line, " ")
		switch fields[0] {
		case "INITREMOTE":
			e.initRemote()

		case "PREPARE":
			e.prepare()

		case "TRANSFER":
			if len(fields) < 4 {
				return errors.New("less than 4 fields in TRANSFER")
			}

			switch fields[1] {
			case "STORE":
				e.store(fields[2], fields[3])
			case "RETRIEVE":
				e.retrieve(fields[2], fields[3])
			default:
				fmt.Fprintf(e.out, "UNSUPPORTED-REQUEST\n")
			}

		case "CHECKPRESENT":
			if len(fields) < 2 {
				return errors.New("less than 2 fields in CHECKPRESENT")
			}

			e.checkPresent(fields[1])

		case "REMOVE":
			if len(fields) < 2 {
				return errors.New("less than 2 fields in REMOVE")
			}

			e.remove(fields[1])

		case "GETCOST":
			e.getCost()

		case "GETAVAILABILITY":
			e.getAvailability()

		case "WHEREIS":
			if len(fields) < 2 {
				return errors.New("less than 2 fields in WHEREIS")
			}

			e.whereIs(fields[1])

		case "ERROR":
			e.hasErrored = true
			return errors.New(strings.Join(fields[1:], " "))

		default:
			fmt.Fprintf(e.out, "UNSUPPORTED-REQUEST\n")
		}
	}
}
