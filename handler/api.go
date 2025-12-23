package handler

import (
	"net/url"
	"strconv"

	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/field"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/search"
	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"

	"github.com/gofiber/fiber/v2"
)

func getQueryValues(c *fiber.Ctx) field.Values {
	queryString := c.Request().URI().QueryString()
	if len(queryString) == 0 {
		return make(field.Values)
	}
	parsed, err := url.ParseQuery(string(queryString))
	if err != nil {
		return make(field.Values)
	}
	return field.Values(parsed)
}

func getFormValues(c *fiber.Ctx) field.Values {
	form, _ := c.MultipartForm()
	if form == nil {
		return make(field.Values)
	}
	return field.Values(form.Value)
}

func GetAllValuesHandler(c *fiber.Ctx) error {
	categoryID, err := param.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	specField, fiberErr := param.GetSpecField(c, categoryID)
	if fiberErr != nil {
		return fiberErr
	}

	fv := getQueryValues(c)

	values, err := specField.GetAllValues(fv)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(values)
}

func GetAnyValuesHandler(c *fiber.Ctx) error {
	categoryID, err := param.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	specField, fiberErr := param.GetSpecField(c, categoryID)
	if fiberErr != nil {
		return fiberErr
	}

	fv := getQueryValues(c)

	values, err := specField.GetAnyValues(fv)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(values)
}

func GetAdValuesHandler(c *fiber.Ctx) error {
	categoryID, err := param.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
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

	fv := make(field.Values)
	for k, v := range formValues {
		if k != "ad_ids" {
			fv[k] = v
		}
	}

	values, err := specField.GetAdValues(adIDs, fv)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(values)
}

func GetChainsHandler(c *fiber.Ctx) error {
	categoryID, err := param.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	chains, err := field.GetCategoryChains(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(chains)
}

func GetAdFilterValuesHandler(c *fiber.Ctx) error {
	adID, fiberErr := param.GetAdID(c)
	if fiberErr != nil {
		return fiberErr
	}

	fv, err := ad.LoadFieldValues(adID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fv)
}

func SearchHandler(c *fiber.Ctx) error {
	categoryID, err := param.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	// TODO revisit this
	fv := getFormValues(c)
	if len(fv) == 0 {
		fv = getQueryValues(c)
	}

	adIDs, err := search.Search(categoryID, fv)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	response := map[string]interface{}{
		"ad_ids": adIDs,
		"count":  len(adIDs),
	}

	return c.JSON(response)
}

func GetFirstSpecFieldsHandler(c *fiber.Ctx) error {
	categoryID, err := param.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	fields, err := field.GetFirstSpecFields(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	response := make([]map[string]string, len(fields))
	for i, f := range fields {
		field := f.GetField()
		response[i] = map[string]string{
			"name":         field.Name,
			"display_name": field.DisplayName,
		}
	}

	return c.JSON(response)
}

func GetLastSpecFieldHandler(c *fiber.Ctx) error {
	categoryID, err := param.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	last, err := field.GetLastSpecField(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	field := last.GetField()
	response := map[string]string{
		"name":         field.Name,
		"display_name": field.DisplayName,
	}

	return c.JSON(response)
}

func SwitchCategoryHandler(c *fiber.Ctx) error {
	categoryID, err := param.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	categoryName, err := ad.GetCategoryName(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	categoryImage, err := ad.GetCategoryImageFile(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	cookie.SetCategoryID(c, categoryID)

	userID := local.GetUserID(c)

	return render(c, ui.SearchContainerRefresh(userID, categoryName, categoryImage))
}

func ShowFiltersHandler(c *fiber.Ctx) error {
	categoryID, err := cookie.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	fields, err := field.GetFields(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	fv := getQueryValues(c)

	filters := make([]g.Node, 0, len(fields))
	for _, field := range fields {
		filters = append(filters, field.FilterNode(fv))
	}

	userID := local.GetUserID(c)

	return render(c, ui.SearchWidget(userID, fv.Get("q"), filters))
}
