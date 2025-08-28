package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"shop-microservice/internal/domain/model"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderRepository_Save_Unit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOrderRepository(db)

	order := createTestOrder()
	ctx := context.Background()

	mock.ExpectBegin()

	mock.ExpectExec("INSERT INTO orders").
		WithArgs(
			order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
			order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO deliveries").
		WithArgs(
			order.OrderUID,
			order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
			order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO payments").
		WithArgs(
			order.OrderUID,
			order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
			order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
			order.Payment.GoodsTotal, order.Payment.CustomFee,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("DELETE FROM items WHERE order_uid = ?").WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectPrepare("INSERT INTO items")
	mock.ExpectExec("INSERT INTO items").
		WithArgs(
			order.OrderUID,
			order.Items[0].ChrtID, order.Items[0].TrackNumber, order.Items[0].Price, order.Items[0].Rid,
			order.Items[0].Name, order.Items[0].Sale, order.Items[0].Size, order.Items[0].TotalPrice,
			order.Items[0].NmID, order.Items[0].Brand, order.Items[0].Status,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.Save(ctx, order)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_Save_TransactionRollbackOnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOrderRepository(db)

	order := createTestOrder()
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").
		WithArgs(
			order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
			order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard,
		).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err = repo.Save(ctx, order)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_FindByID_Unit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOrderRepository(db)

	orderUID := "test-order-uid"
	expectedOrder := createTestOrder()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"order_uid", "track_number", "entry", "locale", "internal_signature",
		"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard",
		"name", "phone", "zip", "city", "address", "region", "email",
		"transaction", "request_id", "currency", "provider", "amount", "payment_dt",
		"bank", "delivery_cost", "goods_total", "custom_fee",
	}).AddRow(
		expectedOrder.OrderUID, expectedOrder.TrackNumber, expectedOrder.Entry, expectedOrder.Locale, expectedOrder.InternalSignature,
		expectedOrder.CustomerID, expectedOrder.DeliveryService, expectedOrder.Shardkey, expectedOrder.SmID, expectedOrder.DateCreated, expectedOrder.OofShard,
		expectedOrder.Delivery.Name, expectedOrder.Delivery.Phone, expectedOrder.Delivery.Zip, expectedOrder.Delivery.City, expectedOrder.Delivery.Address, expectedOrder.Delivery.Region, expectedOrder.Delivery.Email,
		expectedOrder.Payment.Transaction, expectedOrder.Payment.RequestID, expectedOrder.Payment.Currency, expectedOrder.Payment.Provider, expectedOrder.Payment.Amount, expectedOrder.Payment.PaymentDt,
		expectedOrder.Payment.Bank, expectedOrder.Payment.DeliveryCost, expectedOrder.Payment.GoodsTotal, expectedOrder.Payment.CustomFee,
	)

	mock.ExpectQuery("SELECT o.order_uid").
		WithArgs(orderUID).
		WillReturnRows(rows)

	itemRows := sqlmock.NewRows([]string{
		"chrt_id", "track_number", "price", "rid", "name", "sale", "size", "total_price", "nm_id", "brand", "status",
	}).AddRow(
		expectedOrder.Items[0].ChrtID, expectedOrder.Items[0].TrackNumber, expectedOrder.Items[0].Price, expectedOrder.Items[0].Rid, expectedOrder.Items[0].Name,
		expectedOrder.Items[0].Sale, expectedOrder.Items[0].Size, expectedOrder.Items[0].TotalPrice, expectedOrder.Items[0].NmID, expectedOrder.Items[0].Brand, expectedOrder.Items[0].Status,
	)

	mock.ExpectQuery("SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = ?").
		WithArgs(orderUID).
		WillReturnRows(itemRows)

	result, err := repo.FindByID(ctx, orderUID)
	require.NoError(t, err)

	assert.Equal(t, expectedOrder.OrderUID, result.OrderUID)
	assert.Equal(t, expectedOrder.TrackNumber, result.TrackNumber)
	assert.Equal(t, expectedOrder.Delivery.Name, result.Delivery.Name)
	assert.Equal(t, expectedOrder.Payment.Amount, result.Payment.Amount)
	require.Len(t, result.Items, 1)
	assert.Equal(t, expectedOrder.Items[0].Name, result.Items[0].Name)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_FindByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOrderRepository(db)

	orderUID := "non-existent-id"
	ctx := context.Background()

	mock.ExpectQuery("SELECT o.order_uid").
		WithArgs(orderUID).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.FindByID(ctx, orderUID)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "order not found")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_FindAll_Unit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOrderRepository(db)

	expectedOrder := createTestOrder()
	ctx := context.Background()

	orderRows := sqlmock.NewRows([]string{
		"order_uid", "track_number", "entry", "locale", "internal_signature",
		"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard",
		"name", "phone", "zip", "city", "address", "region", "email",
		"transaction", "request_id", "currency", "provider", "amount", "payment_dt",
		"bank", "delivery_cost", "goods_total", "custom_fee",
	}).AddRow(
		expectedOrder.OrderUID, expectedOrder.TrackNumber, expectedOrder.Entry, expectedOrder.Locale, expectedOrder.InternalSignature,
		expectedOrder.CustomerID, expectedOrder.DeliveryService, expectedOrder.Shardkey, expectedOrder.SmID, expectedOrder.DateCreated, expectedOrder.OofShard,
		expectedOrder.Delivery.Name, expectedOrder.Delivery.Phone, expectedOrder.Delivery.Zip, expectedOrder.Delivery.City, expectedOrder.Delivery.Address, expectedOrder.Delivery.Region, expectedOrder.Delivery.Email,
		expectedOrder.Payment.Transaction, expectedOrder.Payment.RequestID, expectedOrder.Payment.Currency, expectedOrder.Payment.Provider, expectedOrder.Payment.Amount, expectedOrder.Payment.PaymentDt,
		expectedOrder.Payment.Bank, expectedOrder.Payment.DeliveryCost, expectedOrder.Payment.GoodsTotal, expectedOrder.Payment.CustomFee,
	)

	mock.ExpectQuery("SELECT o.order_uid").
		WillReturnRows(orderRows)

	itemRows := sqlmock.NewRows([]string{
		"order_uid", "chrt_id", "track_number", "price", "rid", "name", "sale", "size", "total_price", "nm_id", "brand", "status",
	}).AddRow(
		expectedOrder.OrderUID,
		expectedOrder.Items[0].ChrtID, expectedOrder.Items[0].TrackNumber, expectedOrder.Items[0].Price, expectedOrder.Items[0].Rid, expectedOrder.Items[0].Name,
		expectedOrder.Items[0].Sale, expectedOrder.Items[0].Size, expectedOrder.Items[0].TotalPrice, expectedOrder.Items[0].NmID, expectedOrder.Items[0].Brand, expectedOrder.Items[0].Status,
	)

	mock.ExpectQuery("SELECT order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid IN").
		WithArgs(expectedOrder.OrderUID).
		WillReturnRows(itemRows)

	result, err := repo.FindAll(ctx)
	require.NoError(t, err)

	require.Len(t, result, 1)
	assert.Equal(t, expectedOrder.OrderUID, result[0].OrderUID)
	assert.Equal(t, expectedOrder.Delivery.Name, result[0].Delivery.Name)
	require.Len(t, result[0].Items, 1)
	assert.Equal(t, expectedOrder.Items[0].Name, result[0].Items[0].Name)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_FindAll_NoOrders(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOrderRepository(db)
	ctx := context.Background()

	orderRows := sqlmock.NewRows([]string{
		"order_uid", "track_number", "entry", "locale", "internal_signature",
		"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard",
		"name", "phone", "zip", "city", "address", "region", "email",
		"transaction", "request_id", "currency", "provider", "amount", "payment_dt",
		"bank", "delivery_cost", "goods_total", "custom_fee",
	})

	mock.ExpectQuery("SELECT o.order_uid").
		WillReturnRows(orderRows)

	result, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExtractOrderUIDs(t *testing.T) {
	orders := []model.Order{
		{OrderUID: "order1"},
		{OrderUID: "order2"},
		{OrderUID: "order3"},
	}

	result := extractOrderUIDs(orders)
	expected := []string{"order1", "order2", "order3"}

	assert.Equal(t, expected, result)
}

func TestOrderRepository_FindAll_ContextTimeout(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOrderRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond) // Very short timeout
	defer cancel()
	time.Sleep(2 * time.Microsecond)

	result, err := repo.FindAll(ctx)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestExtractOrderUIDs_Empty(t *testing.T) {
	orders := []model.Order{}
	result := extractOrderUIDs(orders)
	assert.Empty(t, result)
}
