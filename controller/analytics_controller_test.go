package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sebaespinosa/test_NF/model"
	"github.com/stretchr/testify/assert"
)

type stubAnalyticsService struct {
	resp      *model.IrrigationAnalyticsResponse
	err       error
	lastLimit int
	lastPage  int
}

func (s *stubAnalyticsService) GetAnalytics(ctx context.Context, farmID uint, startDate, endDate *time.Time, sectorID *uint, aggregation string, page, limit int) (*model.IrrigationAnalyticsResponse, error) {
	s.lastLimit = limit
	s.lastPage = page
	return s.resp, s.err
}

func newTestRouter(svc AnalyticsService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ctrl := &AnalyticsController{service: svc}
	r.GET("/v1/farms/:farm_id/irrigation/analytics", ctrl.GetAnalytics)
	return r
}

func TestGetAnalytics_StatusOK(t *testing.T) {
	svc := &stubAnalyticsService{
		resp: &model.IrrigationAnalyticsResponse{
			Metrics:          model.AnalyticsMetrics{TotalIrrigationEvents: 1},
			SamePeriod1Y:     &model.YoYComparison{DataIncomplete: false},
			SamePeriod2Y:     &model.YoYComparison{DataIncomplete: false},
			TimeSeries:       model.TimeSeries{Pagination: model.PaginationMetadata{TotalCount: 1, TotalPages: 1}},
			PeriodComparison: &model.PeriodComparisonSet{},
		},
	}
	router := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/farms/1/irrigation/analytics?aggregation=weekly&limit=20&page=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 20, svc.lastLimit)
	assert.Equal(t, 2, svc.lastPage)
}

func TestGetAnalytics_StatusPartialContent(t *testing.T) {
	svc := &stubAnalyticsService{
		resp: &model.IrrigationAnalyticsResponse{
			SamePeriod1Y: &model.YoYComparison{DataIncomplete: true},
			TimeSeries:   model.TimeSeries{Pagination: model.PaginationMetadata{TotalCount: 0, TotalPages: 0}},
		},
	}
	router := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/farms/1/irrigation/analytics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusPartialContent, w.Code)
}

func TestGetAnalytics_InvalidDate(t *testing.T) {
	svc := &stubAnalyticsService{}
	router := newTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/farms/1/irrigation/analytics?start_date=bad-date", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
