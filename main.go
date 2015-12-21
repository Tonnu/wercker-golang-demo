package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"

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

	f := &HandlerFactory{options: o, store: NewMemoryTodoStore()}

	r.Methods("GET").Path("/todo").Handler(f.NewGetAllTodoHandler())

	r.Methods("POST").Path("/todo").Handler(f.NewCreateTodoHandler())
	r.Methods("GET").Path("/todo/{todo:[0-9]+}").Handler(f.NewGetTodoHandler())
	r.Methods("PUT").Path("/todo/{todo:[0-9]+}").Handler(f.NewUpdateTodoHandler())
	r.Methods("DELETE").Path("/todo/{todo:[0-9]+}").Handler(f.NewDeleteTodoHandler())

	address := fmt.Sprintf(":%d", o.Port)
	log.Println("Listening on", address)

	return http.ListenAndServe(address, r)
}

type createTodoHandler struct {
	store TodoStore
}

func (h *createTodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("unable to read body:", err)
		http.Error(w, "unable to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var t Todo
	err = json.Unmarshal(payload, &t)
	if err != nil {
		log.Println("unable to unmarshal body:", err)
		http.Error(w, "unable to unmarshal body", http.StatusInternalServerError)
		return
	}

	id, err := h.store.Create(&t)
	if err != nil {
		log.Println("unable to create todo:", err)
		http.Error(w, "unable to create todo", http.StatusInternalServerError)
		return
	}

	log.Println("created todo", id)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}

type getTodoHandler struct {
	store TodoStore
}

func (h *getTodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["todo"]
	if id == "" {
		http.Error(w, "no id specified", http.StatusInternalServerError)
		return
	}

	t, err := h.store.Get(id)
	if err != nil {
		log.Println("unable to retrieve todo:", err)
		http.Error(w, "unable to retrieve todo", http.StatusInternalServerError)
		return
	}

	if t == nil {
		http.NotFound(w, r)
		return
	}

	p, err := json.Marshal(t)
	if err != nil {
		log.Println("unable to retrieve todo:", err)
		http.Error(w, "unable to retrieve todo", http.StatusInternalServerError)
		return
	}

	w.Write(p)
}

type updateTodoHandler struct {
	store TodoStore
}

func (h *updateTodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["todo"]
	if id == "" {
		http.Error(w, "no id specified", http.StatusInternalServerError)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("unable to read body:", err)
		http.Error(w, "unable to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var t Todo
	err = json.Unmarshal(payload, &t)
	if err != nil {
		log.Println("unable to unmarshal body:", err)
		http.Error(w, "unable to unmarshal body", http.StatusInternalServerError)
		return
	}

	t.ID = id

	err = h.store.Update(&t)
	if err != nil {
		log.Println("unable to update todo:", err)
		http.Error(w, "unable to update todo", http.StatusInternalServerError)
		return
	}
}

type deleteTodoHandler struct {
	store TodoStore
}

func (h *deleteTodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["todo"]
	if id == "" {
		http.Error(w, "no id specified", http.StatusInternalServerError)
		return
	}

	err := h.store.Delete(id)
	if err != nil {
		log.Println("unable to delete todo", err)
		http.Error(w, "unable to delete todo", http.StatusInternalServerError)
		return
	}

	log.Println("deleted todo", id)
	w.WriteHeader(http.StatusOK)
}

type getAllTodoHandler struct {
	store TodoStore
}

func (h *getAllTodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t, err := h.store.GetAll()
	if err != nil {
		log.Println("unable to retrieve todos", err)
		http.Error(w, "unable to retrieve todos", http.StatusInternalServerError)
		return
	}

	if t == nil {
		http.NotFound(w, r)
		return
	}

	p, err := json.Marshal(t)
	if err != nil {
		log.Println("unable to marshal todos", err)
		http.Error(w, "unable to marshal todos", http.StatusInternalServerError)
		return
	}

	w.Write(p)
}

type HandlerFactory struct {
	options *options
	store   TodoStore
}

func (h *HandlerFactory) NewCreateTodoHandler() *createTodoHandler {
	return &createTodoHandler{store: h.store}
}

func (h *HandlerFactory) NewGetTodoHandler() *getTodoHandler {
	return &getTodoHandler{store: h.store}
}

func (h *HandlerFactory) NewUpdateTodoHandler() *updateTodoHandler {
	return &updateTodoHandler{store: h.store}
}

func (h *HandlerFactory) NewDeleteTodoHandler() *deleteTodoHandler {
	return &deleteTodoHandler{store: h.store}
}

func (h *HandlerFactory) NewGetAllTodoHandler() *getAllTodoHandler {
	return &getAllTodoHandler{store: h.store}
}

type Todo struct {
	ID     string `json:"id,omitempty"`
	Title  string `json:"title,omitempty"`
	Status bool   `json:"status,omitempty"`
}

type TodoStore interface {
	Create(t *Todo) (id string, err error)
	Get(id string) (*Todo, error)
	Update(t *Todo) error
	Delete(id string) error

	GetAll() ([]*Todo, error)
}

func NewMemoryTodoStore() *MemoryTodoStore {
	counter := int32(0)
	return &MemoryTodoStore{
		counter: &counter,
		s:       make(map[string]*Todo),
	}
}

type MemoryTodoStore struct {
	s map[string]*Todo

	counter *int32
}

func (s *MemoryTodoStore) createID() string {
	id := atomic.AddInt32(s.counter, 1)

	return strconv.Itoa(int(id))
}

func (s *MemoryTodoStore) Create(t *Todo) (id string, err error) {
	id = s.createID()
	t.ID = id

	s.s[id] = t

	return id, nil
}

func (s *MemoryTodoStore) Get(id string) (*Todo, error) {
	if t, exists := s.s[id]; exists {
		return t, nil
	}

	return nil, nil
}

func (s *MemoryTodoStore) Update(t *Todo) error {
	if _, exists := s.s[t.ID]; exists {
		s.s[t.ID] = t
		return nil
	}

	return errors.New("TODO not found")
}

func (s *MemoryTodoStore) Delete(id string) error {
	delete(s.s, id)

	return nil
}

func (s *MemoryTodoStore) GetAll() ([]*Todo, error) {
	buf := []*Todo{}
	for _, todo := range s.s {
		buf = append(buf, todo)
	}
	return buf, nil
}
