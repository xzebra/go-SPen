package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	mouse "github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
)

// Config is used to decode the config.json
// file into a Config variable
type Config struct {
	ScreenWidth  float64 `json:"screen-width"`
	ScreenHeight float64 `json:"screen-height"`
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

func moveMouse(x, y int) {
	// height ratio * screen width
	tx := int((float64(y) / DeviceHeight) * config.ScreenWidth)
	// width ratio (inverted) * screen width
	ty := int(((DeviceWidth - float64(x)) / DeviceWidth) * config.ScreenHeight)
	mouse.Move(tx, ty)
}

func setPressing(b bool) {
	pressing = b
	if b {
		mouse.MouseToggle("down")
	} else {
		mouse.MouseToggle("up")
	}
}

func main() {
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
					x, _ := strconv.Atoi(input[0])
					y, _ := strconv.Atoi(input[1])
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
				_, _, err := conn.ReadMessage()
				if err != nil {
					conn.Close()
					break
				}

				/*input := strings.Split(string(msg), ",")
				if len(input) == 2 {
					x, _ := strconv.Atoi(input[0])
					y, _ := strconv.Atoi(input[1])
					moveMouse(x, y)
				} else if input[0] == "screen" {
					// Init device screen size
					tempWidth, _ := strconv.Atoi(input[1])
					DeviceWidth = float64(tempWidth)
					tempHeight, _ := strconv.Atoi(input[2])
					DeviceHeight = float64(tempHeight)
				}*/
			}
		}(conn)
	})

	fmt.Println("Server running at localhost:8080")
	loadConfig()
	http.ListenAndServe("192.168.100.9:8080", nil)
}
