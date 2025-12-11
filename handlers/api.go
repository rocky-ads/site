package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rocky-ads/site/cache"
	"github.com/rocky-ads/site/models"
	"github.com/rocky-ads/site/services"

	"github.com/gorilla/mux"
)

func getCategoryID(w http.ResponseWriter, r *http.Request) (int, bool) {
	vars := mux.Vars(r)
	categoryName := vars["category"]
	categoryID, err := cache.GetCategoryIDByName(categoryName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Category not found: %s", categoryName), http.StatusNotFound)
		return 0, false
	}
	return categoryID, true
}

func getSpecField(w http.ResponseWriter, r *http.Request, categoryID int) (models.SpecField, bool) {
	vars := mux.Vars(r)
	fieldName := vars["field"]
	specField, err := cache.GetSpecField(categoryID, fieldName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return models.SpecField{}, false
	}
	return specField, true
}

func getAdID(w http.ResponseWriter, r *http.Request) (int, bool) {
	vars := mux.Vars(r)
	adIDStr := vars["id"]
	adID, err := strconv.Atoi(adIDStr)
	if err != nil {
		http.Error(w, "Invalid ad ID", http.StatusBadRequest)
		return 0, false
	}
	return adID, true
}

func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func GetAllValuesHandler(w http.ResponseWriter, r *http.Request) {
	categoryID, ok := getCategoryID(w, r)
	if !ok {
		return
	}

	specField, ok := getSpecField(w, r, categoryID)
	if !ok {
		return
	}

	fv := models.FieldValues(r.URL.Query())

	values, err := services.GetAllValues(specField, fv, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, values)
}

func GetAnyValuesHandler(w http.ResponseWriter, r *http.Request) {
	categoryID, ok := getCategoryID(w, r)
	if !ok {
		return
	}

	specField, ok := getSpecField(w, r, categoryID)
	if !ok {
		return
	}

	fv := models.FieldValues(r.URL.Query())

	values, err := services.GetAnyValues(specField, fv, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, values)
}

func GetAdValuesHandler(w http.ResponseWriter, r *http.Request) {
	categoryID, ok := getCategoryID(w, r)
	if !ok {
		return
	}

	specField, ok := getSpecField(w, r, categoryID)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	var adIDs []int
	if adIDsVals := r.PostForm["ad_ids"]; len(adIDsVals) > 0 {
		adIDs = make([]int, 0, len(adIDsVals))
		for _, val := range adIDsVals {
			if id, err := strconv.Atoi(val); err == nil {
				adIDs = append(adIDs, id)
			}
		}
	}

	fv := make(models.FieldValues)
	for k, v := range r.PostForm {
		if k != "ad_ids" {
			fv[k] = v
		}
	}

	values, err := services.GetAdValues(adIDs, specField, fv, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, values)
}

func GetChainsHandler(w http.ResponseWriter, r *http.Request) {
	categoryID, ok := getCategoryID(w, r)
	if !ok {
		return
	}

	chains, err := services.GetCategoryChains(categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, chains)
}

// GetAdFilterValuesHandler handles GET /api/ads/:id/filter-values
func GetAdFilterValuesHandler(w http.ResponseWriter, r *http.Request) {
	adID, ok := getAdID(w, r)
	if !ok {
		return
	}

	fv, err := services.LoadFilterValues(adID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, fv)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	categoryID, ok := getCategoryID(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	fv := models.FieldValues(r.PostForm)

	adIDs, err := services.Search(fv, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"ad_ids": adIDs,
		"count":  len(adIDs),
	}

	writeJSONResponse(w, response)
}

func GetFirstSpecFieldsHandler(w http.ResponseWriter, r *http.Request) {
	categoryID, ok := getCategoryID(w, r)
	if !ok {
		return
	}

	fields, err := services.FirstSpecFields(categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]map[string]string, len(fields))
	for i, f := range fields {
		response[i] = map[string]string{
			"name":         f.Name,
			"display_name": f.DisplayName,
		}
	}

	writeJSONResponse(w, response)
}

func GetLastSpecFieldHandler(w http.ResponseWriter, r *http.Request) {
	categoryID, ok := getCategoryID(w, r)
	if !ok {
		return
	}

	field, err := services.LastSpecField(categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"name":         field.Name,
		"display_name": field.DisplayName,
	}

	writeJSONResponse(w, response)
}
