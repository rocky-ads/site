package handler

import (
	"net/url"
	"strconv"

	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/field"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/search"
	"github.com/rocky-ads/site/ui"
	uiads "github.com/rocky-ads/site/ui/ads"
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
	categoryID := param.GetCategoryID(c)

	specField, err := param.GetSpecField(c, categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	fv := getQueryValues(c)

	values, err := specField.GetAllValues(fv)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(values)
}

func GetAnyValuesHandler(c *fiber.Ctx) error {
	categoryID := param.GetCategoryID(c)

	specField, err := param.GetSpecField(c, categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	fv := getQueryValues(c)

	values, err := specField.GetAnyValues(fv)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(values)
}

func GetAdValuesHandler(c *fiber.Ctx) error {
	categoryID := param.GetCategoryID(c)

	specField, err := param.GetSpecField(c, categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
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
	categoryID := param.GetCategoryID(c)

	chains, err := field.GetCategoryChains(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(chains)
}

func GetAdFilterValuesHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	fv, err := ad.LoadFieldValues(adID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fv)
}

func SearchHandler(c *fiber.Ctx) error {
	categoryID := param.GetCategoryID(c)

	// TODO revisit this
	fv := getFormValues(c)
	if len(fv) == 0 {
		fv = getQueryValues(c)
	}

	limit, offset := param.GetPageLimitOffset(c)
	adIDs, err := search.Search(categoryID, limit, offset, fv)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	response := map[string]any{
		"ad_ids": adIDs,
		"count":  len(adIDs),
	}

	return c.JSON(response)
}

func GetFirstSpecFieldsHandler(c *fiber.Ctx) error {
	categoryID := param.GetCategoryID(c)

	fields, err := field.GetFirstSpecFields(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
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
	categoryID := param.GetCategoryID(c)

	last, err := field.GetLastSpecField(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	response := map[string]string{
		"name":         last.Name,
		"display_name": last.DisplayName,
	}

	return c.JSON(response)
}

func SwitchCategoryHandler(c *fiber.Ctx) error {
	categoryID := param.GetCategoryID(c)

	if _, err := ad.GetCategoryName(categoryID); err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	cookie.SetCategoryID(c, categoryID)

	redirect := c.Query("return")
	if redirect == "" || redirect[0] != '/' || (len(redirect) > 1 && redirect[1] == '/') {
		redirect = "/"
	}
	c.Set("HX-Redirect", redirect)
	return c.Send(nil)
}

func ShowFiltersHandler(c *fiber.Ctx) error {
	categoryID := cookie.GetCategoryID(c)

	fv := getQueryValues(c)
	q := fv.Get("q")
	searchFV := field.FilterSpecFields(categoryID, fv)

	adFilters := parseAdFilters(c)
	chains, optionsMap, err := field.FilterFieldsOptions(categoryID, adFilters)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	userID := local.GetUserID(c)
	view := cookie.GetView(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)
	limit := config.SearchPageSize
	offset := 0

	results, err := searchAndRenderAds(categoryID, limit, offset, userID, view, searchFV, loc, csrfToken)
	if err != nil {
		return err
	}

	filterNode := uiads.FilterContent(uiads.FilterView{
		CategoryID: categoryID,
		Category:   chains,
		OptionsMap: optionsMap,
		Filters:    adFilters,
	})

	logger.Debug("ShowFiltersHandler results",
		"resultsCount", len(results),
	)

	return render(c, ui.SearchWidget(userID, view, q, results, filterNode))
}

func SearchPageHandler(c *fiber.Ctx) error {
	categoryID := cookie.GetCategoryID(c)

	fv := getQueryValues(c)
	fv = field.FilterSpecFields(categoryID, fv)

	userID := local.GetUserID(c)
	view := cookie.GetView(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)
	limit, offset := param.GetPageLimitOffset(c)

	results, err := searchAndRenderAds(categoryID, limit, offset, userID, view, fv, loc, csrfToken)
	if err != nil {
		return err
	}

	return render(c, g.Group(results))
}
