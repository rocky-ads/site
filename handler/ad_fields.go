package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/currency"
	"github.com/rocky-ads/site/field"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/ui"
	uiads "github.com/rocky-ads/site/ui/ads"
	"github.com/rocky-ads/site/user"
)

func NewAdHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	if userID == 0 {
		return redirectToLogin(c)
	}

	categoryID := cookie.GetCategoryID(c)
	categoryName, err := ad.GetCategoryName(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	chains, optionsMap, err := field.CategoryFieldsOptions(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	fieldsNode := uiads.CategoryFieldsPartial(uiads.FieldsView{
		Category:        chains,
		OptionsMap:      optionsMap,
		DefaultCurrency: defaultCurrencyForUser(c),
	})

	return renderPage(c, "New Ad", ui.NewAd(categoryName, fieldsNode))
}

func NewAdFieldsHandler(c *fiber.Ctx) error {
	categoryID := cookie.GetCategoryID(c)
	if categoryID == 0 {
		if id, err := strconv.Atoi(c.Query("category_id")); err == nil {
			categoryID = id
		}
	}
	if categoryID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "category required")
	}

	chains, optionsMap, err := field.CategoryFieldsOptions(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return render(c, uiads.CategoryFieldsPartial(uiads.FieldsView{
		Category:        chains,
		OptionsMap:      optionsMap,
		DefaultCurrency: defaultCurrencyForUser(c),
	}))
}

func NewAdPriceFieldHandler(c *fiber.Ctx) error {
	categoryID, err := strconv.Atoi(c.Query("category_id"))
	if err != nil || categoryID <= 0 {
		categoryID = cookie.GetCategoryID(c)
	}
	if categoryID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "category required")
	}

	chainID, err := strconv.Atoi(c.Query("chain_id"))
	if err != nil || chainID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid chain")
	}

	chains, err := field.GetCategoryChainsMetadata(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	chain, ok := field.FindChain(chains.Chains, chainID)
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "Chain not found")
	}

	priceField, ok := findChainFieldByName(chain, "price")
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "Price field not found")
	}

	def := defaultCurrencyForUser(c)
	isFree := c.Query("price_free") == "1"
	amount := strings.TrimSpace(c.Query("price"))
	if isFree {
		amount = ""
	} else if amount == "0" {
		amount = ""
	}

	priceCurrency := c.Query("price_currency")
	if priceCurrency == "" {
		priceCurrency = def
	}

	return render(c, uiads.PriceFieldPartial(categoryID, chain, priceField, uiads.PriceFieldView{
		DefaultCurrency: def,
		IsFree:          isFree,
		Amount:          amount,
		Currency:        priceCurrency,
	}))
}

func NewAdNextFieldHandler(c *fiber.Ctx) error {
	categoryID, err := strconv.Atoi(c.Query("category_id"))
	if err != nil || categoryID <= 0 {
		categoryID = cookie.GetCategoryID(c)
	}
	if categoryID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid category")
	}

	chainID, err := strconv.Atoi(c.Query("chain_id"))
	if err != nil || chainID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid chain")
	}

	afterField := c.Query("after")
	fieldName := c.Query("field")

	chains, err := field.GetCategoryChainsMetadata(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	chain, ok := field.FindChain(chains.Chains, chainID)
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "Chain not found")
	}

	parents := parentValuesFromQuery(c, chain.Fields)

	if fieldName == "" || !field.ParentsReady(chain, fieldName, parents) {
		return render(c, uiads.NextFieldPartial(categoryID, chain, afterField, "", nil, false))
	}

	opts, err := field.ListSpecOptions(chain, fieldName, categoryID, parents)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	noMatch := len(opts) == 0
	return render(c, uiads.NextFieldPartial(categoryID, chain, afterField, fieldName, opts, noMatch))
}

func FilterNextFieldHandler(c *fiber.Ctx) error {
	categoryID := cookie.GetCategoryID(c)
	if categoryID == 0 {
		if id, err := strconv.Atoi(c.Query("category_id")); err == nil {
			categoryID = id
		}
	}
	if categoryID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "category required")
	}

	chainID, err := strconv.Atoi(c.Query("chain_id"))
	if err != nil || chainID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid chain")
	}

	afterField := c.Query("after")
	fieldName := c.Query("field")

	chains, err := field.GetCategoryChainsMetadata(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	chain, ok := field.FindChain(chains.Chains, chainID)
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "Chain not found")
	}

	parents := parentValuesFromQuery(c, chain.Fields)

	if fieldName == "" || !field.ParentsReady(chain, fieldName, parents) {
		return render(c, uiads.FilterNextFieldPartial(categoryID, chain, afterField, "", nil, false))
	}

	opts, err := field.ListAdFilterOptions(categoryID, chain, fieldName, parents)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	noMatch := len(opts) == 0
	return render(c, uiads.FilterNextFieldPartial(categoryID, chain, afterField, fieldName, opts, noMatch))
}

func defaultCurrencyForUser(c *fiber.Ctx) string {
	userID := local.GetUserID(c)
	if userID == 0 {
		return currency.Default
	}
	u, err := user.GetByID(userID)
	if err != nil {
		return currency.Default
	}
	return currency.DefaultFromPhone(u.PhoneE64)
}

func findChainFieldByName(chain field.ChainGroup, fieldName string) (field.ChainField, bool) {
	for _, f := range chain.Fields {
		if f.FieldName == fieldName {
			return f, true
		}
	}
	return field.ChainField{}, false
}
