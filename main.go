package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
)

func main() {
	app := cli.NewApp()

	app.Name = "todo"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "port",
			Value: 5000,
		},
	}
	app.Action = func(c *cli.Context) {
		o, err := parseOptions(c)
		if err != nil {
			log.Fatalln("Unable to parse options", err)
		}

		err = serve(o)
		if err != nil {
			log.Fatalln("Unable to parse options", err)
		}
	}

	app.Run(os.Args)
}

type options struct {
	Port int
}

func parseOptions(c *cli.Context) (*options, error) {
	port := c.Int("port")
	if port <= 0 || port > 65535 {
		return nil, errors.New("Invalid port number")
	}

	return &options{
		Port: port,
	}, nil
}

func serve(o *options) error {
	r := mux.NewRouter()

	f := &HandlerFactory{options: o}

	r.Methods("GET").Path("/todo").Handler(f.NewGetAllTodoHandler())
	r.Methods("GET").Path("/todo/{todo:[0-9]+}").Handler(f.NewGetTodoHandler())
	r.Methods("POST").Path("/todo").Handler(f.NewCreateTodoHandler())

	address := fmt.Sprintf(":%d", o.Port)
	log.Println("Listening on", address)

	return http.ListenAndServe(address, r)
}

type createTodoHandler struct{}

func (h *createTodoHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
	log.Println("got request on createTodoHandler")
}

type getTodoHandler struct{}

func (h *getTodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	log.Println("got request on getTodoHandler", vars["todo"])
}

type getAllTodoHandler struct{}

func (h *getAllTodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("got request on getAllTodoHandler")
}

type HandlerFactory struct {
	options *options
}

func (h *HandlerFactory) NewGetAllTodoHandler() *getAllTodoHandler {
	return &getAllTodoHandler{}
}

func (h *HandlerFactory) NewGetTodoHandler() *getTodoHandler {
	return &getTodoHandler{}
}

func (h *HandlerFactory) NewCreateTodoHandler() *createTodoHandler {
	return &createTodoHandler{}
}
