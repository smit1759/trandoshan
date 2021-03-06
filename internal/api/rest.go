package api

import (
	"encoding/base64"
	"github.com/creekorful/trandoshan/api"
	"github.com/creekorful/trandoshan/internal/api/database"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func searchResourcesEndpoint(s service) echo.HandlerFunc {
	return func(c echo.Context) error {
		searchParams, err := newSearchParams(c)
		if err != nil {
			return err
		}

		resources, total, err := s.searchResources(searchParams)
		if err != nil {
			return err
		}

		writePagination(c, searchParams, total)

		return c.JSON(http.StatusOK, resources)
	}
}

func addResourceEndpoint(s service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var res api.ResourceDto
		if err := c.Bind(&res); err != nil {
			return err
		}

		res, err := s.addResource(res)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, res)
	}
}

func scheduleURLEndpoint(s service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var url string
		if err := c.Bind(&url); err != nil {
			return err
		}

		return s.scheduleURL(url)
	}
}

func authenticateEndpoint(s service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Validate provided credentials
		var credentials api.CredentialsDto
		if err := c.Bind(&credentials); err != nil {
			return err
		}

		token, err := s.authenticate(credentials)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, token)
	}
}

func readPagination(c echo.Context) (int, int) {
	paginationPage, err := strconv.Atoi(c.QueryParam(api.PaginationPageQueryParam))
	if err != nil {
		paginationPage = 1
	}
	paginationSize, err := strconv.Atoi(c.QueryParam(api.PaginationSizeQueryParam))
	if err != nil {
		paginationSize = defaultPaginationSize
	}
	// Prevent too much results from being returned
	if paginationSize > maxPaginationSize {
		paginationSize = maxPaginationSize
	}

	return paginationPage, paginationSize
}

func writePagination(c echo.Context, s *database.ResSearchParams, totalCount int64) {
	c.Response().Header().Set(api.PaginationPageHeader, strconv.Itoa(s.PageNumber))
	c.Response().Header().Set(api.PaginationSizeHeader, strconv.Itoa(s.PageSize))
	c.Response().Header().Set(api.PaginationCountHeader, strconv.FormatInt(totalCount, 10))
}

func newSearchParams(c echo.Context) (*database.ResSearchParams, error) {
	params := &database.ResSearchParams{}

	params.Keyword = c.QueryParam("keyword")

	if c.QueryParam("with-body") == "true" {
		params.WithBody = true
	}

	// extract raw query params (unescaped to keep + sign when parsing date)
	rawQueryParams := getRawQueryParam(c.QueryString())

	if val, exist := rawQueryParams["start-date"]; exist {
		d, err := time.Parse(time.RFC3339, val)
		if err == nil {
			params.StartDate = d
		} else {
			return nil, err
		}
	}

	if val, exist := rawQueryParams["end-date"]; exist {
		d, err := time.Parse(time.RFC3339, val)
		if err == nil {
			params.EndDate = d
		} else {
			return nil, err
		}
	}

	// First of all base64decode the URL
	b64URL := c.QueryParam("url")
	b, err := base64.URLEncoding.DecodeString(b64URL)
	if err != nil {
		return nil, err
	}
	params.URL = string(b)

	// Acquire pagination
	page, size := readPagination(c)
	params.PageNumber = page
	params.PageSize = size

	return params, nil
}

func getRawQueryParam(url string) map[string]string {
	if url == "" {
		return map[string]string{}
	}

	val := map[string]string{}
	parts := strings.Split(url, "&")

	for _, part := range parts {
		p := strings.Split(part, "=")
		val[p[0]] = p[1]
	}

	return val
}
