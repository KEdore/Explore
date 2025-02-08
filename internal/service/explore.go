package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	pb "github.com/KEdore/explore/proto"
)

// Exported constant for pagination limit.
const DefaultLimit = 20

// ExploreServer implements the ExploreService as defined in the proto file.
// It holds a pointer to an SQL database and provides methods to record and query decisions.
type ExploreServer struct {
	pb.UnimplementedExploreServiceServer

	db *sql.DB
}

// NewExploreServer returns a new ExploreServer instance with the provided database connection.
func NewExploreServer(db *sql.DB) *ExploreServer {
	return &ExploreServer{db: db}
}

// PutDecision records or updates a decision from one user (actor) to another (recipient).
// If the actor liked the recipient, it also checks whether the recipient has liked the actor in return,
// which would indicate a mutual like.
func (s *ExploreServer) PutDecision(ctx context.Context, req *pb.PutDecisionRequest) (*pb.PutDecisionResponse, error) {
	// SQL query to insert or update a decision based on a unique key (actor_user_id, recipient_user_id).
	query := `
		INSERT INTO decisions (actor_user_id, recipient_user_id, liked_recipient, timestamp)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			liked_recipient = VALUES(liked_recipient),
			timestamp = VALUES(timestamp)
	`
	// Use the current Unix timestamp (in seconds).
	timestamp := time.Now().Unix()

	// Execute the insert/update query.
	_, err := s.db.ExecContext(
		ctx,
		query,
		req.GetActorUserId(),
		req.GetRecipientUserId(),
		req.GetLikedRecipient(),
		timestamp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to put decision: %w", err)
	}

	// Determine if a mutual like exists.
	mutual := false
	// Only check for mutual like if the actor has liked the recipient.
	if req.GetLikedRecipient() {
		mutualQuery := `
			SELECT COUNT(*) FROM decisions
			WHERE actor_user_id = ? AND recipient_user_id = ? AND liked_recipient = TRUE
		`
		var count int
		err = s.db.QueryRowContext(ctx, mutualQuery, req.GetRecipientUserId(), req.GetActorUserId()).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to check mutual like: %w", err)
		}
		if count > 0 {
			mutual = true
		}
	}

	return &pb.PutDecisionResponse{
		MutualLikes: mutual,
	}, nil
}

// ListLikedYou returns a list of users who have liked the specified recipient.
// An optional pagination token (a string pointer) is used for pagination.
func (s *ExploreServer) ListLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	// Determine the numeric offset from the pagination token.
	offset := 0
	if token := req.GetPaginationToken();  token != "" {
		var err error
		offset, err = strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("invalid pagination token: %w", err)
		}
	}

	query := `
		SELECT actor_user_id, timestamp
		FROM decisions
		WHERE recipient_user_id = ? AND liked_recipient = TRUE
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`
	// Execute the query to retrieve likers.
	rows, err := s.db.QueryContext(ctx, query, req.GetRecipientUserId(), DefaultLimit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query liked decisions: %w", err)
	}
	defer rows.Close()

	var likers []*pb.ListLikedYouResponse_Liker
	count := 0
	for rows.Next() {
		var actorID string
		var ts int64
		if err := rows.Scan(&actorID, &ts); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		// Convert the timestamp to uint64 as defined in the proto.
		likers = append(likers, &pb.ListLikedYouResponse_Liker{
			ActorId:       actorID,
			UnixTimestamp: uint64(ts),
		})
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Prepare the next pagination token if the page is full.
	nextToken := ""
	if count == DefaultLimit {
		nextToken = strconv.Itoa(offset + DefaultLimit)
	}

	return &pb.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: &nextToken,
	}, nil
}

// ListNewLikedYou returns a list of users who have liked the recipient,
// excluding those users who the recipient has already liked in return.
func (s *ExploreServer) ListNewLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	// Determine the numeric offset from the pagination token.
	offset := 0
	if token := req.GetPaginationToken(); token != "" {
		var err error
		offset, err = strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("invalid pagination token: %w", err)
		}
	}

	// Query that selects likers who have not been liked back by the recipient.
	query := `
		SELECT d.actor_user_id, d.timestamp
		FROM decisions d
		WHERE d.recipient_user_id = ? AND d.liked_recipient = TRUE
		  AND NOT EXISTS (
			  SELECT 1 FROM decisions d2
			  WHERE d2.actor_user_id = ? AND d2.recipient_user_id = d.actor_user_id AND d2.liked_recipient = TRUE
		  )
		ORDER BY d.timestamp DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, query, req.GetRecipientUserId(), req.GetRecipientUserId(), DefaultLimit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query new liked decisions: %w", err)
	}
	defer rows.Close()

	var likers []*pb.ListLikedYouResponse_Liker
	count := 0
	for rows.Next() {
		var actorID string
		var ts int64
		if err := rows.Scan(&actorID, &ts); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		likers = append(likers, &pb.ListLikedYouResponse_Liker{
			ActorId:       actorID,
			UnixTimestamp: uint64(ts),
		})
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Prepare the next pagination token if the page is full.
	nextToken := ""
	if count == DefaultLimit {
		nextToken = strconv.Itoa(offset + DefaultLimit)
	}

	return &pb.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: &nextToken,
	}, nil
}

// CountLikedYou returns the total count of users who have liked the recipient.
func (s *ExploreServer) CountLikedYou(ctx context.Context, req *pb.CountLikedYouRequest) (*pb.CountLikedYouResponse, error) {
	query := `
		SELECT COUNT(*)
		FROM decisions
		WHERE recipient_user_id = ? AND liked_recipient = TRUE
	`
	var count int
	err := s.db.QueryRowContext(ctx, query, req.GetRecipientUserId()).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count liked decisions: %w", err)
	}
	return &pb.CountLikedYouResponse{
		Count: uint64(count),
	}, nil
}
