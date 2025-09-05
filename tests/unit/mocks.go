package unit

import (
	"context"
	"shop-microservice/internal/domain/model"

	"github.com/stretchr/testify/mock"
)

type mockOrderRepository struct {
	mock.Mock
}

func (m *mockOrderRepository) Save(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockOrderRepository) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *mockOrderRepository) FindAll(ctx context.Context) ([]*model.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

type repoForPreload struct {
	mock.Mock
}

func (r *repoForPreload) Save(ctx context.Context, order *model.Order) error {
	args := r.Called(ctx, order)
	return args.Error(0)
}

func (r *repoForPreload) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	args := r.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (r *repoForPreload) FindAll(ctx context.Context) ([]*model.Order, error) {
	args := r.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}
