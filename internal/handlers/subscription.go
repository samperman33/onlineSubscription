package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"onlineSubscription/internal/models"
	"onlineSubscription/internal/storage"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// POST /subscriptions
func CreateSubscriptionHandler(st *storage.PostgresStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateSubRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("invalid request body", "error", err.Error())
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		uid, err := uuid.Parse(req.UserID)
		if err != nil {
			slog.Error("invalid user_id", "user_id", req.UserID, "error", err.Error())
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}

		start, err := parseMonthYear(req.StartDate)
		if err != nil {
			slog.Error("invalid start_date", "start_date", req.StartDate, "error", err.Error())
			http.Error(w, "invalid start_date, expected MM-YYYY", http.StatusBadRequest)
			return
		}

		var end sql.NullTime
		if strings.TrimSpace(req.EndDate) != "" {
			et, err := parseMonthYear(req.EndDate)
			if err != nil {
				slog.Error("invalid end_date", "end_date", req.EndDate, "error", err.Error())
				http.Error(w, "invalid end_date, expected MM-YYYY", http.StatusBadRequest)
				return
			}
			end = sql.NullTime{Time: et, Valid: true}
		}

		id := uuid.New()
		now := time.Now().UTC()

		q := `INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
		      VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
		if _, err := st.DB.Exec(q, id, req.ServiceName, req.Price, uid, start, end, now, now); err != nil {
			slog.Error("insert subscription failed", "error", err.Error(), "user_id", uid.String())
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		slog.Info("subscription created", "subscription_id", id.String(), "user_id", uid.String())

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": id.String()})
	}
}

// GET /subscriptions/{id}
func GetSubscriptionHandler(st *storage.PostgresStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			slog.Error("invalid id", "id", idStr, "error", err.Error())
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var s models.Subscription
		if err := st.DB.Get(&s, "SELECT * FROM subscriptions WHERE id=$1", id); err != nil {
			if err == sql.ErrNoRows {
				slog.Warn("subscription not found", "subscription_id", id.String())
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			slog.Error("get subscription failed", "subscription_id", id.String(), "error", err.Error())
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(s)
	}
}

// PUT /subscriptions/{id}
func UpdateSubscriptionHandler(st *storage.PostgresStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			slog.Error("invalid id", "id", idStr, "error", err.Error())
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var req models.CreateSubRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("invalid request body", "error", err.Error(), "subscription_id", id.String())
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		uid, err := uuid.Parse(req.UserID)
		if err != nil {
			slog.Error("invalid user_id", "user_id", req.UserID, "error", err.Error(), "subscription_id", id.String())
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}

		start, err := parseMonthYear(req.StartDate)
		if err != nil {
			slog.Error("invalid start_date", "start_date", req.StartDate, "error", err.Error(), "subscription_id", id.String())
			http.Error(w, "invalid start_date", http.StatusBadRequest)
			return
		}

		var end sql.NullTime
		if strings.TrimSpace(req.EndDate) != "" {
			et, err := parseMonthYear(req.EndDate)
			if err != nil {
				slog.Error("invalid end_date", "end_date", req.EndDate, "error", err.Error(), "subscription_id", id.String())
				http.Error(w, "invalid end_date", http.StatusBadRequest)
				return
			}
			end = sql.NullTime{Time: et, Valid: true}
		}

		now := time.Now().UTC()
		q := `UPDATE subscriptions
		      SET service_name=$1, price=$2, user_id=$3, start_date=$4, end_date=$5, updated_at=$6
		      WHERE id=$7`
		if _, err := st.DB.Exec(q, req.ServiceName, req.Price, uid, start, end, now, id); err != nil {
			slog.Error("update failed", "subscription_id", id.String(), "error", err.Error())
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// DELETE /subscriptions/{id}
func DeleteSubscriptionHandler(st *storage.PostgresStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			slog.Error("invalid id", "id", idStr, "error", err.Error())
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		if _, err := st.DB.Exec("DELETE FROM subscriptions WHERE id=$1", id); err != nil {
			slog.Error("delete failed", "subscription_id", id.String(), "error", err.Error())
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// GET /subscriptions (list with filters)
func ListSubscriptionsHandler(st *storage.PostgresStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := "SELECT * FROM subscriptions WHERE 1=1"
		args := []interface{}{}
		idx := 1

		if uid := r.URL.Query().Get("user_id"); uid != "" {
			q += fmt.Sprintf(" AND user_id=$%d", idx)
			args = append(args, uid)
			idx++
		}
		if sname := r.URL.Query().Get("service_name"); sname != "" {
			q += fmt.Sprintf(" AND service_name ILIKE $%d", idx)
			args = append(args, "%"+sname+"%")
			idx++
		}

		limit := 100
		if l := r.URL.Query().Get("limit"); l != "" {
			v, err := strconv.Atoi(l)
			if err != nil || v <= 0 {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = v
		}

		q += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(idx)
		args = append(args, limit)
		idx++

		var res []models.Subscription
		if err := st.DB.Select(&res, q, args...); err != nil {
			slog.Error("list failed", "error", err.Error(), "query", q, "args", args)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(res)
	}
}

// GET /subscriptions/aggregate
func AggregateHandler(st *storage.PostgresStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")
		if fromStr == "" || toStr == "" {
			slog.Error("missing required parameters", "from", fromStr, "to", toStr)
			http.Error(w, "from and to parameters required (MM-YYYY)", http.StatusBadRequest)
			return
		}
		from, err := parseMonthYear(fromStr)
		if err != nil {
			slog.Error("invalid from param", "from", fromStr, "error", err.Error())
			http.Error(w, "invalid from param", http.StatusBadRequest)
			return
		}
		to, err := parseMonthYear(toStr)
		if err != nil {
			slog.Error("invalid to param", "to", toStr, "error", err.Error())
			http.Error(w, "invalid to param", http.StatusBadRequest)
			return
		}
		if from.After(to) {
			slog.Error("from date after to date", "from", from, "to", to)
			http.Error(w, "from must be <= to", http.StatusBadRequest)
			return
		}

		q := `SELECT id, service_name, price, user_id, start_date, end_date
		      FROM subscriptions
		      WHERE (end_date IS NULL OR end_date >= $1) AND start_date <= $2`
		args := []interface{}{from, to}
		argIdx := 3

		if uid := r.URL.Query().Get("user_id"); uid != "" {
			q += fmt.Sprintf(" AND user_id = $%d", argIdx)
			args = append(args, uid)
			argIdx++
		}
		if sname := r.URL.Query().Get("service_name"); sname != "" {
			q += fmt.Sprintf(" AND service_name ILIKE $%d", argIdx)
			args = append(args, "%"+sname+"%")
			argIdx++
		}

		var subs []models.Subscription
		if err := st.DB.Select(&subs, q, args...); err != nil {
			slog.Error("aggregate query failed", "error", err.Error(), "query", q, "args", args)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		total := 0
		for _, s := range subs {
			sStart := s.StartDate
			sEnd := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
			if s.EndDate.Valid {
				sEnd = s.EndDate.Time
			}
			months := monthsOverlap(sStart, sEnd, from, to)
			total += months * s.Price
		}

		_ = json.NewEncoder(w).Encode(map[string]int{"total": total})
	}
}

// --- helpers ---

func parseMonthYear(s string) (time.Time, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("bad format")
	}
	month, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, err
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC), nil
}

func monthsOverlap(aStart, aEnd, bStart, bEnd time.Time) int {
	aStart = time.Date(aStart.Year(), aStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	aEnd = time.Date(aEnd.Year(), aEnd.Month(), 1, 0, 0, 0, 0, time.UTC)
	bStart = time.Date(bStart.Year(), bStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	bEnd = time.Date(bEnd.Year(), bEnd.Month(), 1, 0, 0, 0, 0, time.UTC)

	if aEnd.Before(bStart) || bEnd.Before(aStart) {
		return 0
	}
	start := aStart
	if bStart.After(start) {
		start = bStart
	}
	end := aEnd
	if bEnd.Before(end) {
		end = bEnd
	}
	years := end.Year() - start.Year()
	months := int(end.Month()) - int(start.Month())
	return years*12 + months + 1
}
