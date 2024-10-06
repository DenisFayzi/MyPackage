package main

import (
	"container/list"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Структура для хранения очередей
// карта для очереди и для ожидания сообщений

type BrokerQueue struct {
	queues map[string]*list.List
	conds  map[string]*sync.Cond
	mu     sync.Mutex
}

// конструктор

func NewQueueService() *BrokerQueue {
	return &BrokerQueue{
		queues: make(map[string]*list.List),
		conds:  make(map[string]*sync.Cond),
	}
}

// Необходимо создать очередь новую
func (b *BrokerQueue) getQueue(nameQueue string) (*list.List, *sync.Cond) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.queues[nameQueue]; !ok {
		b.queues[nameQueue] = list.New()
		b.conds[nameQueue] = sync.NewCond(&b.mu)
	}
	return b.queues[nameQueue], b.conds[nameQueue]
}

func (b *BrokerQueue) put(w http.ResponseWriter, r *http.Request) {
	queueName := strings.TrimPrefix(r.URL.Path, "/")
	value := r.URL.Query().Get("v")
	if value == "" {
		http.Error(w, "No value provided", http.StatusBadRequest)
		return
	}
	//  необходимо получить очередь
	queue, cond := b.getQueue(queueName)
	b.mu.Lock()
	queue.PushBack(value)
	//  уведомить ожид кл
	cond.Signal()
	b.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func (b *BrokerQueue) get(w http.ResponseWriter, r *http.Request) {
	queueName := strings.TrimPrefix(r.URL.Path, "/")
	queue, cond := b.getQueue(queueName)

	// Получаем таймаут из параметров
	timeoutStr := r.URL.Query().Get("timeout")
	var timeout time.Duration

	if timeoutStr != "" {
		timeoutSeconds, err := strconv.Atoi(timeoutStr)
		if err != nil {
			http.Error(w, "Invalid timeout parameter", http.StatusBadRequest)
			return
		}
		timeout = time.Duration(timeoutSeconds) * time.Second
	}

	b.mu.Lock()

	// для извлечения сообщения
	extractMessage := func() (string, bool) {
		if queue.Len() > 0 {
			element := queue.Front()
			msg := element.Value.(string)
			queue.Remove(element)
			return msg, true
		}
		return "", false
	}

	if msg, found := extractMessage(); found {
		b.mu.Unlock()
		w.Write([]byte(msg))
		return
	}

	if timeout == 0 {
		b.mu.Unlock()
		http.Error(w, "No messages available", http.StatusNotFound)
		return
	}

	// Канал для передачи сообщения
	messageChan := make(chan string, 1)
	done := make(chan struct{})

	// Горутина для ожидания сообщения
	go func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		// Ждем появления сообщения или истечения таймаута
		for {
			if msg, found := extractMessage(); found {
				messageChan <- msg
				close(done)
				return
			}

			cond.Wait() // Ожидаем уведомления о новом сообщении
		}
	}()

	b.mu.Unlock()

	// Ожидание сообщения или истечения таймаута
	select {
	case msg := <-messageChan:
		w.Write([]byte(msg))
	case <-time.After(timeout):
		// Если таймаут истек, возвращаем 404
		http.Error(w, "No messages available", http.StatusNotFound)
		// Закрываем горутину
		close(done)
	}

	// Ждем завершения горутины
	<-done
}

func main() {
	// Получаем порт из аргументов командной строки
	port := flag.Int("port", 8080, "Port for the queue service")
	flag.Parse()

	// Инициализируем сервис очередей
	queueService := NewQueueService()

	// Определяем обработку POST и GET запросов
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			queueService.put(w, r) // Обрабатываем POST-запрос
		} else if r.Method == http.MethodGet {
			queueService.get(w, r) // Обрабатываем GET-запрос
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	// Запуск HTTP-сервера
	log.Printf("Starting server on port %d...", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
