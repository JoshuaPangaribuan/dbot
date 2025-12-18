package service

import "context"

type Service interface {
	Pong(ctx context.Context) string
}

type service struct{}

func New() Service {
	return service{}
}

func (service) Pong(context.Context) string {
	return "pong"
}
