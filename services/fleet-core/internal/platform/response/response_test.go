package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	data := map[string]string{"key": "value"}
	resp := Success(data)

	assert.True(t, resp.Success)
	assert.Equal(t, "Request successful", resp.Message)
	assert.Equal(t, data, resp.Data)
	assert.NotNil(t, resp.Timestamp)
	assert.Nil(t, resp.Error)
}

func TestSuccessWithMessage(t *testing.T) {
	data := map[string]string{"key": "value"}
	resp := SuccessWithMessage(data, "Custom message")

	assert.True(t, resp.Success)
	assert.Equal(t, "Custom message", resp.Message)
	assert.Equal(t, data, resp.Data)
}

func TestSuccessWithMeta(t *testing.T) {
	data := []string{"item1", "item2"}
	resp := SuccessWithMeta(data, 100, 20, 0)

	assert.True(t, resp.Success)
	assert.Equal(t, int64(100), int64(resp.Meta.TotalCount))
	assert.Equal(t, int64(20), int64(resp.Meta.Limit))
	assert.Equal(t, int64(0), int64(resp.Meta.Offset))
}

func TestCreated(t *testing.T) {
	data := map[string]string{"id": "123"}
	resp := Created(data)

	assert.True(t, resp.Success)
	assert.Equal(t, "Resource created successfully", resp.Message)
	assert.Equal(t, data, resp.Data)
}

func TestDeleted(t *testing.T) {
	resp := Deleted()

	assert.True(t, resp.Success)
	assert.Equal(t, "Resource deleted successfully", resp.Message)
	assert.Nil(t, resp.Data)
}

func TestError(t *testing.T) {
	resp := Error("VALIDATION_ERROR", "Invalid email format")

	assert.False(t, resp.Success)
	assert.Equal(t, "Invalid email format", resp.Message)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
}

func TestWrite(t *testing.T) {
	w := httptest.NewRecorder()
	resp := Success(map[string]string{"test": "data"})

	Write(w, http.StatusOK, resp)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	
	var decoded APIResponse
	err := json.NewDecoder(w.Body).Decode(&decoded)
	assert.NoError(t, err)
	assert.True(t, decoded.Success)
}

func TestConvenienceHandlers(t *testing.T) {
	t.Run("Success200", func(t *testing.T) {
		w := httptest.NewRecorder()
		Success200(w, map[string]string{"key": "value"})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Created201", func(t *testing.T) {
		w := httptest.NewRecorder()
		Created201(w, map[string]string{"id": "123"})
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Error400", func(t *testing.T) {
		w := httptest.NewRecorder()
		Error400(w, "BAD_REQUEST", "Invalid input")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error401", func(t *testing.T) {
		w := httptest.NewRecorder()
		Error401(w)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Error403", func(t *testing.T) {
		w := httptest.NewRecorder()
		Error403(w)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Error404", func(t *testing.T) {
		w := httptest.NewRecorder()
		Error404(w, "Resource")
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
