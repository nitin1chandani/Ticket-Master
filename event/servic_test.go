package event

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeEventRepo struct {
	createFn        func(ctx context.Context, input EventDetails) (*Event, error)
	getDetailsFn    func(ctx context.Context, id int64) (*EventDetailsWithAvailableTickets, error)
	lastCreateInput EventDetails
}

func (f *fakeEventRepo) CreateEvent(ctx context.Context, input EventDetails) (*Event, error) {
	f.lastCreateInput = input
	if f.createFn != nil {
		return f.createFn(ctx, input)
	}
	return &Event{}, nil
}

func (f *fakeEventRepo) GetByID(ctx context.Context, id int64) (*Event, error) {
	return nil, errors.New("not used")
}

func (f *fakeEventRepo) GetEventDetailsWithTickets(ctx context.Context, id int64) (*EventDetailsWithAvailableTickets, error) {
	if f.getDetailsFn != nil {
		return f.getDetailsFn(ctx, id)
	}
	return nil, nil
}

func TestEventService_CreateEvent_ValidationFailures(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name    string
		input   EventDetails
		wantErr error
	}{
		{
			name: "start time in past",
			input: EventDetails{
				Name:          "A",
				Location:      "Mumbai",
				PerformerName: "Artist",
				Description:   "Great live show",
				StartTime:     now.Add(-1 * time.Hour),
				EndTime:       now.Add(2 * time.Hour),
				TicketCount:   10,
				TicketPrice:   1000,
			},
			wantErr: ErrInvalidStartTime,
		},
		{
			name: "start time zero",
			input: EventDetails{
				Name:          "A",
				Location:      "Mumbai",
				PerformerName: "Artist",
				Description:   "Great live show",
				StartTime:     time.Time{},
				EndTime:       now.Add(2 * time.Hour),
				TicketCount:   10,
				TicketPrice:   1000,
			},
			wantErr: ErrInvalidStartTime,
		},
		{
			name: "end time zero",
			input: EventDetails{
				Name:          "A",
				Location:      "Mumbai",
				PerformerName: "Artist",
				Description:   "Great live show",
				StartTime:     now.Add(1 * time.Hour),
				EndTime:       time.Time{},
				TicketCount:   10,
				TicketPrice:   1000,
			},
			wantErr: ErrInvalidEndTime,
		},
		{
			name: "end time before start time",
			input: EventDetails{
				Name:          "A",
				Location:      "Mumbai",
				PerformerName: "Artist",
				Description:   "Great live show",
				StartTime:     now.Add(2 * time.Hour),
				EndTime:       now.Add(1 * time.Hour),
				TicketCount:   10,
				TicketPrice:   1000,
			},
			wantErr: ErrInvalidEndTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeEventRepo{}
			svc := NewEventService(repo)

			_, err := svc.CreateEvent(context.Background(), &tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestEventService_CreateEvent_Success_TrimsAndDefaults(t *testing.T) {
	now := time.Now().UTC()

	repo := &fakeEventRepo{
		createFn: func(ctx context.Context, input EventDetails) (*Event, error) {
			return &Event{
				ID:            1,
				Name:          input.Name,
				Location:      input.Location,
				PerformerName: input.PerformerName,
				Description:   input.Description,
				StartTime:     input.StartTime,
				EndTime:       input.EndTime,
			}, nil
		},
	}

	svc := NewEventService(repo)

	in := &EventDetails{
		Name:          "  Coldplay Live  ",
		Location:      "  Mumbai  ",
		PerformerName: "  Coldplay  ",
		Description:   "  Great concert night  ",
		StartTime:     now.Add(2 * time.Hour),
		EndTime:       now.Add(4 * time.Hour),
		TicketCount:   100,
		TicketPrice:   1999,
		TicketStatus:  "",
	}

	out, err := svc.CreateEvent(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out == nil {
		t.Fatalf("expected event, got nil")
	}

	if repo.lastCreateInput.Name != "Coldplay Live" {
		t.Fatalf("expected trimmed name, got %q", repo.lastCreateInput.Name)
	}
	if repo.lastCreateInput.Location != "Mumbai" {
		t.Fatalf("expected trimmed location, got %q", repo.lastCreateInput.Location)
	}
	if repo.lastCreateInput.PerformerName != "Coldplay" {
		t.Fatalf("expected trimmed performer, got %q", repo.lastCreateInput.PerformerName)
	}
	if repo.lastCreateInput.Description != "Great concert night" {
		t.Fatalf("expected trimmed description, got %q", repo.lastCreateInput.Description)
	}
	if repo.lastCreateInput.TicketStatus != "available" {
		t.Fatalf("expected default ticket status 'available', got %q", repo.lastCreateInput.TicketStatus)
	}
}

func TestEventService_CreateEvent_RepoError(t *testing.T) {
	repoErr := errors.New("db failed")
	repo := &fakeEventRepo{
		createFn: func(ctx context.Context, input EventDetails) (*Event, error) {
			return nil, repoErr
		},
	}

	svc := NewEventService(repo)

	_, err := svc.CreateEvent(context.Background(), &EventDetails{
		Name:          "Show",
		Location:      "Mumbai",
		PerformerName: "Artist",
		Description:   "Great live show",
		StartTime:     time.Now().UTC().Add(1 * time.Hour),
		EndTime:       time.Now().UTC().Add(2 * time.Hour),
		TicketCount:   100,
		TicketPrice:   999,
	})

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestEventService_GetAvailableTicketsWithEventDetails(t *testing.T) {
	expected := &EventDetailsWithAvailableTickets{
		Name:             "Show",
		Location:         "Mumbai",
		PerformerName:    "Artist",
		Description:      "Great show",
		AvailableTickets: []int64{1, 2, 3},
	}

	repo := &fakeEventRepo{
		getDetailsFn: func(ctx context.Context, id int64) (*EventDetailsWithAvailableTickets, error) {
			if id != 99 {
				t.Fatalf("expected id 99, got %d", id)
			}
			return expected, nil
		},
	}
	svc := NewEventService(repo)

	got, err := svc.GetAvailableTicketsWithEventDetails(context.Background(), 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected result, got nil")
	}
	if len(got.AvailableTickets) != 3 {
		t.Fatalf("expected 3 tickets, got %d", len(got.AvailableTickets))
	}
}

func TestEventService_GetAvailableTicketsWithEventDetails_RepoError(t *testing.T) {
	repoErr := errors.New("query failed")
	repo := &fakeEventRepo{
		getDetailsFn: func(ctx context.Context, id int64) (*EventDetailsWithAvailableTickets, error) {
			return nil, repoErr
		},
	}
	svc := NewEventService(repo)

	_, err := svc.GetAvailableTicketsWithEventDetails(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
