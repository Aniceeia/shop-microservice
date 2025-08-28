package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"shop-microservice/internal/domain/model"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func errFail(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// Save - saves order, delivery, payment, items
func (r *OrderRepository) Save(ctx context.Context, order *model.Order) error {
	fail := func(err error) error {
		return fmt.Errorf("Save Order: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	if err := r.saveOrder(ctx, tx, order); err != nil {
		return fail(err)
	}

	if err := r.saveDelivery(ctx, tx, order); err != nil {
		return fail(err)
	}

	if err := r.savePayment(ctx, tx, order); err != nil {
		return fail(err)
	}

	if err := r.saveItems(ctx, tx, order); err != nil {
		return fail(err)
	}

	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	return nil
}

func (r *OrderRepository) saveOrder(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	query := `
        INSERT INTO orders (
            order_uid, track_number, entry, locale, internal_signature, 
            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid) DO UPDATE SET
            track_number = EXCLUDED.track_number,
            entry = EXCLUDED.entry,
            locale = EXCLUDED.locale,
            internal_signature = EXCLUDED.internal_signature,
            customer_id = EXCLUDED.customer_id,
            delivery_service = EXCLUDED.delivery_service,
            shardkey = EXCLUDED.shardkey,
            sm_id = EXCLUDED.sm_id,
            date_created = EXCLUDED.date_created,
            oof_shard = EXCLUDED.oof_shard
	`
	_, err := tx.ExecContext(ctx, query,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	return err
}

func (r *OrderRepository) saveDelivery(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	query := `
        INSERT INTO deliveries (
            order_uid, name, phone, zip, city, address, region, email
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (order_uid) DO UPDATE SET
            name = EXCLUDED.name,
            phone = EXCLUDED.phone,
            zip = EXCLUDED.zip,
            city = EXCLUDED.city,
            address = EXCLUDED.address,
            region = EXCLUDED.region,
            email = EXCLUDED.email
    `

	_, err := tx.ExecContext(ctx, query,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	return err
}

func (r *OrderRepository) savePayment(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	query := `
        INSERT INTO payments (
            order_uid, transaction, request_id, currency, provider, 
            amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid) DO UPDATE SET
            transaction = EXCLUDED.transaction,
            request_id = EXCLUDED.request_id,
            currency = EXCLUDED.currency,
            provider = EXCLUDED.provider,
            amount = EXCLUDED.amount,
            payment_dt = EXCLUDED.payment_dt,
            bank = EXCLUDED.bank,
            delivery_cost = EXCLUDED.delivery_cost,
            goods_total = EXCLUDED.goods_total,
            custom_fee = EXCLUDED.custom_fee
    `

	_, err := tx.ExecContext(ctx, query,
		order.OrderUID,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	return err
}

func (r *OrderRepository) saveItems(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM items WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		return err
	}

	if len(order.Items) == 0 {
		return nil
	}

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO items (
            order_uid, chrt_id, track_number, price, rid, name, 
            sale, size, total_price, nm_id, brand, status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range order.Items {
		_, err := stmt.ExecContext(ctx,
			order.OrderUID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// FindByID - finds orders by uid
func (r *OrderRepository) FindByID(ctx context.Context, uid string) (*model.Order, error) {
	order, err := r.findOrderByID(ctx, uid)
	if err != nil {
		return nil, errFail("Find By ID Order: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) findOrderByID(ctx context.Context, uid string) (*model.Order, error) {
	order, err := r.queryOrderWithDeliveryAndPayment(ctx, uid)
	if err != nil {
		return nil, err
	}

	items, err := r.queryOrderItems(ctx, uid)
	if err != nil {
		return nil, err
	}

	order.Items = items
	return order, nil
}

func (r *OrderRepository) queryOrderWithDeliveryAndPayment(ctx context.Context, uid string) (*model.Order, error) {
	query := `
        SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
               o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
               d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
               p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
               p.bank, p.delivery_cost, p.goods_total, p.custom_fee
        FROM orders o
        LEFT JOIN deliveries d ON o.order_uid = d.order_uid
        LEFT JOIN payments p ON o.order_uid = p.order_uid
        WHERE o.order_uid = $1
    `

	order := &model.Order{}
	var delivery model.Delivery
	var payment model.Payment

	err := r.db.QueryRowContext(ctx, query, uid).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
		&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
		&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDt,
		&payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errFail("order not found: %w", err)
		}
		return nil, errFail("failed to query order: %w", err)
	}

	order.Delivery = delivery
	order.Payment = payment

	return order, nil
}

func (r *OrderRepository) queryOrderItems(ctx context.Context, uid string) ([]model.Item, error) {
	query := `
        SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
        FROM items WHERE order_uid = $1
    `

	rows, err := r.db.QueryContext(ctx, query, uid)
	if err != nil {
		return nil, errFail("failed to query items: %w", err)
	}
	defer rows.Close()

	items, err := r.scanItems(rows)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *OrderRepository) scanItems(rows *sql.Rows) ([]model.Item, error) {
	var items []model.Item

	for rows.Next() {
		item, err := r.scanSingleItem(rows)
		if err != nil {
			return nil, errFail("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, errFail("rows iteration error: %w", err)
	}

	return items, nil
}

func (r *OrderRepository) scanSingleItem(rows *sql.Rows) (model.Item, error) {
	var item model.Item
	err := rows.Scan(
		&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
		&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
	)
	if err != nil {
		return model.Item{}, err
	}
	return item, nil
}

// FindAll - finds all orders
func (r *OrderRepository) FindAll(ctx context.Context) ([]*model.Order, error) {
	orders, err := r.findAllOrders(ctx)
	if err != nil {
		return nil, errFail("FindAll: %w", err)
	}
	// Конвертируем []model.Order в []*model.Order
	result := make([]*model.Order, len(orders))
	for i := range orders {
		result[i] = &orders[i]
	}
	return result, nil
}

func (r *OrderRepository) findAllOrders(ctx context.Context) ([]model.Order, error) {
	orders, err := r.getOrdersWithDeliveryAndPayment(ctx)
	if err != nil {
		return nil, err
	}

	orderUIDs := extractOrderUIDs(orders)

	itemsByOrder, err := r.getItemsForOrders(ctx, orderUIDs)
	if err != nil {
		return nil, err
	}

	for i := range orders {
		orders[i].Items = itemsByOrder[orders[i].OrderUID]
	}

	return orders, nil
}

func extractOrderUIDs(orders []model.Order) []string {
	uids := make([]string, len(orders))
	for i, order := range orders {
		uids[i] = order.OrderUID
	}
	return uids
}

func (r *OrderRepository) getItemsForOrders(ctx context.Context, orderUIDs []string) (map[string][]model.Item, error) {
	if len(orderUIDs) == 0 {
		return make(map[string][]model.Item), nil
	}

	placeholders := make([]string, len(orderUIDs))
	args := make([]interface{}, len(orderUIDs))
	for i, uid := range orderUIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = uid
	}
	// Убедимся что запрос использует индекс
	// EXPLAIN ANALYZE SELECT * FROM items WHERE order_uid IN (...);
	// first query
	query := fmt.Sprintf(`
        SELECT order_uid, chrt_id, track_number, price, rid, name, 
               sale, size, total_price, nm_id, brand, status
        FROM items 
        WHERE order_uid IN (%s)
        ORDER BY order_uid
    `, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errFail("failed to query items for orders: %w", err)
	}
	defer rows.Close()

	itemsByOrder := make(map[string][]model.Item)
	for rows.Next() {
		var orderUID string
		item := model.Item{}

		err := rows.Scan(
			&orderUID,
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, errFail("failed to scan item: %w", err)
		}

		itemsByOrder[orderUID] = append(itemsByOrder[orderUID], item)
	}

	if err = rows.Err(); err != nil {
		return nil, errFail("rows iteration error: %w", err)
	}

	return itemsByOrder, nil
}

func (r *OrderRepository) getOrdersWithDeliveryAndPayment(ctx context.Context) ([]model.Order, error) {
	//second query
	rows, err := r.queryAllOrdersWithDeliveryAndPayment(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		order, err := r.scanOrderWithDeliveryAndPayment(rows)
		if err != nil {
			return nil, errFail("failed to scan order: %w", err)
		}
		orders = append(orders, *order)
	}

	if err = rows.Err(); err != nil {
		return nil, errFail("rows iteration error: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) queryAllOrdersWithDeliveryAndPayment(ctx context.Context) (*sql.Rows, error) {
	query := `
        SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
               o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
               d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
               p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
               p.bank, p.delivery_cost, p.goods_total, p.custom_fee
        FROM orders o
        LEFT JOIN deliveries d ON o.order_uid = d.order_uid
        LEFT JOIN payments p ON o.order_uid = p.order_uid
        ORDER BY o.date_created DESC
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errFail("failed to query orders: %w", err)
	}
	return rows, nil
}

func (r *OrderRepository) scanOrderWithDeliveryAndPayment(rows *sql.Rows) (*model.Order, error) {
	var order model.Order
	var delivery model.Delivery
	var payment model.Payment

	err := rows.Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
		&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
		&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDt,
		&payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)

	if err != nil {
		return nil, err
	}

	order.Delivery = delivery
	order.Payment = payment

	return &order, nil
}
