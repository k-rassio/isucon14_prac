package main

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB
var latestRideStatusCache = struct {
	mu sync.RWMutex
	m  map[string]string
}{m: make(map[string]string)}

type totalDistanceEntry struct {
	Distance  int
	UpdatedAt time.Time
}

var totalDistanceCache = struct {
	mu sync.RWMutex
	m  map[string]totalDistanceEntry
}{m: make(map[string]totalDistanceEntry)}

type RideCache struct {
	mu    sync.RWMutex
	items map[string]Ride // ridesテーブルのデータ.
}

var rideCache = &RideCache{
	items: make(map[string]Ride),
}

type ChairRideCache struct {
	mu    sync.RWMutex
	items map[string][]string // chairID -> []rideID
}

var chairRideCache = &ChairRideCache{
	items: make(map[string][]string),
}

func loadLatestRideStatusCache(ctx context.Context) error {
	rows := []struct {
		RideID string `db:"ride_id"`
		Status string `db:"status"`
	}{}
	query := `
SELECT rs.ride_id, rs.status
FROM ride_statuses rs
JOIN (
  SELECT ride_id, MAX(created_at) AS max_created_at FROM ride_statuses GROUP BY ride_id
) t ON rs.ride_id = t.ride_id AND rs.created_at = t.max_created_at
`
	if err := db.SelectContext(ctx, &rows, query); err != nil {
		return err
	}
	latestRideStatusCache.mu.Lock()
	defer latestRideStatusCache.mu.Unlock()
	for _, r := range rows {
		latestRideStatusCache.m[r.RideID] = r.Status
	}
	return nil
}

func setLatestRideStatus(rideID, status string) {
	latestRideStatusCache.mu.Lock()
	latestRideStatusCache.m[rideID] = status
	latestRideStatusCache.mu.Unlock()
}

func resetTotalDistanceCache() {
	totalDistanceCache.mu.Lock()
	defer totalDistanceCache.mu.Unlock()
	totalDistanceCache.m = make(map[string]totalDistanceEntry)
}

func setTotalDistanceCacheValue(chairID string, distance int, updatedAt time.Time) {
	totalDistanceCache.mu.Lock()
	defer totalDistanceCache.mu.Unlock()
	totalDistanceCache.m[chairID] = totalDistanceEntry{Distance: distance, UpdatedAt: updatedAt}
}

func getTotalDistanceCacheValue(chairID string) (int, time.Time, bool) {
	totalDistanceCache.mu.RLock()
	defer totalDistanceCache.mu.RUnlock()
	entry, ok := totalDistanceCache.m[chairID]
	if !ok {
		return 0, time.Time{}, false
	}
	return entry.Distance, entry.UpdatedAt, true
}

func buildTotalDistanceEntry(locations []ChairLocation) (int, time.Time, bool) {
	if len(locations) == 0 {
		return 0, time.Time{}, false
	}

	totalDistance := 0
	for i := 1; i < len(locations); i++ {
		totalDistance += abs(locations[i].Latitude-locations[i-1].Latitude) + abs(locations[i].Longitude-locations[i-1].Longitude)
	}
	return totalDistance, locations[len(locations)-1].CreatedAt, true
}

func loadTotalDistanceCache(ctx context.Context) error {
	var chairIDs []string
	if err := db.SelectContext(ctx, &chairIDs, `SELECT id FROM chairs`); err != nil {
		return err
	}

	totalDistanceCache.mu.Lock()
	defer totalDistanceCache.mu.Unlock()
	totalDistanceCache.m = make(map[string]totalDistanceEntry, len(chairIDs))

	for _, chairID := range chairIDs {
		var locations []ChairLocation
		if err := db.SelectContext(ctx, &locations, `SELECT chair_id, latitude, longitude, created_at FROM chair_locations WHERE chair_id = ? ORDER BY created_at`, chairID); err != nil {
			return err
		}
		if distance, updatedAt, ok := buildTotalDistanceEntry(locations); ok {
			totalDistanceCache.m[chairID] = totalDistanceEntry{Distance: distance, UpdatedAt: updatedAt}
		}
	}
	return nil
}

func getLatestRideStatusFromCache(rideID string) (string, bool) {
	latestRideStatusCache.mu.RLock()
	s, ok := latestRideStatusCache.m[rideID]
	latestRideStatusCache.mu.RUnlock()
	return s, ok
}

func main() {
	mux := setup()
	slog.Info("Listening on :8080")
	http.ListenAndServe(":8080", mux)
}

func setup() http.Handler {
	host := os.Getenv("ISUCON_DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("ISUCON_DB_PORT")
	if port == "" {
		port = "3306"
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		panic(fmt.Sprintf("failed to convert DB port number from ISUCON_DB_PORT environment variable into int: %v", err))
	}
	user := os.Getenv("ISUCON_DB_USER")
	if user == "" {
		user = "isucon"
	}
	password := os.Getenv("ISUCON_DB_PASSWORD")
	if password == "" {
		password = "isucon"
	}
	dbname := os.Getenv("ISUCON_DB_NAME")
	if dbname == "" {
		dbname = "isuride"
	}

	dbConfig := mysql.NewConfig()
	dbConfig.User = user
	dbConfig.Passwd = password
	dbConfig.Addr = net.JoinHostPort(host, port)
	dbConfig.Net = "tcp"
	dbConfig.DBName = dbname
	dbConfig.ParseTime = true

	_db, err := sqlx.Connect("mysql", dbConfig.FormatDSN())
	if err != nil {
		panic(err)
	}
	db = _db

	// load latest ride status cache
	if err := loadLatestRideStatusCache(context.Background()); err != nil {
		slog.Warn("failed to load latest ride status cache", "err", err)
	}
	if err := loadTotalDistanceCache(context.Background()); err != nil {
		slog.Warn("failed to load total distance cache", "err", err)
	}

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.HandleFunc("POST /api/initialize", postInitialize)

	// app handlers
	{
		mux.HandleFunc("POST /api/app/users", appPostUsers)

		authedMux := mux.With(appAuthMiddleware)
		authedMux.HandleFunc("POST /api/app/payment-methods", appPostPaymentMethods)
		authedMux.HandleFunc("GET /api/app/rides", appGetRides)
		authedMux.HandleFunc("POST /api/app/rides", appPostRides)
		authedMux.HandleFunc("POST /api/app/rides/estimated-fare", appPostRidesEstimatedFare)
		authedMux.HandleFunc("POST /api/app/rides/{ride_id}/evaluation", appPostRideEvaluatation)
		authedMux.HandleFunc("GET /api/app/notification", appGetNotification)
		authedMux.HandleFunc("GET /api/app/nearby-chairs", appGetNearbyChairs)
	}

	// owner handlers
	{
		mux.HandleFunc("POST /api/owner/owners", ownerPostOwners)

		authedMux := mux.With(ownerAuthMiddleware)
		authedMux.HandleFunc("GET /api/owner/sales", ownerGetSales)
		authedMux.HandleFunc("GET /api/owner/chairs", ownerGetChairs)
	}

	// chair handlers
	{
		mux.HandleFunc("POST /api/chair/chairs", chairPostChairs)

		authedMux := mux.With(chairAuthMiddleware)
		authedMux.HandleFunc("POST /api/chair/activity", chairPostActivity)
		authedMux.HandleFunc("POST /api/chair/coordinate", chairPostCoordinate)
		authedMux.HandleFunc("GET /api/chair/notification", chairGetNotification)
		authedMux.HandleFunc("POST /api/chair/rides/{ride_id}/status", chairPostRideStatus)
	}

	// internal handlers
	{
		mux.HandleFunc("GET /api/internal/matching", internalGetMatching)
	}

	return mux
}

type postInitializeRequest struct {
	PaymentServer string `json:"payment_server"`
}

type postInitializeResponse struct {
	Language string `json:"language"`
}

func postInitialize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &postInitializeRequest{}
	if err := bindJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if out, err := exec.Command("../sql/init.sh").CombinedOutput(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to initialize: %s: %w", string(out), err))
		return
	}

	if err := loadTotalDistanceCache(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to rebuild total_distance cache: %w", err))
		return
	}

	if _, err := db.ExecContext(ctx, "UPDATE settings SET value = ? WHERE name = 'payment_gateway_url'", req.PaymentServer); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, postInitializeResponse{Language: "go"})
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type Coordinate struct {
	Latitude  int `json:"latitude"`
	Longitude int `json:"longitude"`
}

func bindJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	buf, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(buf)
}

func writeError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(statusCode)
	buf, marshalError := json.Marshal(map[string]string{"message": err.Error()})
	if marshalError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"marshaling error failed"}`))
		return
	}
	w.Write(buf)

	slog.Error("error response wrote", "err", err)
}

func secureRandomStr(b int) string {
	k := make([]byte, b)
	if _, err := crand.Read(k); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", k)
}
