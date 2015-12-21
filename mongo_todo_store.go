package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	CollectionName = "todo"
)

type mgoTodo struct {
	ID     bson.ObjectId `bson:"_id"`
	Title  string        `bson:"title"`
	Status string        `bson:"status"`
}

func (t *mgoTodo) ToTodo() *Todo {
	return &Todo{
		ID:     t.ID.Hex(),
		Title:  t.Title,
		Status: t.Status,
	}
}

func fromTodo(t *Todo, id bson.ObjectId) *mgoTodo {
	if id == "" {
		id = bson.ObjectIdHex(t.ID)
	}

	return &mgoTodo{
		ID:     id,
		Title:  t.Title,
		Status: t.Status,
	}
}

func NewMongoTodoStore(db *mgo.Database) (*MongoTodoStore, error) {
	return &MongoTodoStore{
		db: db,
	}, nil
}

type MongoTodoStore struct {
	db *mgo.Database
}

func (s *MongoTodoStore) Create(t *Todo) (string, error) {
	id := bson.NewObjectId()
	doc := fromTodo(t, id)

	s.db.C(CollectionName).Insert(doc)

	return id.Hex(), nil
}

func (s *MongoTodoStore) Get(id string) (*Todo, error) {
	var t *mgoTodo
	err := s.db.C(CollectionName).FindId(bson.ObjectIdHex(id)).One(&t)

	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	return t.ToTodo(), nil
}

func (s *MongoTodoStore) Update(t *Todo) error {
	id := bson.ObjectIdHex(t.ID)
	return s.db.C(CollectionName).UpdateId(id, fromTodo(t, id))
}

func (s *MongoTodoStore) Delete(id string) error {
	return s.db.C(CollectionName).RemoveId(bson.ObjectIdHex(id))
}

func (s *MongoTodoStore) GetAll() ([]*Todo, error) {
	results := []*mgoTodo{}
	err := s.db.C(CollectionName).Find(nil).All(&results)

	buf := make([]*Todo, len(results), len(results))

	for i, todo := range results {
		buf[i] = todo.ToTodo()
	}

	return buf, err
}
