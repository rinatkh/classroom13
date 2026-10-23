package note

import "context"

type Repository interface {
	Create(ctx context.Context, value Note) (Note, error)
	ListByOwner(ctx context.Context, ownerID int64) ([]Note, error)
	Get(ctx context.Context, id int64) (Note, error)
	Update(ctx context.Context, value Note) (Note, error)
	Delete(ctx context.Context, id int64) error
}
