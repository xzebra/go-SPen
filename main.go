package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	control "github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
)

// Config is used to decode the config.json
// file into a Config variable
type Config struct {
	ScreenWidth  float64 `json:"screen-width"`
	ScreenHeight float64 `json:"screen-height"`
	IP           string  `json:"ip"`
	Port         string  `json:"port"`
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
	config   Config
	// DeviceWidth refers to the phone or tablet
	// screen width
	DeviceWidth float64
	// DeviceHeight refers to the phone or tablet
	// screen height
	DeviceHeight float64

	fingers [2]finger
)

func loadConfig() {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&config)
	if err != nil {
		panic(err)
	}
}

func moveMouse(x, y float64) {
	// height ratio * screen width
	tx := int((y / DeviceHeight) * config.ScreenWidth)
	// width ratio (inverted) * screen width
	ty := int(((DeviceWidth - x) / DeviceWidth) * config.ScreenHeight)
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

func main() {
	fingers[0].direction = 0
	fingers[1].direction = 0

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/favicon.ico")
	})
	http.Handle("/", http.FileServer(http.Dir("public/")))

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

	fmt.Println("Server running at", config.IP + ":" + config.Port)
	loadConfig()
	http.ListenAndServe(config.IP + ":" + Config.Port, nil)
}
