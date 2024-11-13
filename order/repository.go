package order

import (
	"context"
	"database/sql"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Repository interface {
	Close()
	PutOrder(ctx context.Context, o Order) error
	GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func (p PostgresRepository) Close() {
	p.db.Close()
}

func (p PostgresRepository) PutOrder(ctx context.Context, o Order) error {
	t, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			t.Rollback()
			return
		}
		err = t.Commit()
	}()

	t.ExecContext(ctx, "INSERT INTO orders(id,created_at,account_id,total_price) VALUES ($1,$2,$3,$4)", o.ID, o.CreatedAt, o.AccountID, o.TotalPrice)
	if err != nil {
		return err
	}
	smt, _ := t.PrepareContext(ctx, pq.CopyIn("order_products", "order_id", "product_id", "quantity"))
	for _, p := range o.Products {
		_, err = smt.ExecContext(ctx, o.ID, p.ID, p.Quantity)
		if err != nil {
			return err
		}
	}

	_, err = smt.ExecContext(ctx)
	if err != nil {
		return err
	}
	smt.Close()
	return nil
}

func (p PostgresRepository) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT
      o.id,
      o.created_at,
      o.account_id,
      o.total_price::money::numeric::float8,
      op.product_id,
      op.quantity
    FROM orders o JOIN order_products op ON (o.id = op.order_id)
    WHERE o.account_id = $1
    ORDER BY o.id`,
		accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []Order{}
	order := &Order{}
	lastOrder := &Order{}
	orderedProduct := &OrderedProduct{}
	products := []OrderedProduct{}

	// Scan rows into Order structs
	for rows.Next() {
		if err = rows.Scan(
			&order.ID,
			&order.CreatedAt,
			&order.AccountID,
			&order.TotalPrice,
			&orderedProduct.ID,
			&orderedProduct.Quantity,
		); err != nil {
			return nil, err
		}
		// Scan order
		if lastOrder.ID != "" && lastOrder.ID != order.ID {
			newOrder := Order{
				ID:         lastOrder.ID,
				AccountID:  lastOrder.AccountID,
				CreatedAt:  lastOrder.CreatedAt,
				TotalPrice: lastOrder.TotalPrice,
				Products:   products,
			}
			orders = append(orders, newOrder)
			products = []OrderedProduct{}
		}
		// Scan products
		products = append(products, OrderedProduct{
			ID:       orderedProduct.ID,
			Quantity: orderedProduct.Quantity,
		})

		*lastOrder = *order
	}

	// Add last order (or first :D)
	if lastOrder != nil {
		newOrder := Order{
			ID:         lastOrder.ID,
			AccountID:  lastOrder.AccountID,
			CreatedAt:  lastOrder.CreatedAt,
			TotalPrice: lastOrder.TotalPrice,
			Products:   products,
		}
		orders = append(orders, newOrder)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{
		db,
	}, nil
}
