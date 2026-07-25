package accrual

import "go-musthave-diploma-tpl/internal/model"

type Worker struct {
	repository OrderRepository
	workers    int
	inputCh    <-chan model.Order
}

func NewWorker(repository OrderRepository) *Worker {
	return &Worker{repository: repository}
}

func (w *Worker) Run() {

}
