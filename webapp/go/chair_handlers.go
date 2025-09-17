package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/oklog/ulid/v2"
)

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

	location := &ChairLocation{}
	if err := tx.GetContext(ctx, location, `SELECT * FROM chair_locations WHERE id = ?`, chairLocationID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	// total_distanceを計算してtotal_distanceテーブルに反映
	var prevLocation ChairLocation
	err = tx.GetContext(ctx, &prevLocation, `SELECT * FROM chair_locations WHERE chair_id = ? AND id != ? ORDER BY created_at DESC LIMIT 1`, chair.ID, chairLocationID)
	var distance int
	if err == nil {
		distance = abs(location.Latitude-prevLocation.Latitude) + abs(location.Longitude-prevLocation.Longitude)
	} else if errors.Is(err, sql.ErrNoRows) {
		distance = 0
	} else {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	// 直近の合計距離を取得
	var totalDistance int
	err = tx.GetContext(ctx, &totalDistance, `SELECT distance FROM total_distance WHERE chair_id = ?`, chair.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	newTotalDistance := totalDistance + distance

	// total_distanceテーブルにUPSERT
	_, err = tx.ExecContext(ctx, `
    INSERT INTO total_distance (chair_id, distance, updated_at)
    VALUES (?, ?, ?)
    ON DUPLICATE KEY UPDATE distance = VALUES(distance), updated_at = VALUES(updated_at)
`, chair.ID, newTotalDistance, location.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	ride := &Ride{}
	if err := tx.GetContext(ctx, ride, `SELECT * FROM rides WHERE chair_id = ? ORDER BY updated_at DESC LIMIT 1`, chair.ID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		status, err := getLatestRideStatus(ctx, tx, ride.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if status != "COMPLETED" && status != "CANCELED" {
			if req.Latitude == ride.PickupLatitude && req.Longitude == ride.PickupLongitude && status == "ENROUTE" {
				newStatusID := ulid.Make().String()
				if _, err := tx.ExecContext(ctx, "INSERT INTO ride_statuses (id, ride_id, status) VALUES (?, ?, ?)", newStatusID, ride.ID, "PICKUP"); err != nil {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
				slog.Info("INSERT ride_statuses", "ride_id", ride.ID, "status", "PICKUP", "status_id", newStatusID)
			}

			if req.Latitude == ride.DestinationLatitude && req.Longitude == ride.DestinationLongitude && status == "CARRYING" {
				newStatusID := ulid.Make().String()
				if _, err := tx.ExecContext(ctx, "INSERT INTO ride_statuses (id, ride_id, status) VALUES (?, ?, ?)", newStatusID, ride.ID, "ARRIVED"); err != nil {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
				slog.Info("INSERT ride_statuses", "ride_id", ride.ID, "status", "ARRIVED", "status_id", newStatusID)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, &chairPostCoordinateResponse{
		RecordedAt: location.CreatedAt.UnixMilli(),
	})
}

// グローバル: chairごとの通知チャネル
var chairNotificationChans = make(map[string]chan struct{})

// 通知関数（rides/ride_statuses更新時に呼び出す）
func notifyChair(chairID string) {
	if ch, ok := chairNotificationChans[chairID]; ok {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

type simpleUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type chairGetNotificationResponse struct {
	Data *chairGetNotificationResponseData `json:"data"`
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

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// チャネル登録
	ch := make(chan struct{}, 1)
	chairNotificationChans[chair.ID] = ch
	defer delete(chairNotificationChans, chair.ID)

	for {
		tx, err := db.Beginx()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		ride := &Ride{}
		yetSentRideStatus := RideStatus{}
		status := ""

		if err := tx.GetContext(ctx, ride, `SELECT * FROM rides WHERE chair_id = ? ORDER BY updated_at DESC LIMIT 1`, chair.ID); err != nil {
			tx.Rollback()
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Fprintf(w, "data: {}\n\n")
				flusher.Flush()
				return
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		if err := tx.GetContext(ctx, &yetSentRideStatus, `SELECT * FROM ride_statuses WHERE ride_id = ? AND chair_sent_at IS NULL ORDER BY created_at ASC LIMIT 1`, ride.ID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				status, err = getLatestRideStatus(ctx, tx, ride.ID)
				if err != nil {
					tx.Rollback()
					fmt.Fprintf(w, "event: error\ndata: %q\n\n", err.Error())
					flusher.Flush()
					return
				}
			} else {
				tx.Rollback()
				fmt.Fprintf(w, "event: error\ndata: %q\n\n", err.Error())
				flusher.Flush()
				return
			}
		} else {
			status = yetSentRideStatus.Status
		}

		user := &User{}
		err = tx.GetContext(ctx, user, "SELECT * FROM users WHERE id = ? FOR SHARE", ride.UserID)
		if err != nil {
			tx.Rollback()
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		response := &chairGetNotificationResponse{
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
		}

		b, err := json.Marshal(response.Data)
		if err != nil {
			tx.Rollback()
			fmt.Fprintf(w, "event: error\ndata: %q\n\n", err.Error())
			flusher.Flush()
			return
		}
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()

		if yetSentRideStatus.ID != "" {
			_, err := tx.ExecContext(ctx, `UPDATE ride_statuses SET chair_sent_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, yetSentRideStatus.ID)
			if err != nil {
				tx.Rollback()
				fmt.Fprintf(w, "event: error\ndata: %q\n\n", err.Error())
				flusher.Flush()
				return
			}
		}

		if err := tx.Commit(); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		// チャネル通知を待つ（rides/ride_statuses更新時にnotifyChair(chair.ID)を呼ぶこと）
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			// 通知が来たら次の処理へ
		}
	}
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
		slog.Info("INSERT ride_statuses", "ride_id", ride.ID, "status", "ENROUTE", "status_id", newStatusID)
		// ride_statuses更新後にappNotificationChansへ通知
		notifyApp(ride.UserID)
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
		slog.Info("INSERT ride_statuses", "ride_id", ride.ID, "status", "CARRYING", "status_id", newStatusID)
		// ride_statuses更新後にappNotificationChansへ通知
		notifyApp(ride.UserID)
	default:
		writeError(w, http.StatusBadRequest, errors.New("invalid status"))
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
