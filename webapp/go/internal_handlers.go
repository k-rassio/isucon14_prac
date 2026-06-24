package main

import (
	"database/sql"
	"errors"
	"net/http"
)

// このAPIをインスタンス内から一定間隔で叩かせることで、椅子とライドをマッチングさせる
func internalGetMatching(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// MEMO: 一旦最も待たせているリクエストに適当な空いている椅子マッチさせる実装とする。おそらくもっといい方法があるはず…

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback() // エラー時は自動でロールバックされる

	ride := &Ride{}
	if err := tx.GetContext(ctx, ride, `SELECT * FROM rides WHERE chair_id IS NULL ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	matched := &Chair{}
	empty := false

	err = tx.GetContext(ctx, matched, `
		SELECT c.*
		FROM chairs c
		INNER JOIN chair_latest_locations cl ON c.id = cl.chair_id
		WHERE c.is_active = TRUE
		AND c.id NOT IN (
			-- 「まだ終わっていないライド」を担当している椅子IDを除外する
			SELECT DISTINCT r.chair_id
			FROM rides r
			INNER JOIN ride_statuses rs ON r.id = rs.ride_id
			WHERE r.chair_id IS NOT NULL
			GROUP BY r.id
			HAVING COUNT(rs.chair_sent_at) < 6
		)
		ORDER BY ABS(cl.latitude - ?) + ABS(cl.longitude - ?) ASC
		LIMIT 1
	`, ride.PickupLatitude, ride.PickupLongitude)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	// 椅子が見つかれば、empty = true とみなしてループ後の処理へ
	empty = true

	if !empty {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := tx.ExecContext(ctx, "UPDATE rides SET chair_id = ? WHERE id = ?", matched.ID, ride.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	chairRideCache.mu.Lock()
	chairRideCache.items[matched.ID] = append(chairRideCache.items[matched.ID], ride.ID)
	chairRideCache.mu.Unlock()

	rideCache.mu.Lock()
	rideCache.items[matched.ID] = *ride
	rideCache.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}
