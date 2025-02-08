package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	pb "github.com/KEdore/explore/proto"
)

const DefaultLimit = 20

type ExploreServer struct {
	pb.UnimplementedExploreServiceServer
	db *sql.DB
}

func NewExploreServer(db *sql.DB) *ExploreServer {
	return &ExploreServer{db: db}
}

// PutDecision inserts (or updates) a decision without any timestamp logic.
func (s *ExploreServer) PutDecision(ctx context.Context, req *pb.PutDecisionRequest) (*pb.PutDecisionResponse, error) {
	query := `
		INSERT INTO decisions (actor_user_id, recipient_user_id, liked_recipient)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			liked_recipient = VALUES(liked_recipient)
	`

	_, err := s.db.ExecContext(
		ctx,
		query,
		req.GetActorUserId(),
		req.GetRecipientUserId(),
		req.GetLikedRecipient(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to put decision: %w", err)
	}

	// Check for mutual like.
	mutual := false
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

// ListLikedYou returns a list of users who liked the recipient.
func (s *ExploreServer) ListLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	offset := 0
	if token := req.GetPaginationToken(); token != "" {
		var err error
		offset, err = strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("invalid pagination token: %w", err)
		}
	}

	query := `
		SELECT actor_user_id
		FROM decisions
		WHERE recipient_user_id = ? AND liked_recipient = TRUE
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, query, req.GetRecipientUserId(), DefaultLimit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query liked decisions: %w", err)
	}
	defer rows.Close()

	var likers []*pb.ListLikedYouResponse_Liker
	count := 0
	for rows.Next() {
		var actorID string
		// Since there is no timestamp column, we set ts to 0.
		ts := int64(0)
		if err := rows.Scan(&actorID); err != nil {
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

	nextToken := ""
	if count == DefaultLimit {
		nextToken = strconv.Itoa(offset + DefaultLimit)
	}

	return &pb.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: &nextToken,
	}, nil
}

// ListNewLikedYou returns users who liked the recipient excluding those who have already liked back.
func (s *ExploreServer) ListNewLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	offset := 0
	if token := req.GetPaginationToken(); token != "" {
		var err error
		offset, err = strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("invalid pagination token: %w", err)
		}
	}

	query := `
		SELECT d.actor_user_id
		FROM decisions d
		WHERE d.recipient_user_id = ? AND d.liked_recipient = TRUE
		  AND NOT EXISTS (
			  SELECT 1 FROM decisions d2
			  WHERE d2.actor_user_id = ? AND d2.recipient_user_id = d.actor_user_id AND d2.liked_recipient = TRUE
		  )
		ORDER BY d.id DESC
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
		ts := int64(0)
		if err := rows.Scan(&actorID); err != nil {
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

	nextToken := ""
	if count == DefaultLimit {
		nextToken = strconv.Itoa(offset + DefaultLimit)
	}

	return &pb.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: &nextToken,
	}, nil
}

// CountLikedYou returns the count of users who liked the recipient.
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
