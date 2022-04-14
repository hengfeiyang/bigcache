package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hengfeiyang/bigcache"
)

const defaultDapacity = 10000
const defaultShards = 32
const defaultMaxEntrySize = 1024 * 1024 // 1M

var ErrorCommandNotFound = errors.New("command not found")

const help = `Command not found:
DB n                   select db of version, support: 1,2,3,4,5, default is 4
SET key value [TTL]    set a value by the key and optional ttl, unit is: time.Second
GET key                get a value by the key
TTL key                get the remain ttl by the key
DEL key                delete the key
LEN                    get total entries of current db
`

var dbs []bigcache.Cacher
var currentDB = 4

type command struct {
	Command string
	Key     string
	Value   string
	TTL     time.Duration
	DB      int
}

func main() {
	dbs = make([]bigcache.Cacher, 10)
	dbs[1] = bigcache.NewCacheV1(defaultDapacity)
	dbs[2] = bigcache.NewCacheV2(defaultDapacity)
	dbs[3] = bigcache.NewCacheV3(defaultDapacity)
	dbs[4] = bigcache.NewCacheV4(defaultDapacity, defaultShards)
	dbs[5] = bigcache.NewCacheV5(defaultDapacity, defaultMaxEntrySize, defaultShards)

	fmt.Println("bigcache cli v0.0.1")
	reader := bufio.NewReader(os.Stdin)
	var output string
	for {
		fmt.Printf("db[%d]> ", currentDB)
		s, _, _ := reader.ReadLine()
		cmd, err := parseCommand(string(s))
		if err != nil {
			if err == ErrorCommandNotFound {
				fmt.Print(help)
			} else {
				fmt.Println(err)
			}
			continue
		}

		switch cmd.Command {
		case "DB":
			output, err = cmd.db()
		case "SET":
			output, err = cmd.set()
		case "GET":
			output, err = cmd.get()
		case "TTL":
			output, err = cmd.ttl()
		case "DEL":
			output, err = cmd.delete()
		case "LEN":
			output, err = cmd.len()
		}
		if err != nil {
			fmt.Println(err)
		} else {
			if len(output) > 0 {
				fmt.Println(output)
			}
		}
	}
}

func parseCommand(s string) (*command, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrorCommandNotFound
	}

	column := strings.Split(s, " ")
	cmd := new(command)
	cmd.Command = strings.ToUpper(column[0])
	switch cmd.Command {
	case "DB":
		if len(column) < 2 {
			return nil, fmt.Errorf("DB n")
		}
		cmd.DB, _ = strconv.Atoi(column[1])
		if cmd.DB > 5 {
			return nil, fmt.Errorf("db range is [0-5]")
		}
	case "SET":
		if len(column) < 3 {
			return nil, fmt.Errorf("SET key value [TTL], unit is: time.Second")
		}
		cmd.Key = column[1]
		cmd.Value = column[2]
		if len(column) == 4 {
			ttl, _ := strconv.Atoi(column[3])
			cmd.TTL = time.Duration(ttl) * time.Second
		}
	case "GET":
		if len(column) < 2 {
			return nil, fmt.Errorf("GET key")
		}
		cmd.Key = column[1]
	case "TTL":
		if len(column) < 2 {
			return nil, fmt.Errorf("TTL key")
		}
		cmd.Key = column[1]
	case "DEL":
		if len(column) < 2 {
			return nil, fmt.Errorf("DEL key")
		}
		cmd.Key = column[1]
	case "LEN":
		// noop
	default:
		return nil, ErrorCommandNotFound
	}

	return cmd, nil
}

func (t *command) db() (string, error) {
	switch t.DB {
	case 1, 2, 3, 4, 5:
		currentDB = t.DB
	default:
		currentDB = 4
	}
	return "", nil
}

func (t *command) set() (string, error) {
	err := dbs[currentDB].Set(t.Key, []byte(t.Value), t.TTL)
	return "OK", err
}

func (t *command) get() (string, error) {
	val, err := dbs[currentDB].Get(t.Key)
	return string(val), err
}

func (t *command) ttl() (string, error) {
	ttl, err := dbs[currentDB].TTL(t.Key)
	if err != nil {
		return "", err
	}
	if ttl > 0 {
		ttl /= time.Second
	}
	return fmt.Sprintf("(integer) %d", ttl), nil
}

func (t *command) delete() (string, error) {
	dbs[currentDB].Delete(t.Key)
	return fmt.Sprintf("(integer) %d", 1), nil
}

func (t *command) len() (string, error) {
	n := dbs[currentDB].Len()
	return strconv.Itoa(n), nil
}
