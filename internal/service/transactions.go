package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/satriaardiperdana-2020/monelog-api/internal/repository/sqlc"
)

var ErrUnavailable = errors.New("service unavailable")

type TransactionScope struct {
	Actor     Actor
	Owner     int64
	Admin     bool
	RequestID string
}

type Transaction struct {
	ID, UserID, CategoryID, ClientRequestID, CreatedBy, UpdatedBy int64
	Type, CategoryName, Title, Amount                             string
	TransactionDate, CreatedAt, UpdatedAt                         time.Time
	IsDelete                                                      bool
	Version                                                       int32
}

type TransactionInput struct {
	TransactionDate string
	Type            string
	CategoryID      int64
	Amount          string
	Title           string
	ClientRequestID int64
	Version         int32
}

type TransactionFilter struct {
	StartDate, EndDate string
	Type               string
	CategoryID         *int64
	Deleted            bool
	Limit              int32
	Cursor             string
}

type TransactionPage struct {
	Items      []Transaction
	NextCursor *string
}

type CreateTransactionResult struct {
	Transaction Transaction
	Created     bool
}

type DailySummary struct {
	Date                        time.Time
	Income, Expense, Difference string
}

type DailySummaryPage struct {
	Items      []DailySummary
	NextCursor *string
}

type ReportFilter struct {
	Range, StartDate, EndDate, GroupBy string
}

type ReportCategoryTotal struct {
	CategoryID         int64
	Name, Type, Amount string
}

type ReportPeriodTotal struct {
	PeriodStart                 time.Time
	Income, Expense, Difference string
}

type ReportSummary struct {
	StartDate, EndDate                        time.Time
	Income, Expense, Difference               string
	TopIncomeCategories, TopExpenseCategories []ReportCategoryTotal
}

type ReportBreakdown struct {
	StartDate, EndDate time.Time
	GroupBy            string
	Periods            []ReportPeriodTotal
	Categories         []ReportCategoryTotal
}

type Transactions struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	now     func() time.Time
}

func NewTransactions(pool *pgxpool.Pool) *Transactions {
	return &Transactions{pool: pool, queries: db.New(pool), now: time.Now}
}

type transactionCursor struct {
	Version                                        int `json:"v"`
	Kind, Actor, Owner, Type, Category, Start, End string
	Deleted                                        bool `json:"deleted"`
	Admin                                          bool `json:"admin"`
	Date, Timestamp, ID                            string
}

func (s *Transactions) List(ctx context.Context, scope TransactionScope, filter TransactionFilter) (TransactionPage, error) {
	if err := s.validateReadScope(ctx, scope); err != nil {
		return TransactionPage{}, err
	}
	start, end, err := validateDateRange(filter.StartDate, filter.EndDate)
	if err != nil || (filter.Type != "" && !validCategoryType(filter.Type)) {
		return TransactionPage{}, ErrValidation
	}
	limit := normalizeLimit(filter.Limit)
	if limit < 1 {
		return TransactionPage{}, ErrValidation
	}
	if filter.CategoryID != nil {
		category, err := s.queries.GetCategoryByID(ctx, db.GetCategoryByIDParams{ID: *filter.CategoryID, UserID: scope.Owner})
		if errors.Is(err, pgx.ErrNoRows) {
			return TransactionPage{}, ErrNotFound
		} else if err != nil {
			return TransactionPage{}, err
		}
		if filter.Type != "" && category.Type != filter.Type {
			return TransactionPage{}, ErrValidation
		}
	}
	cursor, err := decodeTransactionCursor(filter.Cursor)
	if err != nil || !cursorMatches(cursor, scope, filter, "transactions") {
		return TransactionPage{}, ErrValidation
	}
	paramsActive := db.ListActiveTransactionsParams{UserID: scope.Owner, StartDate: pgDate(start), EndDate: pgDate(end), FilterType: filter.Type, PageSize: limit + 1}
	paramsDeleted := db.ListDeletedTransactionsParams{UserID: scope.Owner, StartDate: pgDate(start), EndDate: pgDate(end), FilterType: filter.Type, PageSize: limit + 1}
	if filter.CategoryID != nil {
		paramsActive.HasCategory, paramsDeleted.HasCategory = true, true
		paramsActive.CategoryID, paramsDeleted.CategoryID = *filter.CategoryID, *filter.CategoryID
	}
	if cursor != nil {
		paramsActive.HasCursor, paramsDeleted.HasCursor = true, true
		paramsActive.CursorDate = pgDate(mustDate(cursor.Date))
		paramsActive.CursorCreatedAt = pgtype.Timestamptz{Time: mustTimestamp(cursor.Timestamp), Valid: true}
		paramsActive.CursorID = mustInt64(cursor.ID)
		paramsDeleted.CursorUpdatedAt = pgtype.Timestamptz{Time: mustTimestamp(cursor.Timestamp), Valid: true}
		paramsDeleted.CursorID = mustInt64(cursor.ID)
	}
	items := make([]Transaction, 0, limit)
	if filter.Deleted {
		rows, queryErr := s.queries.ListDeletedTransactions(ctx, paramsDeleted)
		if queryErr != nil {
			return TransactionPage{}, queryErr
		}
		for _, row := range rows {
			item, e := transactionFromDeletedRow(row)
			if e != nil {
				return TransactionPage{}, e
			}
			items = append(items, item)
		}
	} else {
		rows, queryErr := s.queries.ListActiveTransactions(ctx, paramsActive)
		if queryErr != nil {
			return TransactionPage{}, queryErr
		}
		for _, row := range rows {
			item, e := transactionFromActiveRow(row)
			if e != nil {
				return TransactionPage{}, e
			}
			items = append(items, item)
		}
	}
	page := paginateTransactions(items, limit, scope, filter)
	if scope.Admin {
		if err := s.auditRead(ctx, scope, "list", 0); err != nil {
			return TransactionPage{}, err
		}
	}
	return page, nil
}

func (s *Transactions) Get(ctx context.Context, scope TransactionScope, id int64, deleted bool) (Transaction, error) {
	if id <= 0 {
		return Transaction{}, ErrValidation
	}
	if err := s.validateReadScope(ctx, scope); err != nil {
		return Transaction{}, err
	}
	var item Transaction
	var err error
	if deleted {
		row, e := s.queries.GetDeletedTransaction(ctx, db.GetDeletedTransactionParams{ID: id, UserID: scope.Owner})
		err = e
		if e == nil {
			item, err = transactionFromDeletedRowForGet(row)
		}
	} else {
		row, e := s.queries.GetActiveTransaction(ctx, db.GetActiveTransactionParams{ID: id, UserID: scope.Owner})
		err = e
		if e == nil {
			item, err = transactionFromActiveRowForGet(row)
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrNotFound
	}
	if err != nil {
		return Transaction{}, err
	}
	if scope.Admin {
		if err := s.auditRead(ctx, scope, "read", id); err != nil {
			return Transaction{}, err
		}
	}
	return item, nil
}

func (s *Transactions) Create(ctx context.Context, scope TransactionScope, input TransactionInput) (CreateTransactionResult, error) {
	validated, err := s.validateInput(scope.Owner, input, false)
	if err != nil {
		return CreateTransactionResult{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CreateTransactionResult{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	owner, err := validateWriteUsers(ctx, q, scope)
	if err != nil {
		return CreateTransactionResult{}, err
	}
	if err := validateOwnerDate(owner.Timezone, validated.date, s.now()); err != nil {
		return CreateTransactionResult{}, err
	}
	if err := validateLockedCategory(ctx, q, scope.Owner, validated.categoryID, validated.kind); err != nil {
		return CreateTransactionResult{}, err
	}
	hash := canonicalTransactionHash(scope.Owner, validated)
	created, err := q.CreateTransaction(ctx, db.CreateTransactionParams{UserID: scope.Owner, CategoryID: validated.categoryID, Type: validated.kind, Amount: validated.amount.Numeric(), TransactionDate: pgDate(validated.date), Title: validated.title, ClientRequestID: input.ClientRequestID, RequestHash: hash, ActorUserID: scope.Actor.UserID})
	wasCreated := err == nil
	if errors.Is(err, pgx.ErrNoRows) {
		created, err = q.GetTransactionByRequestIDForUpdate(ctx, db.GetTransactionByRequestIDForUpdateParams{UserID: scope.Owner, ClientRequestID: input.ClientRequestID})
		if err == nil && (created.RequestHash != hash || created.IsDelete) {
			return CreateTransactionResult{}, ErrConflict
		}
	}
	if err != nil {
		return CreateTransactionResult{}, err
	}
	if scope.Admin {
		if err := insertAudit(ctx, q, scope, &created.ID, "create"); err != nil {
			return CreateTransactionResult{}, ErrUnavailable
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateTransactionResult{}, err
	}
	item, err := s.Get(ctx, scope, created.ID, false)
	if err != nil {
		return CreateTransactionResult{}, err
	}
	return CreateTransactionResult{Transaction: item, Created: wasCreated}, nil
}

func (s *Transactions) Update(ctx context.Context, scope TransactionScope, id int64, input TransactionInput) (Transaction, error) {
	validated, err := s.validateInput(scope.Owner, input, true)
	if err != nil {
		return Transaction{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Transaction{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	owner, err := validateWriteUsers(ctx, q, scope)
	if err != nil {
		return Transaction{}, err
	}
	if err := validateOwnerDate(owner.Timezone, validated.date, s.now()); err != nil {
		return Transaction{}, err
	}
	state, err := q.GetTransactionStateForUpdate(ctx, db.GetTransactionStateForUpdateParams{ID: id, UserID: scope.Owner})
	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrNotFound
	}
	if err != nil {
		return Transaction{}, err
	}
	if state.IsDelete || state.Version != input.Version {
		return Transaction{}, ErrConflict
	}
	if err := validateLockedCategory(ctx, q, scope.Owner, validated.categoryID, validated.kind); err != nil {
		return Transaction{}, err
	}
	updated, err := q.UpdateTransaction(ctx, db.UpdateTransactionParams{CategoryID: validated.categoryID, Type: validated.kind, Amount: validated.amount.Numeric(), TransactionDate: pgDate(validated.date), Title: validated.title, ActorUserID: scope.Actor.UserID, ID: id, UserID: scope.Owner, ExpectedVersion: input.Version})
	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrConflict
	}
	if err != nil {
		return Transaction{}, err
	}
	if scope.Admin {
		if err := insertAudit(ctx, q, scope, &updated.ID, "update"); err != nil {
			return Transaction{}, ErrUnavailable
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Transaction{}, err
	}
	return s.Get(ctx, scope, id, false)
}

func (s *Transactions) SoftDeleteTransaction(ctx context.Context, scope TransactionScope, id int64, version int32) error {
	if id <= 0 || version < 1 {
		return ErrValidation
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	if _, err := validateWriteUsers(ctx, q, scope); err != nil {
		return err
	}
	state, err := q.GetTransactionStateForUpdate(ctx, db.GetTransactionStateForUpdateParams{ID: id, UserID: scope.Owner})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if state.IsDelete || state.Version != version {
		return ErrConflict
	}
	updated, err := q.SoftDeleteTransaction(ctx, db.SoftDeleteTransactionParams{ActorUserID: scope.Actor.UserID, ID: id, UserID: scope.Owner, ExpectedVersion: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	if scope.Admin {
		if err := insertAudit(ctx, q, scope, &updated.ID, "delete"); err != nil {
			return ErrUnavailable
		}
	}
	return tx.Commit(ctx)
}

func (s *Transactions) RestoreTransaction(ctx context.Context, scope TransactionScope, id int64, version int32) (Transaction, error) {
	if id <= 0 || version < 1 {
		return Transaction{}, ErrValidation
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Transaction{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	if _, err := validateWriteUsers(ctx, q, scope); err != nil {
		return Transaction{}, err
	}
	state, err := q.GetTransactionStateForUpdate(ctx, db.GetTransactionStateForUpdateParams{ID: id, UserID: scope.Owner})
	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrNotFound
	}
	if err != nil {
		return Transaction{}, err
	}
	if !state.IsDelete || state.Version != version {
		return Transaction{}, ErrConflict
	}
	if err := validateLockedCategory(ctx, q, scope.Owner, state.CategoryID, state.Type); err != nil {
		return Transaction{}, err
	}
	updated, err := q.RestoreTransaction(ctx, db.RestoreTransactionParams{ActorUserID: scope.Actor.UserID, ID: id, UserID: scope.Owner, ExpectedVersion: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrConflict
	}
	if err != nil {
		return Transaction{}, err
	}
	if scope.Admin {
		if err := insertAudit(ctx, q, scope, &updated.ID, "restore"); err != nil {
			return Transaction{}, ErrUnavailable
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Transaction{}, err
	}
	return s.Get(ctx, scope, id, false)
}

func (s *Transactions) ListDailySummaries(ctx context.Context, scope TransactionScope, startValue, endValue string, limit int32, cursorValue string) (DailySummaryPage, error) {
	if err := s.validateReadScope(ctx, scope); err != nil {
		return DailySummaryPage{}, err
	}
	start, end, err := validateDateRange(startValue, endValue)
	if err != nil {
		return DailySummaryPage{}, ErrValidation
	}
	filter := TransactionFilter{StartDate: startValue, EndDate: endValue, Limit: limit, Cursor: cursorValue}
	cursor, err := decodeTransactionCursor(cursorValue)
	if err != nil || !cursorMatches(cursor, scope, filter, "daily") {
		return DailySummaryPage{}, ErrValidation
	}
	pageSize := normalizeLimit(limit)
	if pageSize < 1 {
		return DailySummaryPage{}, ErrValidation
	}
	params := db.ListDailySummariesParams{UserID: scope.Owner, StartDate: pgDate(start), EndDate: pgDate(end), PageSize: pageSize + 1}
	if cursor != nil {
		params.HasCursor = true
		params.CursorDate = pgDate(mustDate(cursor.Date))
	}
	rows, err := s.queries.ListDailySummaries(ctx, params)
	if err != nil {
		return DailySummaryPage{}, err
	}
	items := make([]DailySummary, 0, len(rows))
	for _, row := range rows {
		income, e := moneyFromNumeric(row.Income)
		if e != nil {
			return DailySummaryPage{}, e
		}
		expense, e := moneyFromNumeric(row.Expense)
		if e != nil {
			return DailySummaryPage{}, e
		}
		items = append(items, DailySummary{Date: row.TransactionDate.Time, Income: income.String(), Expense: expense.String(), Difference: formatCents(income.Cents() - expense.Cents())})
	}
	var next *string
	if len(items) > int(pageSize) {
		items = items[:pageSize]
		last := items[len(items)-1]
		encoded := encodeCursor(cursorFor(scope, filter, "daily", last.Date, time.Time{}, 0))
		next = &encoded
	}
	if scope.Admin {
		if err := s.auditRead(ctx, scope, "daily_summary", 0); err != nil {
			return DailySummaryPage{}, err
		}
	}
	return DailySummaryPage{Items: items, NextCursor: next}, nil
}

func (s *Transactions) resolveReportRange(ctx context.Context, scope TransactionScope, filter ReportFilter) (time.Time, time.Time, error) {
	if filter.Range == "custom" {
		start, end, err := validateDateRange(filter.StartDate, filter.EndDate)
		if err != nil {
			return time.Time{}, time.Time{}, ErrValidation
		}
		return start, end, nil
	}
	if filter.StartDate != "" || filter.EndDate != "" || (filter.Range != "last_7_days" && filter.Range != "last_30_days") {
		return time.Time{}, time.Time{}, ErrValidation
	}
	owner, err := s.queries.GetUserByID(ctx, scope.Owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, time.Time{}, ErrNotFound
	}
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	zone, err := ValidateTimezone(owner.Timezone)
	if err != nil {
		return time.Time{}, time.Time{}, ErrValidation
	}
	location, _ := time.LoadLocation(zone)
	now := s.now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	days := 6
	if filter.Range == "last_30_days" {
		days = 29
	}
	return today.AddDate(0, 0, -days), today, nil
}

func (s *Transactions) GetReportSummary(ctx context.Context, scope TransactionScope, filter ReportFilter) (ReportSummary, error) {
	if err := s.validateReadScope(ctx, scope); err != nil {
		return ReportSummary{}, err
	}
	start, end, err := s.resolveReportRange(ctx, scope, filter)
	if err != nil {
		return ReportSummary{}, err
	}
	args := db.GetReportSummaryParams{UserID: scope.Owner, StartDate: pgDate(start), EndDate: pgDate(end)}
	row, err := s.queries.GetReportSummary(ctx, args)
	if err != nil {
		return ReportSummary{}, err
	}
	income, err := moneyFromNumeric(row.Income)
	if err != nil {
		return ReportSummary{}, err
	}
	expense, err := moneyFromNumeric(row.Expense)
	if err != nil {
		return ReportSummary{}, err
	}
	tops, err := s.queries.ListTopReportCategories(ctx, db.ListTopReportCategoriesParams{UserID: args.UserID, StartDate: args.StartDate, EndDate: args.EndDate})
	if err != nil {
		return ReportSummary{}, err
	}
	result := ReportSummary{StartDate: start, EndDate: end, Income: income.String(), Expense: expense.String(), Difference: formatCents(income.Cents() - expense.Cents()), TopIncomeCategories: []ReportCategoryTotal{}, TopExpenseCategories: []ReportCategoryTotal{}}
	for _, item := range tops {
		amount, e := moneyFromNumeric(item.Amount)
		if e != nil {
			return ReportSummary{}, e
		}
		value := ReportCategoryTotal{CategoryID: item.CategoryID, Name: item.Name, Type: item.Type, Amount: amount.String()}
		if item.Type == CategoryTypeIncome {
			result.TopIncomeCategories = append(result.TopIncomeCategories, value)
		} else {
			result.TopExpenseCategories = append(result.TopExpenseCategories, value)
		}
	}
	if scope.Admin {
		if err := s.auditRead(ctx, scope, "report_summary", 0); err != nil {
			return ReportSummary{}, err
		}
	}
	return result, nil
}

func (s *Transactions) GetReportBreakdown(ctx context.Context, scope TransactionScope, filter ReportFilter) (ReportBreakdown, error) {
	if err := s.validateReadScope(ctx, scope); err != nil {
		return ReportBreakdown{}, err
	}
	if filter.GroupBy != "week" && filter.GroupBy != "month" && filter.GroupBy != "category" {
		return ReportBreakdown{}, ErrValidation
	}
	start, end, err := s.resolveReportRange(ctx, scope, filter)
	if err != nil {
		return ReportBreakdown{}, err
	}
	args := db.ListReportPeriodBreakdownParams{GroupBy: filter.GroupBy, UserID: scope.Owner, StartDate: pgDate(start), EndDate: pgDate(end)}
	result := ReportBreakdown{StartDate: start, EndDate: end, GroupBy: filter.GroupBy, Periods: []ReportPeriodTotal{}, Categories: []ReportCategoryTotal{}}
	if filter.GroupBy == "category" {
		rows, e := s.queries.ListReportCategoryBreakdown(ctx, db.ListReportCategoryBreakdownParams{UserID: args.UserID, StartDate: args.StartDate, EndDate: args.EndDate})
		if e != nil {
			return ReportBreakdown{}, e
		}
		for _, item := range rows {
			amount, e := moneyFromNumeric(item.Amount)
			if e != nil {
				return ReportBreakdown{}, e
			}
			result.Categories = append(result.Categories, ReportCategoryTotal{CategoryID: item.CategoryID, Name: item.Name, Type: item.Type, Amount: amount.String()})
		}
	} else {
		rows, e := s.queries.ListReportPeriodBreakdown(ctx, args)
		if e != nil {
			return ReportBreakdown{}, e
		}
		for _, item := range rows {
			income, e := moneyFromNumeric(item.Income)
			if e != nil {
				return ReportBreakdown{}, e
			}
			expense, e := moneyFromNumeric(item.Expense)
			if e != nil {
				return ReportBreakdown{}, e
			}
			result.Periods = append(result.Periods, ReportPeriodTotal{PeriodStart: item.PeriodStart.Time, Income: income.String(), Expense: expense.String(), Difference: formatCents(income.Cents() - expense.Cents())})
		}
	}
	if scope.Admin {
		if err := s.auditRead(ctx, scope, "report_breakdown", 0); err != nil {
			return ReportBreakdown{}, err
		}
	}
	return result, nil
}

type validatedTransaction struct {
	date       time.Time
	kind       string
	categoryID int64
	amount     Money
	title      string
}

func (s *Transactions) validateInput(owner int64, input TransactionInput, update bool) (validatedTransaction, error) {
	date, err := parseDate(input.TransactionDate)
	if err != nil || !validCategoryType(input.Type) || input.CategoryID <= 0 || owner <= 0 || (!update && input.ClientRequestID <= 0) || (update && input.Version < 1) {
		return validatedTransaction{}, ErrValidation
	}
	amount, err := ParseMoney(input.Amount)
	if err != nil {
		return validatedTransaction{}, ErrValidation
	}
	title, err := normalizeTitle(input.Title)
	if err != nil {
		return validatedTransaction{}, ErrValidation
	}
	return validatedTransaction{date: date, kind: input.Type, categoryID: input.CategoryID, amount: amount, title: title}, nil
}

func (s *Transactions) validateReadScope(ctx context.Context, scope TransactionScope) error {
	if scope.Actor.UserID <= 0 || scope.Owner <= 0 {
		return ErrValidation
	}
	if !scope.Admin {
		if scope.Actor.UserID != scope.Owner {
			return ErrForbidden
		}
		return nil
	}
	if scope.Actor.Role != RoleAdmin {
		return ErrForbidden
	}
	if _, err := s.queries.GetUserByID(ctx, scope.Owner); errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	} else {
		return err
	}
}

func validateWriteUsers(ctx context.Context, q *db.Queries, scope TransactionScope) (db.User, error) {
	if scope.Actor.UserID <= 0 || scope.Owner <= 0 || (!scope.Admin && scope.Actor.UserID != scope.Owner) || (scope.Admin && scope.Actor.Role != RoleAdmin) {
		return db.User{}, ErrForbidden
	}
	ids := []int64{scope.Actor.UserID}
	if scope.Owner != scope.Actor.UserID {
		ids = append(ids, scope.Owner)
	}
	users, err := q.LockUsersForUpdate(ctx, ids)
	if err != nil {
		return db.User{}, err
	}
	var owner db.User
	actorOK := false
	for _, user := range users {
		id := user.ID
		if id == scope.Actor.UserID && !user.IsDelete && (!scope.Admin || user.Role == RoleAdmin) {
			actorOK = true
		}
		if id == scope.Owner {
			owner = user
		}
	}
	if !actorOK {
		return db.User{}, ErrAuthentication
	}
	if owner.ID == 0 {
		return db.User{}, ErrNotFound
	}
	if owner.IsDelete {
		return db.User{}, ErrConflict
	}
	return owner, nil
}

func validateLockedCategory(ctx context.Context, q *db.Queries, owner, categoryID int64, kind string) error {
	category, err := q.LockScopedCategoryForTransaction(ctx, db.LockScopedCategoryForTransactionParams{CategoryID: categoryID, UserID: owner})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if category.IsDelete {
		return ErrConflict
	}
	if category.Type != kind {
		return ErrValidation
	}
	return nil
}
func validateOwnerDate(timezone string, date time.Time, now time.Time) error {
	zone, err := ValidateTimezone(timezone)
	if err != nil {
		return ErrValidation
	}
	location, _ := time.LoadLocation(zone)
	current := now.In(location)
	today := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC)
	if date.After(today) {
		return ErrValidation
	}
	return nil
}
func validateDateRange(startValue, endValue string) (time.Time, time.Time, error) {
	start, err := parseDate(startValue)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseDate(endValue)
	if err != nil || start.After(end) || end.Sub(start) > 365*24*time.Hour {
		return time.Time{}, time.Time{}, errors.New("invalid date range")
	}
	return start, end, nil
}
func parseDate(value string) (time.Time, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil || parsed.Format(time.DateOnly) != value {
		return time.Time{}, errors.New("invalid date")
	}
	return parsed, nil
}
func pgDate(value time.Time) pgtype.Date { return pgtype.Date{Time: value, Valid: true} }
func normalizeLimit(limit int32) int32 {
	if limit == 0 {
		return 30
	}
	if limit < 1 || limit > 100 {
		return -1
	}
	return limit
}
func canonicalTransactionHash(owner int64, value validatedTransaction) string {
	canonical := fmt.Sprintf("v1|%d|%s|%s|%d|%s|%s", owner, value.date.Format(time.DateOnly), value.kind, value.categoryID, value.amount.String(), value.title)
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

func transactionFromDB(item db.Transaction, categoryName string) (Transaction, error) {
	amount, err := moneyFromNumeric(item.Amount)
	if err != nil {
		return Transaction{}, err
	}
	return Transaction{ID: item.ID, UserID: item.UserID, CategoryID: item.CategoryID, Type: item.Type, CategoryName: categoryName, Amount: amount.String(), TransactionDate: item.TransactionDate.Time, Title: item.Title, ClientRequestID: item.ClientRequestID, CreatedBy: item.CreatedBy, UpdatedBy: item.UpdatedBy, IsDelete: item.IsDelete, Version: item.Version, CreatedAt: item.CreatedAt.Time, UpdatedAt: item.UpdatedAt.Time}, nil
}
func transactionFromActiveRow(row db.ListActiveTransactionsRow) (Transaction, error) {
	return transactionFromDB(db.Transaction{ID: row.ID, UserID: row.UserID, CategoryID: row.CategoryID, Type: row.Type, Amount: row.Amount, TransactionDate: row.TransactionDate, Title: row.Title, ClientRequestID: row.ClientRequestID, RequestHash: row.RequestHash, CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy, IsDelete: row.IsDelete, Version: row.Version, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, row.CategoryName)
}
func transactionFromDeletedRow(row db.ListDeletedTransactionsRow) (Transaction, error) {
	return transactionFromDB(db.Transaction{ID: row.ID, UserID: row.UserID, CategoryID: row.CategoryID, Type: row.Type, Amount: row.Amount, TransactionDate: row.TransactionDate, Title: row.Title, ClientRequestID: row.ClientRequestID, RequestHash: row.RequestHash, CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy, IsDelete: row.IsDelete, Version: row.Version, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, row.CategoryName)
}

func paginateTransactions(items []Transaction, limit int32, scope TransactionScope, filter TransactionFilter) TransactionPage {
	var next *string
	if limit > 0 && len(items) > int(limit) {
		items = items[:limit]
		last := items[len(items)-1]
		keyTime := last.CreatedAt
		if filter.Deleted {
			keyTime = last.UpdatedAt
		}
		encoded := encodeCursor(cursorFor(scope, filter, "transactions", last.TransactionDate, keyTime, last.ID))
		next = &encoded
	}
	return TransactionPage{Items: items, NextCursor: next}
}
func cursorFor(scope TransactionScope, filter TransactionFilter, kind string, date, timestamp time.Time, id int64) transactionCursor {
	category := ""
	if filter.CategoryID != nil {
		category = strconv.FormatInt(*filter.CategoryID, 10)
	}
	return transactionCursor{Version: 1, Kind: kind, Actor: strconv.FormatInt(scope.Actor.UserID, 10), Owner: strconv.FormatInt(scope.Owner, 10), Type: filter.Type, Category: category, Start: filter.StartDate, End: filter.EndDate, Deleted: filter.Deleted, Admin: scope.Admin, Date: date.Format(time.DateOnly), Timestamp: timestamp.UTC().Format(time.RFC3339Nano), ID: strconv.FormatInt(id, 10)}
}
func encodeCursor(cursor transactionCursor) string {
	raw, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(raw)
}
func decodeTransactionCursor(value string) (*transactionCursor, error) {
	if value == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	var cursor transactionCursor
	if json.Unmarshal(raw, &cursor) != nil || cursor.Version != 1 {
		return nil, errors.New("invalid cursor")
	}
	return &cursor, nil
}
func cursorMatches(cursor *transactionCursor, scope TransactionScope, filter TransactionFilter, kind string) bool {
	if cursor == nil {
		return true
	}
	category := ""
	if filter.CategoryID != nil {
		category = strconv.FormatInt(*filter.CategoryID, 10)
	}
	if cursor.Kind != kind || cursor.Actor != strconv.FormatInt(scope.Actor.UserID, 10) || cursor.Owner != strconv.FormatInt(scope.Owner, 10) || cursor.Admin != scope.Admin || cursor.Type != filter.Type || cursor.Category != category || cursor.Start != filter.StartDate || cursor.End != filter.EndDate || cursor.Deleted != filter.Deleted {
		return false
	}
	if _, err := parseDate(cursor.Date); err != nil {
		return false
	}
	if kind == "transactions" {
		if _, err := time.Parse(time.RFC3339Nano, cursor.Timestamp); err != nil {
			return false
		}
		if id, err := strconv.ParseInt(cursor.ID, 10, 64); err != nil || id <= 0 {
			return false
		}
	}
	return true
}
func mustDate(value string) time.Time { date, _ := parseDate(value); return date }
func mustTimestamp(value string) time.Time {
	timestamp, _ := time.Parse(time.RFC3339Nano, value)
	return timestamp
}

func mustInt64(value string) int64 {
	id, _ := strconv.ParseInt(value, 10, 64)
	return id
}

func insertAudit(ctx context.Context, q *db.Queries, scope TransactionScope, resourceID *int64, action string) error {
	requestID := scope.RequestID
	if requestID == "" {
		requestID = uuid.NewString()
	}
	_, err := q.InsertAdminAccessEvent(ctx, db.InsertAdminAccessEventParams{ActorUserID: scope.Actor.UserID, TargetUserID: &scope.Owner, ResourceType: "transaction", ResourceID: resourceID, Action: action, Outcome: "success", RequestID: requestID, SafeMetadata: []byte(`{}`)})
	return err
}
func (s *Transactions) auditRead(ctx context.Context, scope TransactionScope, action string, resourceID int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ErrUnavailable
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	actor, err := q.GetActiveUserByID(ctx, scope.Actor.UserID)
	if err != nil || actor.Role != RoleAdmin {
		return ErrForbidden
	}
	if _, err := q.GetUserByID(ctx, scope.Owner); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return ErrUnavailable
	}
	var resource *int64
	if resourceID != 0 {
		resource = &resourceID
	}
	if err := insertAudit(ctx, q, scope, resource, action); err != nil {
		return ErrUnavailable
	}
	if err := tx.Commit(ctx); err != nil {
		return ErrUnavailable
	}
	return nil
}

func transactionFromActiveRowForGet(row db.GetActiveTransactionRow) (Transaction, error) {
	return transactionFromDB(db.Transaction{ID: row.ID, UserID: row.UserID, CategoryID: row.CategoryID, Type: row.Type, Amount: row.Amount, TransactionDate: row.TransactionDate, Title: row.Title, ClientRequestID: row.ClientRequestID, RequestHash: row.RequestHash, CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy, IsDelete: row.IsDelete, Version: row.Version, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, row.CategoryName)
}
func transactionFromDeletedRowForGet(row db.GetDeletedTransactionRow) (Transaction, error) {
	return transactionFromDB(db.Transaction{ID: row.ID, UserID: row.UserID, CategoryID: row.CategoryID, Type: row.Type, Amount: row.Amount, TransactionDate: row.TransactionDate, Title: row.Title, ClientRequestID: row.ClientRequestID, RequestHash: row.RequestHash, CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy, IsDelete: row.IsDelete, Version: row.Version, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, row.CategoryName)
}
