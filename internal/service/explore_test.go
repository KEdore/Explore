package service_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/KEdore/explore/internal/service"
	pb "github.com/KEdore/explore/proto"
)

// Helper function to get a pointer to a string.
func strPtr(s string) *string {
	return &s
}

// TestPutDecision_NoMutual tests PutDecision when there is no mutual like.
func TestPutDecision_NoMutual(t *testing.T) {
	// Setup sqlmock database.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	srv := service.NewExploreServer(db)
	ctx := context.Background()

	// Expect the Exec call for inserting/updating the decision.
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO decisions (actor_user_id, recipient_user_id, liked_recipient)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			liked_recipient = VALUES(liked_recipient)
	`)).
		WithArgs("actor1", "recipient1", true).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect the mutual like query returning count = 0.
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM decisions
		WHERE actor_user_id = ? AND recipient_user_id = ? AND liked_recipient = TRUE
	`)).
		WithArgs("recipient1", "actor1").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	req := &pb.PutDecisionRequest{
		ActorUserId:     "actor1",
		RecipientUserId: "recipient1",
		LikedRecipient:  true,
	}

	res, err := srv.PutDecision(ctx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if res.MutualLikes != false {
		t.Errorf("expected MutualLikes to be false, got %v", res.MutualLikes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

// TestPutDecision_Mutual tests PutDecision when a mutual like exists.
func TestPutDecision_Mutual(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	srv := service.NewExploreServer(db)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO decisions (actor_user_id, recipient_user_id, liked_recipient)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			liked_recipient = VALUES(liked_recipient)
	`)).
		WithArgs("actor1", "recipient1", true).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Simulate a mutual like: count > 0.
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM decisions
		WHERE actor_user_id = ? AND recipient_user_id = ? AND liked_recipient = TRUE
	`)).
		WithArgs("recipient1", "actor1").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1))

	req := &pb.PutDecisionRequest{
		ActorUserId:     "actor1",
		RecipientUserId: "recipient1",
		LikedRecipient:  true,
	}

	res, err := srv.PutDecision(ctx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if res.MutualLikes != true {
		t.Errorf("expected MutualLikes to be true, got %v", res.MutualLikes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

// TestPutDecision_NotLiked tests PutDecision when the decision is not a like (i.e. false).
// In this case, the mutual like check should not be performed.
func TestPutDecision_NotLiked(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	srv := service.NewExploreServer(db)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO decisions (actor_user_id, recipient_user_id, liked_recipient)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			liked_recipient = VALUES(liked_recipient)
	`)).
		WithArgs("actor2", "recipient2", false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := &pb.PutDecisionRequest{
		ActorUserId:     "actor2",
		RecipientUserId: "recipient2",
		LikedRecipient:  false,
	}

	res, err := srv.PutDecision(ctx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if res.MutualLikes != false {
		t.Errorf("expected MutualLikes to be false, got %v", res.MutualLikes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

// TestListLikedYou tests the ListLikedYou endpoint when results are returned.
func TestListLikedYou(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	srv := service.NewExploreServer(db)
	ctx := context.Background()

	// Create rows to simulate two likers.
	rows := sqlmock.NewRows([]string{"actor_user_id"}).
		AddRow("actor1").
		AddRow("actor2")
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT actor_user_id
		FROM decisions
		WHERE recipient_user_id = ? AND liked_recipient = TRUE
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`)).
		WithArgs("recipient1", service.DefaultLimit, 0).
		WillReturnRows(rows)

	req := &pb.ListLikedYouRequest{
		RecipientUserId: "recipient1",
		PaginationToken: strPtr(""),
	}
	res, err := srv.ListLikedYou(ctx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(res.Likers) != 2 {
		t.Errorf("expected 2 likers, got %d", len(res.Likers))
	}
	// If fewer than DefaultLimit rows are returned, NextPaginationToken should be empty.
	if res.NextPaginationToken != nil && *res.NextPaginationToken != "" {
		t.Errorf("expected empty pagination token, got %v", *res.NextPaginationToken)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

// TestListNewLikedYou tests the ListNewLikedYou endpoint.
func TestListNewLikedYou(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	srv := service.NewExploreServer(db)
	ctx := context.Background()

	// Create rows to simulate one liker who hasn't been liked back.
	rows := sqlmock.NewRows([]string{"actor_user_id"}).
		AddRow("actor3")
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT d.actor_user_id
		FROM decisions d
		WHERE d.recipient_user_id = ? AND d.liked_recipient = TRUE
		  AND NOT EXISTS (
			  SELECT 1 FROM decisions d2
			  WHERE d2.actor_user_id = ? AND d2.recipient_user_id = d.actor_user_id AND d2.liked_recipient = TRUE
		  )
		ORDER BY d.id DESC
		LIMIT ? OFFSET ?
	`)).
		WithArgs("recipient2", "recipient2", service.DefaultLimit, 0).
		WillReturnRows(rows)

	req := &pb.ListLikedYouRequest{
		RecipientUserId: "recipient2",
		PaginationToken: strPtr(""),
	}
	res, err := srv.ListNewLikedYou(ctx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(res.Likers) != 1 {
		t.Errorf("expected 1 liker, got %d", len(res.Likers))
	}
	if res.NextPaginationToken != nil && *res.NextPaginationToken != "" {
		t.Errorf("expected empty pagination token, got %v", *res.NextPaginationToken)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

// TestCountLikedYou tests the CountLikedYou endpoint.
func TestCountLikedYou(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	srv := service.NewExploreServer(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM decisions
		WHERE recipient_user_id = ? AND liked_recipient = TRUE
	`)).
		WithArgs("recipient3").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(5))

	req := &pb.CountLikedYouRequest{
		RecipientUserId: "recipient3",
	}
	res, err := srv.CountLikedYou(ctx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if res.Count != 5 {
		t.Errorf("expected count 5, got %d", res.Count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
