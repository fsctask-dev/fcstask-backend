package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fcstask/internal/middleware"
)

type mockMetricsClient struct {
	incCalls       []string
	gaugeCalls     []string
	histogramCalls []string
	lastTags       map[string]string
}

func (m *mockMetricsClient) Inc(name string, tags map[string]string) {
	m.incCalls = append(m.incCalls, name)
	m.lastTags = tags
}

func (m *mockMetricsClient) Add(name string, v float64, tags map[string]string) {
}

func (m *mockMetricsClient) Gauge(name string, v float64, tags map[string]string) {
	m.gaugeCalls = append(m.gaugeCalls, name)
}

func (m *mockMetricsClient) Histogram(name string, v float64, tags map[string]string) {
	m.histogramCalls = append(m.histogramCalls, name)
}

func (m *mockMetricsClient) Close() error {
	return nil
}

func TestMetricsMiddleware(t *testing.T) {
	mock := &mockMetricsClient{}
	mw := middleware.NewMetricsMiddleware(mock)

	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/test-path", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Проверяем статус
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Проверяем метрики
	if len(mock.incCalls) == 0 || mock.incCalls[0] != "http_requests_total" {
		t.Error("expected http_requests_total metric")
	}

	if len(mock.histogramCalls) == 0 || mock.histogramCalls[0] != "http_request_duration_seconds" {
		t.Error("expected http_request_duration_seconds metric")
	}

	// Проверяем теги
	if mock.lastTags["path"] != "/test-path" {
		t.Errorf("expected path /test-path, got %v", mock.lastTags["path"])
	}
	if mock.lastTags["method"] != "GET" {
		t.Errorf("expected method GET, got %v", mock.lastTags["method"])
	}
	if mock.lastTags["status"] != "200" {
		t.Errorf("expected status 200, got %v", mock.lastTags["status"])
	}
}

func TestMetricsMiddleware_ErrorStatus(t *testing.T) {
	mock := &mockMetricsClient{}
	mw := middleware.NewMetricsMiddleware(mock)

	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest("GET", "/not-found", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}

	if mock.lastTags["status"] != "404" {
		t.Errorf("expected status 404, got %v", mock.lastTags["status"])
	}
}

func TestMetricsMiddleware_ServerError(t *testing.T) {
	mock := &mockMetricsClient{}
	mw := middleware.NewMetricsMiddleware(mock)

	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if mock.lastTags["status"] != "500" {
		t.Errorf("expected status 500, got %v", mock.lastTags["status"])
	}
}

func TestMetricsMiddleware_PostRequest(t *testing.T) {
	mock := &mockMetricsClient{}
	mw := middleware.NewMetricsMiddleware(mock)

	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest("POST", "/api/resource", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if mock.lastTags["method"] != "POST" {
		t.Errorf("expected method POST, got %v", mock.lastTags["method"])
	}
	if mock.lastTags["status"] != "201" {
		t.Errorf("expected status 201, got %v", mock.lastTags["status"])
	}
}

type mockInFlightClient struct {
	gaugeCalls  []string
	gaugeValues []float64
}

func (m *mockInFlightClient) Inc(name string, tags map[string]string)            {}
func (m *mockInFlightClient) Add(name string, v float64, tags map[string]string) {}
func (m *mockInFlightClient) Gauge(name string, v float64, tags map[string]string) {
	m.gaugeCalls = append(m.gaugeCalls, name)
	m.gaugeValues = append(m.gaugeValues, v)
}
func (m *mockInFlightClient) Histogram(name string, v float64, tags map[string]string) {}
func (m *mockInFlightClient) Close() error                                             { return nil }

func TestInFlightMiddleware(t *testing.T) {
	mock := &mockInFlightClient{}
	mw := middleware.NewInflightMiddleware(mock)

	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Должен быть вызван дважды: +1 и -1
	if len(mock.gaugeCalls) != 2 {
		t.Errorf("expected 2 gauge calls, got %d", len(mock.gaugeCalls))
	}

	// Проверяем имена метрик
	for _, call := range mock.gaugeCalls {
		if call != "http_requests_in_flight" {
			t.Errorf("expected http_requests_in_flight, got %s", call)
		}
	}

	// Проверяем значения: сначала +1, потом -1
	if len(mock.gaugeValues) == 2 {
		if mock.gaugeValues[0] != 1 {
			t.Errorf("expected first value 1, got %f", mock.gaugeValues[0])
		}
		if mock.gaugeValues[1] != -1 {
			t.Errorf("expected second value -1, got %f", mock.gaugeValues[1])
		}
	}
}

func TestInFlightMiddleware_Concurrent(t *testing.T) {
	mock := &mockInFlightClient{}
	mw := middleware.NewInflightMiddleware(mock)

	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Запускаем несколько параллельных запросов
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Должно быть 10 вызовов (5 запросов × 2 вызова)
	if len(mock.gaugeCalls) != 10 {
		t.Errorf("expected 10 gauge calls, got %d", len(mock.gaugeCalls))
	}
}

func TestInFlightMiddleware_Panic(t *testing.T) {
	mock := &mockInFlightClient{}
	mw := middleware.NewInflightMiddleware(mock)

	handler := mw.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}

		// Даже при панике defer должен сработать
		if len(mock.gaugeCalls) != 2 {
			t.Errorf("expected 2 gauge calls even after panic, got %d", len(mock.gaugeCalls))
		}
	}()

	handler.ServeHTTP(rec, req)
}
