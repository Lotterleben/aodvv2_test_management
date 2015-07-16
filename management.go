package aodvv2_test_management

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
	improvements to be made
	-----------------------
	- make paths in setup_network() work on machines other than mine
	- enable use of other topologies than line4
*/

/* content_type */
const (
	CONTENT_TYPE_JSON  = iota
	CONTENT_TYPE_OTHER = iota
)

/* better safe than sorry */
const CHAN_BUF_SIZE = 500
const EXPECT_TIMEOUT = 5 //seconds

/* All channels for communication of a RIOT node */
type stream_channels struct {
	snd       chan string /* Send commands to the node */
	rcv_json  chan string /* Receive JSONs */
	rcv_other chan string /* Receive other stuff */
}

type Riot_info struct {
	Port     int
	Ip       string
	Channels stream_channels
}

const MAX_LINE_LEN = 10

const desvirt_path = "/home/lotte/riot/desvirt_mehlis/ports.list"

func check(e error) {
	if e != nil {
		fmt.Println("OMG EVERYBODY PANIC")
		panic(e)
	}
}

func check_str(s string, e error) {
	if e != nil {
		fmt.Println("OMG EVERYBODY PANIC")
		fmt.Println("Offending string: ", s)
		panic(e)
	}
}

/* Figure out the type of the content of a string */
func get_content_type(str string) int {
	json_template := make(map[string]interface{})
	err := json.Unmarshal([]byte(str), &json_template)

	if err == nil {
		return CONTENT_TYPE_JSON
	}
	return CONTENT_TYPE_OTHER
}

/* Set up the network. This will be switched to our own abstraction (hopefully soon). */
func setup_network() {
	fmt.Println("Setting up the network (this may take some seconds)...")
	/* Put together shell command which starts desvirt and our init script (TEMPORARY, FIXME) */
	shellstuff := "cd /home/lotte/aodvv2/aodvv2_demo &&" +
				  "make clean all &&" +
				  "cd /home/lotte/riot/desvirt_mehlis &&" +
				  /* kill line in case it's still running */
				  "./vnet -n line4 -q &&" +
				  /* restart network */
				  "./vnet -n line4 -s &&" +
				  "cd /home/lotte/aodvv2/vnet_tester &&" +
				  "./aodv_test.py -ds"

	fmt.Println(shellstuff)

	_, err := exec.Command("bash", "-c", shellstuff).Output()
	fmt.Printf("Errors:\n%s\n", err)
	fmt.Println("done.")
}

/* Create the log directory for experiment experiment_name and return the path towards it */
func setup_logdir_path(experiment_name string) string {
	t := time.Now()
	logdir_path := fmt.Sprintf("./logs/%s_%s", t.Format("02:Jan:06_15:04_MST"), experiment_name)
	err := os.Mkdir(logdir_path, 0776)
	check(err)

	return logdir_path
}

/* load port numbers, sorted by position, from the ports.info file behind path into info. */
func load_position_port_info_line(path string) (info []Riot_info) {
	var keys []int
	/* temporarily store info under wonky position numbers such as -47, -48, ... */
	tmp_map := make(map[int]Riot_info)

	file, err := os.Open(path)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pieces := strings.Split(scanner.Text(), ",")
		if len(pieces) != 2 {
			panic(fmt.Sprintf("Problematic line in ports.list: %s\n", scanner.Text()))
		}

		index_, err := strconv.Atoi(pieces[0])
		check(err)
		port_, err := strconv.Atoi(pieces[1])
		check(err)

		tmp_map[index_] = Riot_info{Port: port_}
	}

	/* Make sure everything went fine */
	if err := scanner.Err(); err != nil {
		check(err)
	}

	/* Since ports.list entries aren't sorted or with sensible position
	 * values, we'll have to rearrange and tweak them a bit */

	/* first, get and sort the indices (i.e. all the index_ values) */
	for key := range tmp_map {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	/* Then, according to the order we just created, add the value associated with
	 * each key to riot_info */
	for i := range keys {
		port_ := tmp_map[keys[i]]
		info = append(info, port_)
	}

	return info
}

/* Sort every line which arrives through reader onto one of the channels from
 * s, depending on the line's content. */
func (s stream_channels) sort_stream(conn *net.Conn, logger *log.Logger) {
	reader := bufio.NewReader(*conn)

	for {
		str, err := reader.ReadString('\n')
		check_str(str, err)

		if len(str) > 0 {
			if strings.HasPrefix(str, ">") {
				/* end of a shell command execution, create clean newline */
				s.rcv_other <- ">\n"
				/* remove > from str*/
				str = strings.TrimPrefix(str, "> ")
			}

			/* If there's something left, log, check line content and sort */
			if len(str) > 0 {
				logger.Print(str)
				switch get_content_type(str) {
				case CONTENT_TYPE_JSON:
					/* this line contains a JSON */
					s.rcv_json <- str
				case CONTENT_TYPE_OTHER:
					/* this line contains something else */
					s.rcv_other <- str
				}
			}
		}
	}
}

/* Send a command to the RIOT behind stream_channels. */
func (s stream_channels) Send(command string) {
	s.snd <- command
}

/* Look for string matching exp in the channels. If none is found,
 * print an error message. */
func (s stream_channels) Expect_JSON(expected_str string) {
	expected := make(map[string]interface{})
	received := make(map[string]interface{})

	err := json.Unmarshal([]byte(expected_str), &expected)
	check(err)

	success := make(chan bool, 1)
	received_log_str := ""

	go func(){
		for {
			received_str := <-s.rcv_json

			err := json.Unmarshal([]byte(received_str), &received)
			check(err)

			if wildcardedDeepEqual(expected, received) {
				/* This is the JSON we're looking for */
				success <- true
				return
			} else {
				/* log the JSON we found on our way for possible error reporting */
				received_log_str += received_str
			}
		}
	}()

	select {
		case <- success:
			/* JSON was found */
			fmt.Print(".")
			return
		case <-time.After(time.Second * EXPECT_TIMEOUT):
			/* call timed out and we didn't find our JSON :( */
			fmt.Printf("\nERROR:\nexpected:\n%s\n\nreceived "+
					   "(may be empty due to earlier errors):\n%s\n",
					   expected_str, received_log_str)
			return
	}
}

/* Look for string matching exp in the channels (TODO: use regex) */
func (s stream_channels) Expect_other(exp string) {
	for {
		content := <-s.rcv_other
		if content == exp {
			fmt.Print(".")
			return
		}
	}
}

/* Goroutine which takes care of the RIOT behind port at place index in the line */
func control_riot(index int, port int, wg *sync.WaitGroup, logdir_path string, riot_line *[]Riot_info) {
	logfile_path := fmt.Sprintf("%s/riot_%d_port_%d.log", logdir_path, index, port)
	logfile, err := os.Create(logfile_path)
	check(err)
	defer logfile.Close()

	logger := log.New(logfile, "", log.Lshortfile)

	conn, err := net.Dial("tcp", fmt.Sprint("localhost:", port))
	check(err)
	defer conn.Close()

	/* create channels and add them to the info about this thread stored in riot_line*/
	send_chan := make(chan string, CHAN_BUF_SIZE)  /* messages from the main routine */
	json_chan := make(chan string, CHAN_BUF_SIZE)  /* JSON messages from the RIOT */
	other_chan := make(chan string, CHAN_BUF_SIZE) /* other messages from the RIOT */
	channels := stream_channels{rcv_json: json_chan, rcv_other: other_chan, snd: send_chan}
	(*riot_line)[index].Channels = channels

	/*sort that stuff out*/
	go channels.sort_stream(&conn, logger)

	_, err = conn.Write([]byte("ifconfig\n"))
	check(err)

	/* find my IP address in the output */
	for {
		str := <-other_chan
		r, _ := regexp.Compile("inet6 addr: (fe80(.*))/.*scope: local")
		match := r.FindAllStringSubmatch(str, -1)

		if len(match) > 0 {
			(*riot_line)[index].Ip = match[0][1]
			fmt.Println(port, "/", index, ": my IP is", match[0][1])
			break
		}
	}

	/* Signal to main that we're ready to go */
	(*wg).Done()

	/* Read and handle any input from the outside (i.e. main)
	 * (just assume every message from main is a command for now) */
	for {
		message := <-send_chan

		if !strings.HasSuffix(message, "\n") {
			// make sure command ends with a newline
			message = fmt.Sprint(message, "\n")
		}

		logger.Print(message)
		_, err := conn.Write([]byte(message))
		check(err)
	}
}

/* Set up connections to all RIOTs  */
func connect_to_RIOTs(logdir_path string) []Riot_info {
	riot_line := load_position_port_info_line(desvirt_path)

	var wg sync.WaitGroup
	wg.Add(len(riot_line))

	for index, elem := range riot_line {
		go control_riot(index, elem.Port, &wg, logdir_path, &riot_line)
	}

	/* wait for all goroutines to finish setup before we go on */
	wg.Wait()
	return riot_line
}

/* create a clean slate:
 * restart all RIOTs and set up logging identifiable by experiment_id,
 * and return the topology that was built */
func Create_clean_setup(experiment_id string) []Riot_info {
	fmt.Println("Setting up clean RIOTs for ", experiment_id)
	setup_network()
	logdir_path := setup_logdir_path(experiment_id)
	riots := connect_to_RIOTs(logdir_path)
	fmt.Println("Setup done.\n")
	return riots
}
