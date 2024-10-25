package app

import (
	"log"
	"messengerexample/internal/chatroom"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type App struct {
	server   http.Server
	upgrader websocket.Upgrader
	pool     *chatroom.Pool
	lastid   int64
}

func NewApp() *App {

	app := &App{
		server: http.Server{
			Addr: "localhost:8080",
		},
		upgrader: websocket.Upgrader{},
		pool:     chatroom.NewPool(),
		lastid:   0,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /chat/{id}", app.handler)

	app.server.Handler = mux
	return app
}

func (a *App) Run() {
	panic(a.server.ListenAndServe())
}

func (a *App) handler(w http.ResponseWriter, r *http.Request) {
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "idi naxui", http.StatusBadRequest)
		return
	}

	cr := a.pool.GetOrSet(id)
	atomic.AddInt64(&a.lastid, 1)
	cr.AddClient(int(a.lastid), conn)
}
