package memory

import (
	"io"
	"sync"

	"github.com/delgus/def-parser/internal/app"
)

// Queue реализует очередь безопасную для вызова в асинхронных потоках
type Queue struct {
	tasks []app.HostTask
	mu    sync.Mutex
}

// NewQueue вернет новую очередь
func NewQueue() *Queue {
	return &Queue{}
}

// Add реализует добавление в очередь
func (q *Queue) Add(task app.HostTask) {
	q.mu.Lock()
	q.tasks = append(q.tasks, task)
	q.mu.Unlock()
}

// Get - получение задачи из очереди
func (q *Queue) Get() (app.HostTask, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tasks) == 0 {
		return app.HostTask{}, io.EOF
	}
	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return task, nil
}
