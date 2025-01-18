package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/NicolasLopes7/ws/protocol"
)

func WSHandle(w http.ResponseWriter, r *http.Request) {
	ws, err := protocol.New(w, r)
	if err != nil {
		fmt.Println("Error creating websocket:", err)
		return
	}
	radio.AddConn(ws)

	fmt.Printf("New connection by client: %s\n", ws.Conn.RemoteAddr())
	defer func() {
		ws.Close()
		radio.RemoveConn(ws)
	}()

	for {
		frame, err := ws.Recv()
		fmt.Println(frame)
		if err != nil {
			fmt.Println("Error receiving frame:", err)
			return
		}

		if frame.Opcode == 8 {
			fmt.Printf("Connection closed by client: %s\n", ws.Conn.RemoteAddr())
			return
		}

		if frame.Opcode == 9 {
			fmt.Printf("Ping received by client: %s\n", ws.Conn.RemoteAddr())
			frame.Opcode = 10
		}

		radio.Broadcast(ws.Conn.RemoteAddr().String(), frame)
	}
}

func main() {
	http.HandleFunc("/ws", WSHandle)
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public", http.StripPrefix("/public", fs))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
}

type Radio struct {
	conns map[string]*protocol.Ws
}

func (r *Radio) Broadcast(whoami string, frame protocol.Frame) {
	for addr, conn := range r.conns {
		if addr == whoami {
			continue
		}

		if err := conn.Send(frame); err != nil {
			fmt.Println("Error sending", err)
		}
	}
}

func (r *Radio) AddConn(ws *protocol.Ws) {
	r.conns[ws.Conn.RemoteAddr().String()] = ws
}

func (r *Radio) RemoveConn(ws *protocol.Ws) {
	delete(r.conns, ws.Conn.RemoteAddr().String())
}

var radio = Radio{conns: make(map[string]*protocol.Ws)}
