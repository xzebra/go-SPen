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

type Config struct {
	ScreenWidth  float64 `json:"screen-width"`
	ScreenHeight float64 `json:"screen-height"`
}

var (
	upgrader     websocket.Upgrader
	config       Config
	DeviceWidth  float64
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
	heightRatio := float64(y) / DeviceHeight
	widthRatio := (DeviceWidth - float64(x)) / DeviceWidth //invert the coord
	mouse.Move(int(heightRatio*config.ScreenWidth), int(widthRatio*config.ScreenHeight))
}

func main() {
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/favicon.ico")
	})
	http.Handle("/", http.FileServer(http.Dir("public/")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		go func(conn *websocket.Conn) {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					fmt.Println("User disconnected")
					conn.Close()
					break
				}

				input := strings.Split(string(msg), ",")
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
				}
			}
		}(conn)
	})

	fmt.Println("Server running at localhost:8080")
	loadConfig()
	http.ListenAndServe("192.168.100.9:8080", nil)
}
