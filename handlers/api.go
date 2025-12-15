package handlers

import (
	"net/url"
	"strconv"

	"github.com/rocky-ads/site/models"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/services"

	"github.com/gofiber/fiber/v2"
)

func getQueryValues(c *fiber.Ctx) models.FieldValues {
	queryString := c.Request().URI().QueryString()
	if len(queryString) == 0 {
		return make(models.FieldValues)
	}
	parsed, err := url.ParseQuery(string(queryString))
	if err != nil {
		return make(models.FieldValues)
	}
	return models.FieldValues(parsed)
}

func getFormValues(c *fiber.Ctx) models.FieldValues {
	form, _ := c.MultipartForm()
	if form == nil {
		return make(models.FieldValues)
	}
	return models.FieldValues(form.Value)
}

func GetAllValuesHandler(c *fiber.Ctx) error {
	categoryID, fiberErr := param.GetCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	specField, fiberErr := param.GetSpecField(c, categoryID)
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
	categoryID, fiberErr := param.GetCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	specField, fiberErr := param.GetSpecField(c, categoryID)
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
	categoryID, fiberErr := param.GetCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	specField, fiberErr := param.GetSpecField(c, categoryID)
	if fiberErr != nil {
		return fiberErr
	}

	formValues := getFormValues(c)

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
	categoryID, fiberErr := param.GetCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	chains, err := services.GetCategoryChains(categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(chains)
}

func GetAdFilterValuesHandler(c *fiber.Ctx) error {
	adID, fiberErr := param.GetAdID(c)
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
	categoryID, fiberErr := param.GetCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	fv := getFormValues(c)
	if len(fv) == 0 {
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
	categoryID, fiberErr := param.GetCategoryID(c)
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
	categoryID, fiberErr := param.GetCategoryID(c)
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
