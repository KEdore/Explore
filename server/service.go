package server

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	pb "github.com/KEdore/explore/proto"
)

// InMemoryStore simulates a database.
type InMemoryStore struct {
	mu        sync.RWMutex
	decisions map[string]*pb.GetDecisionResponse // key: from_user_id + ":" + to_user_id
}

// NewInMemoryStore initializes the store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		decisions: make(map[string]*pb.GetDecisionResponse),
	}
}

func (s *InMemoryStore) key(from, to string) string {
	return fmt.Sprintf("%s:%s", from, to)
}

// SaveDecision saves or overwrites a decision.
func (s *InMemoryStore) SaveDecision(from, to string, decision pb.DecisionType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := s.key(from, to)
	s.decisions[key] = &pb.GetDecisionResponse{
		FromUserId: from,
		ToUserId:   to,
		Decision:   decision,
		Timestamp:  time.Now().Unix(),
	}
}

// GetDecision retrieves a decision.
func (s *InMemoryStore) GetDecision(from, to string) (*pb.GetDecisionResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := s.key(from, to)
	if d, exists := s.decisions[key]; exists {
		return d, nil
	}
	return nil, errors.New("decision not found")
}

// DeleteDecision removes a decision.
func (s *InMemoryStore) DeleteDecision(from, to string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := s.key(from, to)
	if _, exists := s.decisions[key]; exists {
		delete(s.decisions, key)
		return true
	}
	return false
}

// ListDecisions returns decisions for a given user (as the 'from' user).
func (s *InMemoryStore) ListDecisions(userID string, limit, offset int32) []*pb.Decision {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*pb.Decision
	for _, d := range s.decisions {
		if d.FromUserId == userID {
			result = append(result, &pb.Decision{
				FromUserId: d.FromUserId,
				ToUserId:   d.ToUserId,
				Decision:   d.Decision,
				Timestamp:  d.Timestamp,
			})
		}
	}
	// Simple pagination (note: for production, sorting and proper pagination are needed)
	start := int(offset)
	end := start + int(limit)
	if start > len(result) {
		return []*pb.Decision{}
	}
	if end > len(result) {
		end = len(result)
	}
	return result[start:end]
}

// ExploreServiceServer implements the gRPC service.
type ExploreServiceServer struct {
	pb.UnimplementedExploreServiceServer
	store *InMemoryStore
}

// NewExploreServiceServer creates a new server instance.
func NewExploreServiceServer(store *InMemoryStore) *ExploreServiceServer {
	return &ExploreServiceServer{store: store}
}

func (s *ExploreServiceServer) SubmitDecision(ctx context.Context, req *pb.SubmitDecisionRequest) (*pb.SubmitDecisionResponse, error) {
	// Overwrite any existing decision.
	s.store.SaveDecision(req.FromUserId, req.ToUserId, req.Decision)
	return &pb.SubmitDecisionResponse{
		Sucess:  true,
		Message: "Decision recorded successfully",
	}, nil
}

func (s *ExploreServiceServer) GetDecision(ctx context.Context, req *pb.GetDecisionRequest) (*pb.GetDecisionResponse, error) {
	d, err := s.store.GetDecision(req.FromUserId, req.ToUserId)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (s *ExploreServiceServer) ListDecisions(ctx context.Context, req *pb.ListDecisionsRequest) (*pb.ListDecisionsResponse, error) {
	decisions := s.store.ListDecisions(req.UserId, req.Limit, req.Offset)
	return &pb.ListDecisionsResponse{Decisions: decisions}, nil
}

func (s *ExploreServiceServer) DeleteDecision(ctx context.Context, req *pb.DeleteDecisionRequest) (*pb.DeleteDecisionResponse, error) {
	if deleted := s.store.DeleteDecision(req.FromUserId, req.ToUserId); deleted {
		return &pb.DeleteDecisionResponse{
			Success: true,
			Message: "Decision deleted successfully",
		}, nil
	}
	return &pb.DeleteDecisionResponse{
		Success: false,
		Message: "Decision not found",
	}, nil
}
