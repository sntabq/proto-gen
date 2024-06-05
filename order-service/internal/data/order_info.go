package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"order-service/internal/data/dto"
	"order-service/internal/data/models"
	"strings"
)

type OrderStorage struct {
	DB *sql.DB
}

var (
	ErrRecordNotFound = errors.New("record (row, entry) not found")
)

func (os *OrderStorage) SaveOrder(ctx context.Context, orderDTO *dto.OrderDTO) (*dto.OrderDTO, error) {
	const op = "data.SaveOrder"
	fail := func(e error) error {
		return fmt.Errorf("%s: %v", op, e)
	}
	insertItemQuery := `INSERT INTO order_service.orders (user_id, item_id)
            VALUES ($1, $2)
            RETURNING id`
	args := []interface{}{
		orderDTO.UserId,
		orderDTO.ItemId,
	}

	tx, err := os.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fail(err)
	}
	defer tx.Rollback()

	err = os.DB.QueryRowContext(ctx, insertItemQuery, args...).Scan(&orderDTO.ID)
	if err != nil || orderDTO.ID == 0 {
		var pqErr *pq.Error
		switch {
		case errors.As(err, &pqErr) && pqErr.Code == "23505" && strings.Contains(pqErr.Message, "item_info_name_key"):
			return nil, fail(err)
		case errors.As(err, &pqErr) && pqErr.Code == "23503" && strings.Contains(pqErr.Detail, "is not present"):
			return nil, fail(err)
		case errors.Is(err, sql.ErrNoRows):
			return nil, fail(err)
		default:
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fail(err)
	}

	getOrderItemQuery := `
				SELECT * FROM catalogue.item_info
				WHERE id = $1`
	err = os.DB.QueryRowContext(ctx, getOrderItemQuery, orderDTO.ItemId).Scan(
		&orderDTO.Item.ID,
		&orderDTO.Item.Name,
		&orderDTO.Item.Price,
		&orderDTO.Item.Description,
		&orderDTO.Item.Quantity,
		&orderDTO.Item.ImageURL,
	)

	return orderDTO, nil
}

func (os *OrderStorage) GetOrderById(ctx context.Context, id int) (*models.Order, error) {
	const op = "data.GetOrderById"
	fail := func(e error) error {
		return fmt.Errorf("%s: %v", op, e)
	}
	var order models.Order
	query := `SELECT * FROM order_service.orders
			WHERE id = $1`
	err := os.DB.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserId,
		&order.ItemId,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, fail(ErrRecordNotFound)
		default:
			return nil, fail(err)
		}
	}
	return &order, nil
}

func (os *OrderStorage) GetAllOrders(ctx context.Context) ([]*dto.OrderDTO, error) {
	const op = "data.GetAllOrders"
	fail := func(e error) error {
		return fmt.Errorf("%s, %v", op, e)
	}
	query := `
				SELECT *
				FROM order_service.orders
				LEFT JOIN catalogue.item_info iu
				ON orders.item_id = iu.id;`
	rows, err := os.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fail(err)
	}

	var orderDTOS []*dto.OrderDTO
	for rows.Next() {
		var orderDTO dto.OrderDTO
		err := rows.Scan(
			&orderDTO.ID,
			&orderDTO.UserId,
			&orderDTO.ItemId,
			&orderDTO.Item.ID,
			&orderDTO.Item.Name,
			&orderDTO.Item.Price,
			&orderDTO.Item.Description,
			&orderDTO.Item.Quantity,
			&orderDTO.Item.ImageURL,
		)
		if err != nil {
			return nil, fail(err)
		}

		orderDTOS = append(orderDTOS, &orderDTO)
	}

	return orderDTOS, nil
}

func (os *OrderStorage) GetOrdersByUserId(ctx context.Context, userId int) ([]*dto.OrderDTO, error) {
	const op = "data.GetOrdersByUserId"
	fail := func(e error) error {
		return fmt.Errorf("%s: %v", op, e)
	}

	query := `
			SELECT o.*,
			       i.*
			FROM order_service.orders o
			INNER JOIN catalogue.item_info i
			ON o.item_id = i.id
			WHERE user_id = $1
`
	stmt, err := os.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, fail(err)
	}

	rows, err := stmt.QueryContext(ctx, userId)
	if err != nil {
		return nil, fail(err)
	}

	var orders []*dto.OrderDTO
	for rows.Next() {
		var order dto.OrderDTO
		err := rows.Scan(
			&order.ID,
			&order.UserId,
			&order.ItemId,
			&order.Item.ID,
			&order.Item.Name,
			&order.Item.Price,
			&order.Item.Description,
			&order.Item.Quantity,
			&order.Item.ImageURL,
		)
		if err != nil {
			return nil, fail(err)
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

func (os *OrderStorage) DeleteOrderById(ctx context.Context, id int) error {
	const op = "data.DeleteOrderById"
	fail := func(e error) error {
		return fmt.Errorf("%s: %v", op, e)
	}
	query := `
			DELETE FROM order_service.orders
			WHERE id = $1`
	tx, err := os.DB.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return fail(err)
	}

	exec, err := os.DB.Exec(query, id)
	if err != nil {
		return err
	}

	affected, err := exec.RowsAffected()
	if err != nil || affected == 0 {
		return fail(err)
	}

	if err = tx.Commit(); err != nil {
		return fail(err)
	}

	return nil
}
