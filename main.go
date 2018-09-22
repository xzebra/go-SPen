package main

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	control "github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
)

// Config is used to decode the config.json
// file into a Config variable
type Config struct {
	ScreenWidth  float64
	ScreenHeight float64
	Port         string
	IP           string
}

type finger struct {
	direction int
	swiping   bool
	initX     float64
	initY     float64
}

var (
	upgrader websocket.Upgrader
	pressing bool

	// config
	configScreenWidth  float64 = 1920
	configScreenHeight float64 = 1080
	configPort                 = "8080"
	configIP           string  // defined by the user

	// IPPattern is used by regexp to check if it is a valid IP
	IPPattern = "((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.|$)){4}"

	// DeviceWidth refers to the phone or tablet
	// screen width
	DeviceWidth float64
	// DeviceHeight refers to the phone or tablet
	// screen height
	DeviceHeight float64

	fingers [2]finger
)

func moveMouse(x, y float64) {
	// height ratio * screen width
	tx := int((y / DeviceHeight) * configScreenWidth)
	// width ratio (inverted) * screen width
	ty := int(((DeviceWidth - x) / DeviceWidth) * configScreenHeight)
	control.Move(tx, ty)
}

func setPressing(b bool) {
	pressing = b
	if b {
		control.MouseToggle("down")
	} else {
		control.MouseToggle("up")
	}
}

func zoom(dir string) {
	control.KeyToggle("control", "down")
	control.ScrollMouse(1, dir)
	control.KeyToggle("control", "up")
}

func swipe(direction int, initX float64, initY float64, id int) {
	fingers[id].direction = direction
	fingers[id].swiping = true
	fingers[id].initX = initX
	fingers[id].initY = initY
	if fingers[0].swiping && fingers[1].swiping { // both are swiping
		if (fingers[0].direction+fingers[1].direction)%2 == 0 {
			switch fingers[0].direction {
			case 4:
				if fingers[0].initY < fingers[1].initY {
					zoom("up")
				} else {
					zoom("down")
				}
			case 2:
				if fingers[0].initY > fingers[1].initY {
					zoom("up")
				} else {
					zoom("down")
				}
			case 1:
				if fingers[0].initX > fingers[1].initX {
					zoom("up")
				} else {
					zoom("down")
				}
			case 3:
				if fingers[0].initX < fingers[1].initX {
					zoom("up")
				} else {
					zoom("down")
				}
			}
		}
	}
}

func selectIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	var ips []string
	added := make(map[string]bool)
	fmt.Println("Choose your local IP\n===================")
	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			// ignore for now
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			}

			// check if ip is in the form 255.255.255.255
			if matched, err := regexp.Match(IPPattern, []byte(ip.String())); !matched || err != nil {
				continue
			}
			// Print and add to list
			if _, found := added[ip.String()]; !found {
				ips = append(ips, ip.String())
				added[ip.String()] = true
				fmt.Printf("[%d] %s\n", len(ips), ip.String())
			}
		}
	}

	return ""
}

func main() {
	fingers[0].direction = 0
	fingers[1].direction = 0

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("public", "static", "favicon.ico"))
	})

	http.HandleFunc("/static/spen-events.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("public", "static", "spen-events.js"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// filepath.Join to automatically use the separator your OS uses
		t, _ := template.ParseFiles(filepath.Join("public", "templates", "index.html"))
		if t != nil {
			t.Execute(w, Config{configScreenWidth, configScreenHeight, configIP, configPort})
		} else {
			fmt.Println("Error creating template")
		}
	})

	http.HandleFunc("/spen", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		go func(conn *websocket.Conn) {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					conn.Close()
					break
				}

				input := strings.Split(string(msg), ",")
				if len(input) == 2 {
					x, _ := strconv.ParseFloat(input[0], 64)
					y, _ := strconv.ParseFloat(input[1], 64)
					moveMouse(x, y)
				} else if input[0] == "pressing" {
					setPressing(true)
				} else if input[0] == "stoppressing" {
					setPressing(false)
				} else if input[0] == "screen" {
					// Init device screen size
					tempWidth, _ := strconv.Atoi(input[1])
					DeviceWidth = float64(tempWidth)
					tempHeight, _ := strconv.Atoi(input[2])
					DeviceHeight = float64(tempHeight)
				}
			}
		}(conn)
	})

	// This WebSocket will receive only the touch events from the finger
	http.HandleFunc("/finger", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		go func(conn *websocket.Conn) {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					conn.Close()
					break
				}
				//fmt.Println(string(msg))
				input := strings.Split(string(msg), ",")
				if len(input) == 5 {
					if input[0] == "swipe" {
						initX, _ := strconv.ParseFloat(input[1], 32)
						initY, _ := strconv.ParseFloat(input[2], 32)
						direction, _ := strconv.Atoi(input[3])
						id, _ := strconv.Atoi(input[4])
						swipe(direction, initX, initY, id)
					}
				} else if len(input) == 2 {
					if input[0] == "stop" {
						id, _ := strconv.Atoi(input[1])
						fingers[id].swiping = false
					}
				}
			}
		}(conn)
	})

	// iterate all network interfaces and let the user select it
	selectIP()
	// run the server
	fmt.Println("Server running at", configIP+":"+configPort)
	http.ListenAndServe(configIP+":"+configPort, nil)
}
