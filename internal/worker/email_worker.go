package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

type EmailJob struct {
	To      string
	Subject string
	Body    string
}

type EmailSender interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

type EmailWorker struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	started bool
	closed  bool
	jobs    chan EmailJob
	Sender  EmailSender
	Logger  *slog.Logger
}

func NewEmailWorker(sender EmailSender, logger *slog.Logger) *EmailWorker {
	return &EmailWorker{
		jobs:   make(chan EmailJob, 100),
		Sender: sender,
		Logger: logger,
	}
}

func (w *EmailWorker) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return errors.New("worker is closed")
	}

	if w.started {
		return nil
	}

	w.started = true

	w.wg.Go(func() {
		for job := range w.jobs {
			w.process(job)
		}
	})

	w.Logger.Info("worker started working")
	return nil
}

func (w *EmailWorker) process(job EmailJob) {

	maxAttempt := 3

	for attempt := 1; attempt <= maxAttempt; attempt++ {

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		err := w.Sender.Send(ctx, job.To, job.Subject, job.Body)
		cancel()

		if err == nil {
			w.Logger.Info("email sent successfully", "user", job.To)
			return
		}

		w.Logger.Error("email send failed", "user", job.To, "attempt", attempt, "err", err)

		if attempt < maxAttempt {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	w.Logger.Error("email send failed after maxAttempt(3)", "user", job.To)

}

func (w *EmailWorker) Close() {
	w.mu.Lock()

	if w.closed {
		w.mu.Unlock()
		return
	}

	w.closed = true
	close(w.jobs)

	w.mu.Unlock()
	w.wg.Wait()
}

func (w *EmailWorker) Enqueue(job EmailJob) error {

	if strings.TrimSpace(job.To) == "" || strings.TrimSpace(job.Subject) == "" || strings.TrimSpace(job.Body) == "" {
		return errors.New("invalid job provided")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return errors.New("worker is closed")
	}

	if !w.started {
		return errors.New("worker is not started")
	}

	select {
	case w.jobs <- job:
		return nil
	default:
		return errors.New("queue is full")
	}
}

func (w *EmailWorker) NotifyLogin(email, name string) error {
	job := EmailJob{
		To:      email,
		Subject: "New login to ShortUrl account",
		Body:    fmt.Sprintf("Hello %s, \n\nYour account was successfully logged in.", name),
	}

	return w.Enqueue(job)
}
