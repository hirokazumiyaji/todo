package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/hirokazumiyaji/todo/proto"
)

var (
	TodoValidationError = errors.New("text is required")
	TodoNotFoundError   = errors.New("todo is not found")
)

type Todo struct {
	ID        string
	Text      string
	DonedAt   *time.Time
	CreatedAt time.Time
}

func newTodo(text string) *Todo {
	return &Todo{
		ID:        uuid.New().String(),
		Text:      text,
		CreatedAt: time.Now(),
	}
}

func (t *Todo) Done() {
	n := time.Now()
	t.DonedAt = &n
}

func newProtoTodo(t *Todo) *proto.Todo {
	res := &proto.Todo{
		Id:        t.ID,
		Text:      t.Text,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
	}
	if t.DonedAt != nil {
		res.DonedAt = &wrappers.StringValue{
			Value: t.DonedAt.Format(time.RFC3339),
		}
	}
	return res
}

type TodoRepository interface {
	ListAll() []*Todo
	Get(string) (*Todo, error)
	Create(string) (*Todo, error)
	Update(*Todo) (*Todo, error)
}

type todoRepository struct {
	m  map[string]*Todo
	mu sync.RWMutex
}

func newTodoRepository() TodoRepository {
	return &todoRepository{
		m:  make(map[string]*Todo),
		mu: sync.RWMutex{},
	}
}

func (r *todoRepository) ListAll() []*Todo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	todos := make([]*Todo, 0, len(r.m))
	for _, v := range r.m {
		todos = append(todos, v)
	}
	return todos
}

func (r *todoRepository) Get(id string) (*Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	todo, ok := r.m[id]
	if !ok {
		return nil, TodoNotFoundError
	}
	return todo, nil
}

func (r *todoRepository) Create(text string) (*Todo, error) {
	if text == "" {
		return nil, TodoValidationError
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	todo := newTodo(text)
	r.m[todo.ID] = todo
	return todo, nil
}

func (r *todoRepository) Update(t *Todo) (*Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[t.ID] = t
	return t, nil
}

type TodoServiceServer struct {
	r TodoRepository
}

func NewTodoServiceServer() *TodoServiceServer {
	return &TodoServiceServer{r: newTodoRepository()}
}

func (s *TodoServiceServer) ListTodo(ctx context.Context, r *proto.ListTodoRequest) (*proto.ListTodoResponse, error) {
	todos := s.r.ListAll()
	res := &proto.ListTodoResponse{Todos: make([]*proto.Todo, 0)}
	for _, t := range todos {
		res.Todos = append(res.Todos, newProtoTodo(t))
	}
	return res, nil
}

func (s *TodoServiceServer) CreateTodo(ctx context.Context, r *proto.CreateTodoRequest) (*proto.Todo, error) {
	todo, err := s.r.Create(r.Text)
	if err != nil {
		return nil, err
	}
	return newProtoTodo(todo), nil
}

func (s *TodoServiceServer) DoneTodo(ctx context.Context, r *proto.DoneTodoRequest) (*proto.Todo, error) {
	todo, err := s.r.Get(r.TodoId)
	if err != nil {
		return nil, err
	}
	todo.Done()
	todo, err = s.r.Update(todo)
	if err != nil {
		return nil, err
	}
	return newProtoTodo(todo), nil
}
