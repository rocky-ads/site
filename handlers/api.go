package handlers

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/rocky-ads/site/models"
	"github.com/rocky-ads/site/services"

	"github.com/gofiber/fiber/v2"
)

func getCategoryID(c *fiber.Ctx) (int, *fiber.Error) {
	categoryName := c.Params("category")
	// Fiber automatically URL-decodes path parameters, but ensure it's decoded correctly
	decodedName, err := url.PathUnescape(categoryName)
	if err != nil {
		// If PathUnescape fails, try the original (Fiber may have already decoded it)
		decodedName = categoryName
	}
	categoryID, err := services.GetCategoryIDByName(decodedName)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Category not found: %s", decodedName))
	}
	return categoryID, nil
}

func getSpecField(c *fiber.Ctx, categoryID int) (models.SpecField, *fiber.Error) {
	fieldName := c.Params("field")
	specField, err := services.GetSpecField(categoryID, fieldName)
	if err != nil {
		return models.SpecField{}, fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return specField, nil
}

func getAdID(c *fiber.Ctx) (int, *fiber.Error) {
	adIDStr := c.Params("id")
	adID, err := strconv.Atoi(adIDStr)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}
	return adID, nil
}

// getQueryValues converts Fiber's query map (map[string]string) to url.Values (map[string][]string)
func getQueryValues(c *fiber.Ctx) models.FieldValues {
	queries := c.Queries()
	fv := make(models.FieldValues)
	for k, v := range queries {
		fv[k] = []string{v}
	}
	return fv
}

// getFormValues parses form-urlencoded POST data and returns url.Values
func getFormValues(c *fiber.Ctx) (models.FieldValues, error) {
	// Check content type
	contentType := string(c.Request().Header.ContentType())
	if contentType != "application/x-www-form-urlencoded" {
		// If no content type or different type, try to parse as form-urlencoded anyway
		// Some clients might not send the header
	}

	// Parse the body as form-urlencoded
	body := c.Body()
	if len(body) == 0 {
		// If body is empty, return empty FieldValues
		return make(models.FieldValues), nil
	}

	parsed, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}

	return models.FieldValues(parsed), nil
}

func GetAllValuesHandler(c *fiber.Ctx) error {
	categoryID, fiberErr := getCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	specField, fiberErr := getSpecField(c, categoryID)
	if fiberErr != nil {
		return fiberErr
	}

	fv := getQueryValues(c)

	values, err := services.GetAllValues(specField, fv, categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(values)
}

func GetAnyValuesHandler(c *fiber.Ctx) error {
	categoryID, fiberErr := getCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	specField, fiberErr := getSpecField(c, categoryID)
	if fiberErr != nil {
		return fiberErr
	}

	fv := getQueryValues(c)

	values, err := services.GetAnyValues(specField, fv, categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(values)
}

func GetAdValuesHandler(c *fiber.Ctx) error {
	categoryID, fiberErr := getCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	specField, fiberErr := getSpecField(c, categoryID)
	if fiberErr != nil {
		return fiberErr
	}

	formValues, err := getFormValues(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid form data")
	}

	var adIDs []int
	if adIDsVals := formValues["ad_ids"]; len(adIDsVals) > 0 {
		adIDs = make([]int, 0, len(adIDsVals))
		for _, val := range adIDsVals {
			if id, err := strconv.Atoi(val); err == nil {
				adIDs = append(adIDs, id)
			}
		}
	}

	fv := make(models.FieldValues)
	for k, v := range formValues {
		if k != "ad_ids" {
			fv[k] = v
		}
	}

	values, err := services.GetAdValues(adIDs, specField, fv, categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(values)
}

func GetChainsHandler(c *fiber.Ctx) error {
	categoryID, fiberErr := getCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	chains, err := services.GetCategoryChains(categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(chains)
}

// GetAdFilterValuesHandler handles GET /api/ads/:id/filter-values
func GetAdFilterValuesHandler(c *fiber.Ctx) error {
	adID, fiberErr := getAdID(c)
	if fiberErr != nil {
		return fiberErr
	}

	fv, err := services.LoadFilterValues(adID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(fv)
}

func SearchHandler(c *fiber.Ctx) error {
	categoryID, fiberErr := getCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	// Try to get form values first, fall back to query values if body is empty
	fv, err := getFormValues(c)
	if err != nil || len(fv) == 0 {
		// If form parsing failed or no form data, use query parameters
		fv = getQueryValues(c)
	}

	adIDs, err := services.Search(fv, categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	response := map[string]interface{}{
		"ad_ids": adIDs,
		"count":  len(adIDs),
	}

	return c.JSON(response)
}

func GetFirstSpecFieldsHandler(c *fiber.Ctx) error {
	categoryID, fiberErr := getCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	fields, err := services.FirstSpecFields(categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	response := make([]map[string]string, len(fields))
	for i, f := range fields {
		response[i] = map[string]string{
			"name":         f.Name,
			"display_name": f.DisplayName,
		}
	}

	return c.JSON(response)
}

func GetLastSpecFieldHandler(c *fiber.Ctx) error {
	categoryID, fiberErr := getCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	field, err := services.LastSpecField(categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	response := map[string]string{
		"name":         field.Name,
		"display_name": field.DisplayName,
	}

	return c.JSON(response)
}
