package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Item struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// List godoc
// @Summary List items
// @Produce json
// @Success 200 {object} OK
// @Router /api/v1/items [get]
func ItemsList(w http.ResponseWriter, r *http.Request) {
	if runtime.DB == nil {
		writeJSON(w, http.StatusOK, OK{OK: true, Data: []Item{}, Time: now()})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	cur, err := runtime.DB.Collection("items").Find(ctx, bson.D{}, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ERR{OK: false, Error: err.Error(), Code: 500, Time: now()})
		return
	}
	defer cur.Close(ctx)
	var out []Item
	for cur.Next(ctx) {
		var it Item
		_ = cur.Decode(&it)
		out = append(out, it)
	}
	writeJSON(w, http.StatusOK, OK{OK: true, Data: out, Time: now()})
}

// Create godoc
// @Summary Create item
// @Accept json
// @Produce json
// @Param body body Item true "Item"
// @Success 201 {object} OK
// @Router /api/v1/items [post]
func ItemCreate(w http.ResponseWriter, r *http.Request) {
	var in Item
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
		writeJSON(w, http.StatusBadRequest, ERR{OK: false, Error: "invalid body", Code: 400, Time: now()})
		return
	}
	in.ID = primitive.NewObjectID()
	in.CreatedAt = time.Now().UTC()

	if runtime.DB != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if _, err := runtime.DB.Collection("items").InsertOne(ctx, in); err != nil {
			writeJSON(w, http.StatusInternalServerError, ERR{OK: false, Error: err.Error(), Code: 500, Time: now()})
			return
		}
	}
	writeJSON(w, http.StatusCreated, OK{OK: true, Data: in, Time: now()})
}

// Get godoc
// @Summary Get item by ID
// @Produce json
// @Param id path string true "ObjectID"
// @Success 200 {object} OK
// @Router /api/v1/items/{id} [get]
func ItemGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ERR{OK: false, Error: "bad id", Code: 400, Time: now()})
		return
	}
	if runtime.DB == nil {
		writeJSON(w, http.StatusNotFound, ERR{OK: false, Error: "not found", Code: 404, Time: now()})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var it Item
	err = runtime.DB.Collection("items").FindOne(ctx, bson.M{"_id": oid}).Decode(&it)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ERR{OK: false, Error: "not found", Code: 404, Time: now()})
		return
	}
	writeJSON(w, http.StatusOK, OK{OK: true, Data: it, Time: now()})
}

// Put godoc
// @Summary Replace item
// @Accept json
// @Produce json
// @Param id path string true "ObjectID"
// @Param body body Item true "Item"
// @Success 200 {object} OK
// @Router /api/v1/items/{id} [put]
func ItemPut(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ERR{OK: false, Error: "bad id", Code: 400, Time: now()})
		return
	}
	var in Item
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
		writeJSON(w, http.StatusBadRequest, ERR{OK: false, Error: "invalid body", Code: 400, Time: now()})
		return
	}
	in.ID = oid
	if runtime.DB != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		_, err := runtime.DB.Collection("items").ReplaceOne(ctx, bson.M{"_id": oid}, in)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ERR{OK: false, Error: err.Error(), Code: 500, Time: now()})
			return
		}
	}
	writeJSON(w, http.StatusOK, OK{OK: true, Data: in, Time: now()})
}

// Patch godoc
// @Summary Patch item (name only)
// @Accept json
// @Produce json
// @Param id path string true "ObjectID"
// @Param body body map[string]string true "Partial"
// @Success 200 {object} OK
// @Router /api/v1/items/{id} [patch]
func ItemPatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ERR{OK: false, Error: "bad id", Code: 400, Time: now()})
		return
	}
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, ERR{OK: false, Error: "invalid body", Code: 400, Time: now()})
		return
	}
	name := body["name"]
	if name == "" {
		writeJSON(w, http.StatusBadRequest, ERR{OK: false, Error: "name required", Code: 400, Time: now()})
		return
	}
	if runtime.DB != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		_, err := runtime.DB.Collection("items").UpdateByID(ctx, oid, bson.M{"$set": bson.M{"name": name}})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, ERR{OK: false, Error: err.Error(), Code: 500, Time: now()})
			return
		}
	}
	writeJSON(w, http.StatusOK, OK{OK: true, Data: map[string]string{"id": id, "name": name}, Time: now()})
}

// Delete godoc
// @Summary Delete item
// @Produce json
// @Param id path string true "ObjectID"
// @Success 200 {object} OK
// @Router /api/v1/items/{id} [delete]
func ItemDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ERR{OK: false, Error: "bad id", Code: 400, Time: now()})
		return
	}
	if runtime.DB != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		_, _ = runtime.DB.Collection("items").DeleteOne(ctx, bson.M{"_id": oid})
	}
	writeJSON(w, http.StatusOK, OK{OK: true, Data: map[string]string{"deleted": id}, Time: now()})
}
