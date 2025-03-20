package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	socketio "github.com/googollee/go-socket.io"
)

var anather_ip string

type message_struct struct {
	Message string
}

func test(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	// log.Println(string(body))
	var message_str message_struct
	err = json.Unmarshal([]byte(body), &message_str)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\033[31m" + message_str.Message + "\033[0m" + "\n")
}

type ip_struct struct {
	Message string
}

func talk(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	// log.Println(string(body))
	var ip_str ip_struct
	err = json.Unmarshal([]byte(body), &ip_str)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\033[31m" + "Wanna talk with:" + ip_str.Message + "\033[0m" + "\n")
	fmt.Println("y/n")
	anather_ip = ip_str.Message
}

func GetLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}

func Send_request(url string, message string) {
	jsonStr := []byte(`{"message":"` + message + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func main() {
	var port string
	var ip string

	fmt.Println("1.HTTP post request ")
	fmt.Println("2.create websocket server ")
	fmt.Println("3.join websocket server ")
	var ch int
	fmt.Scan(&ch)
	if ch == 1 {
		fmt.Print("Port: ")
		fmt.Scan(&port)

		if port == "d" {
			port = "2525"
		}
		go func() {
			http.HandleFunc("/message", test)
			http.HandleFunc("/wana_talk", talk)
			http.ListenAndServe(":"+port, nil)
		}()
		my_ip := GetLocalIP()
		fmt.Println("Your ip addr:" + my_ip.String() + ":" + port)

		fmt.Println("type friend ip : ")
		fmt.Scan(&ip)
		if ip == "y" {
			ip = anather_ip
		} else if ip == "n" {
			fmt.Println("type friend ip : ")
			fmt.Scan(&ip)
		} else {
			Send_request("http://"+ip+"/wana_talk", my_ip.String()+":"+port)
		}

		fmt.Print("\033[H\033[2J")

		fmt.Println("Your ip addr:" + my_ip.String() + ":" + port)
		fmt.Println("connected to:", ip)
		reader := bufio.NewReader(os.Stdin)

		for true {

			message_from, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				continue
			}

			message_from = strings.TrimSpace(message_from) // Remove newline character

			Send_request("http://"+ip+"/message", message_from)

		}

	} else if ch == 2 {
		fmt.Print("Port: ")
		fmt.Scan(&port)

		if port == "d" {
			port = "8000"
		}
		server := socketio.NewServer(nil)

		server.OnConnect("/", func(s socketio.Conn) error {
			s.SetContext("")
			fmt.Println("connected:", s.ID())
			s.Join("bcast")
			return nil
		})

		server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
			fmt.Println(msg)
			server.ForEach("/", "bcast", func(c socketio.Conn) {
				if c.ID() != s.ID() { // Exclude sender
					color := "\033[31m"

					switch s.ID() {
					case "0":
						color = "\033[31m"
					case "1":
						color = "\033[34m"

					case "2":
						color = "\033[35m"
					case "3":
						color = "\033[36m"
					case "4":
						color = "\033[37m"
					case "5":
						color = "\033[97m"

					default:

						color = "\033[33m"

					}

					c.Emit("reply", msg, color)
				}
			})
		})

		server.OnError("/", func(s socketio.Conn, e error) {
			// server.Remove(s.ID())
			fmt.Println("meet error:", e)
		})

		server.OnDisconnect("/", func(s socketio.Conn, reason string) {
			// Add the Remove session id. Fixed the connection & mem leak
			server.Remove(s.ID())
			fmt.Println("disconnected: ", s.ID())
		})

		go server.Serve()
		defer server.Close()
		fmt.Print("\033[H\033[2J")
		http.Handle("/socket.io/", server)
		http.Handle("/", http.FileServer(http.Dir("./asset")))
		my_ip := GetLocalIP()
		log.Println("Serving at: " + my_ip.String() + ":" + port)
		log.Fatal(http.ListenAndServe(":"+port, nil))

	} else if ch == 3 {
		fmt.Print("type friend ip : ")
		fmt.Scan(&ip)

		//	uri := "http://127.0.0.1:8000"

		client, err := socketio.NewClient("http://"+ip, nil)
		if err != nil {
			panic(err)
		}

		// Handle an incoming event
		client.OnEvent("reply", func(s socketio.Conn, msg string, color string) {
			fmt.Printf(color + msg + "\033[0m " + "\n")
		})

		err = client.Connect()
		if err != nil {
			panic(err)
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("\033[H\033[2J")
		for true {

			message_from, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				continue
			}

			message_from = strings.TrimSpace(message_from) // Remove newline character

			client.Emit("notice", message_from)
		}

		time.Sleep(1 * time.Second)
		err = client.Close()
		if err != nil {
			panic(err)
		}
	}
}
