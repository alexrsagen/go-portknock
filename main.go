package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	var err error
	var laddr, raddr net.Addr
	var conn net.Conn
	var ports []int
	var sleep time.Duration

	knockType := flag.String("t", "seed", "knock type (seed/seq)")
	seed := flag.Int64("s", 0, "port generation seed, used when t=seed")
	count := flag.Int("c", 1, "port count, used when t=seed")
	offset := flag.Int("o", 30000, "port offset, used when t=seed")
	seq := flag.String("q", "", "comma-separated custom port sequence, used when t=seq")
	proto := flag.String("p", "udp", "protocol (tcp/tcp4/tcp6/udp/udp4/udp6)")
	local := flag.String("l", "", "local address")
	remote := flag.String("r", "", "remote address")
	interval := flag.String("i", "", "loop interval")
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()

	// get ports
	switch *knockType {
	case "seed":
		if *count < 1 {
			fmt.Println("port sequence empty")
			return
		}
		if *offset < 0 || *offset > 65535 {
			fmt.Println("offset should be between 0 and 65535")
			return
		}
		rand.Seed(*seed)
		ports = make([]int, *count)
		for i := 0; i < *count; i++ {
			ports[i] = *offset + rand.Intn(65535-*offset)
		}
	case "seq":
		parts := strings.Split(*seq, ",")
		for _, part := range parts {
			port, err := strconv.Atoi(part)
			if err != nil {
				fmt.Printf("invalid port \"%s\": %v\n", part, err)
				return
			}
			ports = append(ports, port)
		}
	default:
		fmt.Println("invalid knock type")
		return
	}

	if len(ports) == 0 {
		fmt.Println("port sequence empty")
		return
	}

	// print ports
	if *verbose {
		fmt.Printf("ports: ")
		for i := 0; i < len(ports); i++ {
			portstr := strconv.Itoa(ports[i])
			if len(portstr) < 6 {
				portstr += strings.Repeat(" ", 6-len(portstr))
			}
			fmt.Printf("%s", portstr)
		}
		fmt.Printf("\n")
	}

	// get protocol, resolve addresses
	switch *proto {
	case "tcp", "tcp4", "tcp6":
		if laddr, err = net.ResolveTCPAddr(*proto, *local+":0"); err != nil {
			fmt.Println(err)
			return
		}
		if raddr, err = net.ResolveTCPAddr(*proto, *remote+":0"); err != nil {
			fmt.Println(err)
			return
		}
		if raddr.(*net.TCPAddr).IP == nil {
			fmt.Println("unable to resolve remote IP")
			return
		}
	case "udp", "udp4", "udp6":
		if laddr, err = net.ResolveUDPAddr(*proto, *local+":0"); err != nil {
			fmt.Println(err)
			return
		}
		if raddr, err = net.ResolveUDPAddr(*proto, *remote+":0"); err != nil {
			fmt.Println(err)
			return
		}
		if raddr.(*net.UDPAddr).IP == nil {
			fmt.Println("unable to resolve remote IP")
			return
		}
	default:
		fmt.Println("invalid protocol")
	}

	// get sleep interval
	if *interval != "" {
		if sleep, err = time.ParseDuration(*interval); err != nil {
			fmt.Println(err)
			return
		}
	}

	for {
		if *verbose {
			fmt.Printf("knock: ")
		}
		for i := 0; i < len(ports); i++ {
			switch v := raddr.(type) {
			case *net.TCPAddr:
				v.Port = ports[i]
				if conn, err = net.DialTCP(*proto, laddr.(*net.TCPAddr), v); err != nil {
					fmt.Println("dial fail")
					return
				}
			case *net.UDPAddr:
				v.Port = ports[i]
				if conn, err = net.DialUDP(*proto, laddr.(*net.UDPAddr), v); err != nil {
					fmt.Println("dial fail")
					return
				}
			}
			if _, err = conn.Write([]byte{}); err != nil {
				fmt.Println("tx fail")
				return
			}
			if err = conn.Close(); err != nil {
				fmt.Println("close fail")
				return
			}
			if *verbose {
				fmt.Printf("ok    ")
			}
			time.Sleep(100 * time.Millisecond)
		}
		if *verbose {
			fmt.Printf("\n")
		}
		if sleep == 0 {
			break
		}
		time.Sleep(sleep)
	}
}
