package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// latestChairLocation holds the most recent reported coordinates for each
// chair in memory.  By keeping this information locally we avoid querying
// the database on every /api/chair/coordinate request (the select at line
// 117 used to do that).  A sync.Map is sufficient because entries are
// written once per request and read concurrently.
var latestChairLocation sync.Map // map[string]ChairLocation

// request body for /api/chair
// this struct was accidentally removed during editing; restore it.
type chairPostChairsRequest struct {
	Name               string `json:"name"`
	Model              string `json:"model"`
	ChairRegisterToken string `json:"chair_register_token"`
}

type chairPostChairsResponse struct {
	ID      string `json:"id"`
	OwnerID string `json:"owner_id"`
}

func chairPostChairs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &chairPostChairsRequest{}
	if err := bindJSON(r, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Model == "" || req.ChairRegisterToken == "" {
		writeError(w, http.StatusBadRequest, errors.New("some of required fields(name, model, chair_register_token) are empty"))
		return
	}

	owner := &Owner{}
	if err := db.GetContext(ctx, owner, "SELECT * FROM owners WHERE chair_register_token = ?", req.ChairRegisterToken); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, errors.New("invalid chair_register_token"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	chairID := ulid.Make().String()
	accessToken := secureRandomStr(32)

	_, err := db.ExecContext(
		ctx,
		"INSERT INTO chairs (id, owner_id, name, model, is_active, access_token) VALUES (?, ?, ?, ?, ?, ?)",
		chairID, owner.ID, req.Name, req.Model, false, accessToken,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	// 2. キャッシュ用の構造体を作成（req の値を正確に反映）
	newChair := &Chair{
		ID:          chairID,     // 生成したID
		OwnerID:     owner.ID,    // ログイン中のオーナーID
		Name:        req.Name,    // ★リクエストから取得
		Model:       req.Model,   // ★リクエストから取得
		IsActive:    false,       // INSERT時と同じデフォルト値
		AccessToken: accessToken, // 生成したトークン
	}

	// 3. キャッシュを更新
	cacheLock.Lock()
	chairCache[accessToken] = newChair
	cacheLock.Unlock()

	http.SetCookie(w, &http.Cookie{
		Path:  "/",
		Name:  "chair_session",
		Value: accessToken,
	})

	writeJSON(w, http.StatusCreated, &chairPostChairsResponse{
		ID:      chairID,
		OwnerID: owner.ID,
	})
}

type postChairActivityRequest struct {
	IsActive bool `json:"is_active"`
}

func chairPostActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chair := ctx.Value("chair").(*Chair)
	req := &postChairActivityRequest{}
	if err := bindJSON(r, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := db.ExecContext(ctx, "UPDATE chairs SET is_active = ? WHERE id = ?", req.IsActive, chair.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	cacheLock.Lock()
	chair.IsActive = req.IsActive // ポインタ経由で実体を書き換え
	cacheLock.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

type chairPostCoordinateResponse struct {
	RecordedAt int64 `json:"recorded_at"`
}

func chairPostCoordinate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &Coordinate{}
	if err := bindJSON(r, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chair := ctx.Value("chair").(*Chair)

	// look up the previous location from memory cache first.  If there's
	// no entry yet we still need to fall back to the database so that we
	// calculate correct distance after a restart or cache miss.  When the
	// DB query succeeds we store the result back into the cache.
	var prevLocation ChairLocation
	hasPrev := false
	if v, ok := latestChairLocation.Load(chair.ID); ok {
		prevLocation = v.(ChairLocation)
		hasPrev = true
	} else {
		// cache miss, try the database once
		if err := db.GetContext(ctx, &prevLocation, `SELECT * FROM chair_locations WHERE chair_id = ? ORDER BY created_at DESC LIMIT 1`, chair.ID); err == nil {
			hasPrev = true
			latestChairLocation.Store(chair.ID, prevLocation)
		} else if !errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	chairLocationID := ulid.Make().String()
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO chair_locations (id, chair_id, latitude, longitude) VALUES (?, ?, ?, ?)`,
		chairLocationID, chair.ID, req.Latitude, req.Longitude,
	); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO chair_latest_locations (chair_id, latitude, longitude) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE latitude = VALUES(latitude), longitude = VALUES(longitude)`, chair.ID, req.Latitude, req.Longitude); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	location := &ChairLocation{}

	location.Latitude = req.Latitude
	location.Longitude = req.Longitude
	location.CreatedAt = time.Now()

	var distance int
	if hasPrev {
		distance = calculateDistance(prevLocation.Latitude, prevLocation.Longitude, location.Latitude, location.Longitude)
	} else {
		distance = 0
	}

	type rideWithLatest struct {
		ID                   string `db:"id"`
		PickupLatitude       int    `db:"pickup_latitude"`
		PickupLongitude      int    `db:"pickup_longitude"`
		DestinationLatitude  int    `db:"destination_latitude"`
		DestinationLongitude int    `db:"destination_longitude"`
	}

	var rwsData Ride
	var rideInCache bool

	rideCache.mu.RLock()
	rwsData, rideInCache = rideCache.items[chair.ID]
	rideCache.mu.RUnlock()

	if !rideInCache {
		rws := &rideWithLatest{}
		query := `
	SELECT
		r.id,
		r.pickup_latitude,
		r.pickup_longitude,
		r.destination_latitude,
		r.destination_longitude
	FROM rides r
	WHERE r.chair_id = ?
	ORDER BY r.updated_at DESC LIMIT 1
	`
		if err := tx.GetContext(ctx, rws, query, chair.ID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// ride がない場合は処理をスキップ
				if err := tx.Commit(); err != nil {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
				writeJSON(w, http.StatusOK, &chairPostCoordinateResponse{
					RecordedAt: location.CreatedAt.UnixMilli(),
				})
				return
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		rwsData = Ride{
			ID:                   rws.ID,
			PickupLatitude:       rws.PickupLatitude,
			PickupLongitude:      rws.PickupLongitude,
			DestinationLatitude:  rws.DestinationLatitude,
			DestinationLongitude: rws.DestinationLongitude,
		}

		rideCache.mu.Lock()
		rideCache.items[chair.ID] = rwsData
		rideCache.mu.Unlock()
	}

	statusStr, err := getLatestRideStatus(ctx, tx, rwsData.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	var updatedStatus string

	if statusStr != "COMPLETED" && statusStr != "CANCELED" {
		if req.Latitude == rwsData.PickupLatitude && req.Longitude == rwsData.PickupLongitude && statusStr == "ENROUTE" {
			newStatusID := ulid.Make().String()
			if _, err := tx.ExecContext(ctx, "INSERT INTO ride_statuses (id, ride_id, status) VALUES (?, ?, ?)", newStatusID, rwsData.ID, "PICKUP"); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			// setLatestRideStatus(rwsData.ID, "PICKUP")
			updatedStatus = "PICKUP"
		}

		if req.Latitude == rwsData.DestinationLatitude && req.Longitude == rwsData.DestinationLongitude && statusStr == "CARRYING" {
			newStatusID := ulid.Make().String()
			if _, err := tx.ExecContext(ctx, "INSERT INTO ride_statuses (id, ride_id, status) VALUES (?, ?, ?)", newStatusID, rwsData.ID, "ARRIVED"); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			// setLatestRideStatus(rwsData.ID, "ARRIVED")
			updatedStatus = "ARRIVED"
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if updatedStatus != "" {
		setLatestRideStatus(rwsData.ID, updatedStatus)
	}

	// update the in‑memory cache with the new location so future requests
	// can compute distance without hitting the database.
	latestChairLocation.Store(chair.ID, ChairLocation{
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		CreatedAt: location.CreatedAt,
	})

	currentDistance, _, _ := getTotalDistanceCacheValue(chair.ID)
	setTotalDistanceCacheValue(chair.ID, currentDistance+distance, location.CreatedAt)

	writeJSON(w, http.StatusOK, &chairPostCoordinateResponse{
		RecordedAt: location.CreatedAt.UnixMilli(),
	})
}

type simpleUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type chairGetNotificationResponse struct {
	Data         *chairGetNotificationResponseData `json:"data"`
	RetryAfterMs int                               `json:"retry_after_ms"`
}

type chairGetNotificationResponseData struct {
	RideID                string     `json:"ride_id"`
	User                  simpleUser `json:"user"`
	PickupCoordinate      Coordinate `json:"pickup_coordinate"`
	DestinationCoordinate Coordinate `json:"destination_coordinate"`
	Status                string     `json:"status"`
}

func chairGetNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chair := ctx.Value("chair").(*Chair)

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()
	ride := &Ride{}
	var rideInCache bool
	yetSentRideStatus := RideStatus{}
	status := ""

	chairRideCache.mu.RLock()
	tmpRide, ok := rideCache.items[chair.ID]
	if ok {
		*ride = tmpRide // ポインタが指す中身を書き換える
		rideInCache = true
	}
	chairRideCache.mu.RUnlock()

	if !rideInCache {
		if err := tx.GetContext(ctx, ride, `SELECT * FROM rides WHERE chair_id = ? ORDER BY updated_at DESC LIMIT 1`, chair.ID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSON(w, http.StatusOK, &chairGetNotificationResponse{
					RetryAfterMs: 1000,
				})
				return
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		rideCache.mu.Lock()
		rideCache.items[chair.ID] = *ride
		rideCache.mu.Unlock()
	}

	if err := tx.GetContext(ctx, &yetSentRideStatus, `SELECT * FROM ride_statuses WHERE ride_id = ? AND chair_sent_at IS NULL ORDER BY created_at ASC LIMIT 1`, ride.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			status, err = getLatestRideStatus(ctx, tx, ride.ID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
		} else {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		status = yetSentRideStatus.Status
	}

	user := &User{}
	err = tx.GetContext(ctx, user, "SELECT * FROM users WHERE id = ? FOR SHARE", ride.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if yetSentRideStatus.ID != "" {
		_, err := tx.ExecContext(ctx, `UPDATE ride_statuses SET chair_sent_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, yetSentRideStatus.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, &chairGetNotificationResponse{
		Data: &chairGetNotificationResponseData{
			RideID: ride.ID,
			User: simpleUser{
				ID:   user.ID,
				Name: fmt.Sprintf("%s %s", user.Firstname, user.Lastname),
			},
			PickupCoordinate: Coordinate{
				Latitude:  ride.PickupLatitude,
				Longitude: ride.PickupLongitude,
			},
			DestinationCoordinate: Coordinate{
				Latitude:  ride.DestinationLatitude,
				Longitude: ride.DestinationLongitude,
			},
			Status: status,
		},
		RetryAfterMs: 1000,
	})
}

type postChairRidesRideIDStatusRequest struct {
	Status string `json:"status"`
}

func chairPostRideStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rideID := r.PathValue("ride_id")

	chair := ctx.Value("chair").(*Chair)

	req := &postChairRidesRideIDStatusRequest{}
	if err := bindJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	ride := &Ride{}
	if err := tx.GetContext(ctx, ride, "SELECT * FROM rides WHERE id = ? FOR UPDATE", rideID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, errors.New("ride not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if ride.ChairID.String != chair.ID {
		writeError(w, http.StatusBadRequest, errors.New("not assigned to this ride"))
		return
	}

	switch req.Status {
	case "ENROUTE":
		newStatusID := ulid.Make().String()
		if _, err := tx.ExecContext(ctx, "INSERT INTO ride_statuses (id, ride_id, status) VALUES (?, ?, ?)", newStatusID, ride.ID, "ENROUTE"); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		setLatestRideStatus(ride.ID, "ENROUTE")
		// slog.Info("INSERT ride_statuses", "ride_id", ride.ID, "status", "ENROUTE", "status_id", newStatusID)
	case "CARRYING":
		status, err := getLatestRideStatus(ctx, tx, ride.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if status != "PICKUP" {
			writeError(w, http.StatusBadRequest, errors.New("chair has not arrived yet"))
			return
		}
		newStatusID := ulid.Make().String()
		if _, err := tx.ExecContext(ctx, "INSERT INTO ride_statuses (id, ride_id, status) VALUES (?, ?, ?)", newStatusID, ride.ID, "CARRYING"); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		setLatestRideStatus(ride.ID, "CARRYING")
		// slog.Info("INSERT ride_statuses", "ride_id", ride.ID, "status", "CARRYING", "status_id", newStatusID)
	default:
		writeError(w, http.StatusBadRequest, errors.New("invalid status"))
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
