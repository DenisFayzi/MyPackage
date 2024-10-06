package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBrokerQueue_put(t *testing.T) {
	// Создаем новый экземпляр BrokerQueue
	bq := NewQueueService()

	// Определяем имя очереди и сообщение
	queueName := "testQueue"
	message := "testMessage"

	// Создаем HTTP-запрос к обработчику put
	req, err := http.NewRequest("POST", "/"+queueName+"?v="+message, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Вызываем метод put
	bq.put(rr, req)

	// Проверяем статус ответа
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Проверяем, что сообщение добавлено в очередь
	bq.mu.Lock()
	queue, exists := bq.queues[queueName]
	if !exists {
		t.Errorf("Queue %s does not exist", queueName)
	} else {
		if queue.Len() != 1 {
			t.Errorf("Queue %s should have 1 message, has %d", queueName, queue.Len())
		} else {
			elem := queue.Front()
			if elem.Value != message {
				t.Errorf("Expected message %s in queue %s, got %v", message, queueName, elem.Value)
			}
		}
	}
	bq.mu.Unlock()
}
func TestBrokerQueue_GetFIFOOrder(t *testing.T) {
	bq := NewQueueService()
	queueName := "testQueue"
	message1 := "testMessage1"
	message2 := "testMessage2"

	// Каналы для получения сообщений от клиентов
	client1Chan := make(chan string)
	client2Chan := make(chan string)

	// Запускаем первого клиента
	go func() {
		req, err := http.NewRequest("GET", "/"+queueName+"?timeout=5", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		bq.get(rr, req)
		if rr.Code == http.StatusOK {
			client1Chan <- rr.Body.String()
		} else {
			client1Chan <- fmt.Sprintf("Error: %d", rr.Code)
		}
	}()

	// Даём время первому клиенту начать ожидание
	time.Sleep(100 * time.Millisecond)

	// Запускаем второго клиента
	go func() {
		req, err := http.NewRequest("GET", "/"+queueName+"?timeout=5", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		bq.get(rr, req)
		if rr.Code == http.StatusOK {
			client2Chan <- rr.Body.String()
		} else {
			client2Chan <- fmt.Sprintf("Error: %d", rr.Code)
		}
	}()

	// Даём время обоим клиентам начать ожидание
	time.Sleep(100 * time.Millisecond)

	// Добавляем первое сообщение в очередь
	req, err := http.NewRequest("POST", "/"+queueName+"?v="+message1, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	bq.put(rr, req)

	// Ожидаем, что первый клиент получит сообщение
	select {
	case msg1 := <-client1Chan:
		if msg1 != message1 {
			t.Errorf("First client received incorrect message: got %v, want %v", msg1, message1)
		} else {
			t.Logf("First client received message: %v", msg1)
		}
	case <-time.After(1 * time.Second):
		t.Error("First client did not receive message in time")
	}

	// Второй клиент всё ещё ждёт. Добавляем второе сообщение
	time.Sleep(100 * time.Millisecond)

	req, err = http.NewRequest("POST", "/"+queueName+"?v="+message2, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	bq.put(rr, req)

	// Ожидаем, что второй клиент получит сообщение
	select {
	case msg2 := <-client2Chan:
		if msg2 != message2 {
			t.Errorf("Second client received incorrect message: got %v, want %v", msg2, message2)
		} else {
			t.Logf("Second client received message: %v", msg2)
		}
	case <-time.After(1 * time.Second):
		t.Error("Second client did not receive message in time")
	}
}
func TestBrokerQueue_FIFOOrderWithoutTimeout(t *testing.T) {
	bq := NewQueueService()
	queueName := "testQueue"
	messages := []string{"message1", "message2", "message3"}

	// Добавляем сообщения в очередь с помощью PUT-запросов
	for _, msg := range messages {
		req, err := http.NewRequest("POST", "/"+queueName+"?v="+msg, nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		bq.put(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("PUT request failed with status %d", status)
		}
	}

	// Извлекаем сообщения из очереди с помощью GET-запросов без таймера
	for i, expectedMsg := range messages {
		req, err := http.NewRequest("GET", "/"+queueName, nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		bq.get(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("GET request failed with status %d", status)
			continue
		}

		receivedMsg := rr.Body.String()
		if receivedMsg != expectedMsg {
			t.Errorf("Message %d incorrect: got %v, want %v", i+1, receivedMsg, expectedMsg)
		} else {
			t.Logf("Message %d received correctly: %v", i+1, receivedMsg)
		}
	}
}
