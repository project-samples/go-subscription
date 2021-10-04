package health

import (
	"context"
	"encoding/json"
	"net/http"
)

const (
	StatusUp   = "UP"
	StatusDown = "DOWN"
)
type Health struct {
	Status  string                 `mapstructure:"status" json:"status,omitempty" gorm:"column:status" bson:"status,omitempty" dynamodbav:"status,omitempty" firestore:"status,omitempty"`
	Data    map[string]interface{} `mapstructure:"data" json:"data,omitempty" gorm:"column:data" bson:"data,omitempty" dynamodbav:"data,omitempty" firestore:"data,omitempty"`
	Details map[string]Health      `mapstructure:"details" json:"details,omitempty" gorm:"column:details" bson:"details,omitempty" dynamodbav:"details,omitempty" firestore:"details,omitempty"`
}
type Checker interface {
	Name() string
	Check(ctx context.Context) (map[string]interface{}, error)
	Build(ctx context.Context, data map[string]interface{}, err error) map[string]interface{}
}

type Handler struct {
	Checkers []Checker
}

func NewHandler(checkers ...Checker) *Handler {
	return &Handler{Checkers: checkers}
}

func (c *Handler) Check(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	h := Check(ctx, c.Checkers)
	bytes, err := json.Marshal(h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if h.Status == StatusDown {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write(bytes)
}

func Check(ctx context.Context, services []Checker) Health {
	health := Health{}
	health.Status = StatusUp
	healths := make(map[string]Health)
	for _, service := range services {
		sub := Health{}
		c := context.Background()
		d0, err := service.Check(c)
		if err == nil {
			sub.Status = StatusUp
			if d0 != nil && len(d0) > 0 {
				sub.Data = d0
			}
		} else {
			sub.Status = StatusDown
			health.Status = StatusDown
			if d0 != nil {
				data := service.Build(c, d0, err)
				if data != nil && len(data) > 0 {
					sub.Data = data
				}
			} else {
				data := make(map[string]interface{}, 0)
				data["error"] = err.Error()
				sub.Data = data
			}
		}
		healths[service.Name()] = sub
	}
	if len(healths) > 0 {
		health.Details = healths
	}
	return health
}
