package polls_postgres

import (
	core_postgres_pool "github.com/oleg-morshel/murmur-api/internal/core/repository/postgres/pool"
)

type PollRepository struct {
	pool core_postgres_pool.Pool
}

func NewPollRepository(pool core_postgres_pool.Pool) *PollRepository {
	return &PollRepository{pool: pool}
}
